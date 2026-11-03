package com.dlddu.pocketaide.affirmations

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class AffirmationsUiState(
    val items: List<Affirmation> = emptyList(),
    val heroId: Long? = null,
    val sortOrder: AffirmationSortOrder = AffirmationSortOrder.PRIORITY,
    val isLoading: Boolean = false,
    val errorMessage: String? = null,
) {
    val heroItem: Affirmation?
        get() {
            val id = heroId ?: return items.firstOrNull()
            return items.firstOrNull { it.id == id } ?: items.firstOrNull()
        }

    val sortedItems: List<Affirmation> get() = sortOrder.sorted(items)
}

class AffirmationsViewModel(
    private val api: ApiClient,
    private val random: RandomSource = RandomSource.system,
    private val selector: RotationSelector = RotationSelector(),
    private val onChanged: () -> Unit = {},
) : ViewModel() {
    private val _state = MutableStateFlow(AffirmationsUiState())
    val state: StateFlow<AffirmationsUiState> = _state.asStateFlow()

    fun load() {
        viewModelScope.launch { reload() }
    }

    fun add(text: String, priority: AffirmationPriority) {
        viewModelScope.launch { performAdd(text, priority) }
    }

    fun update(id: Long, text: String, priority: AffirmationPriority) {
        viewModelScope.launch { performUpdate(id, text, priority) }
    }

    fun delete(id: Long) {
        viewModelScope.launch { performDelete(id) }
    }

    fun setSortOrder(order: AffirmationSortOrder) {
        _state.update { it.copy(sortOrder = order) }
    }

    fun rotateHero() {
        val picked = selector.pick(_state.value.items, random)
        _state.update { it.copy(heroId = picked?.id) }
    }

    internal suspend fun reload() {
        _state.update { it.copy(isLoading = true, errorMessage = null) }
        try {
            val loaded = api.listAffirmations()
            _state.update { it.copy(items = loaded) }
            rotateHero()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun performAdd(text: String, priority: AffirmationPriority) {
        try {
            val created = api.createAffirmation(text, priority)
            _state.update { it.copy(items = listOf(created) + it.items, heroId = created.id) }
            onChanged()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    internal suspend fun performUpdate(id: Long, text: String, priority: AffirmationPriority) {
        try {
            val updated = api.updateAffirmation(id, text, priority)
            _state.update { state ->
                state.copy(
                    items = state.items.map { if (it.id == id) updated else it },
                    heroId = updated.id,
                )
            }
            onChanged()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    internal suspend fun performDelete(id: Long) {
        try {
            api.deleteAffirmation(id)
            _state.update { state -> state.copy(items = state.items.filterNot { it.id == id }) }
            onChanged()
            if (_state.value.heroId == id) rotateHero()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    private fun fail(e: Exception) {
        _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
    }
}
