package com.dlddu.pocketaide.scratchpad

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.affirmations.AffirmationPriority
import com.dlddu.pocketaide.affirmations.updateAffirmation
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.Clock

data class ScratchpadUiState(
    val items: List<ScratchpadItem> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val errorMessage: String? = null,
    val priorityEdit: Affirmation? = null,
) {
    fun sections(clock: Clock = Clock.systemDefaultZone()): List<ScratchpadSection> =
        ScratchpadSections.group(items, clock)
}

class ScratchpadViewModel(private val api: ApiClient) : ViewModel() {
    private val _state = MutableStateFlow(ScratchpadUiState())
    val state: StateFlow<ScratchpadUiState> = _state.asStateFlow()

    private val _unclassifiedCount = MutableStateFlow(0)
    val unclassifiedCount: StateFlow<Int> = _unclassifiedCount.asStateFlow()

    fun load() {
        viewModelScope.launch { reload() }
    }

    fun refresh() {
        viewModelScope.launch {
            _state.update { it.copy(isRefreshing = true) }
            try {
                reload()
            } finally {
                _state.update { it.copy(isRefreshing = false) }
            }
        }
    }

    fun add(text: String) {
        viewModelScope.launch { performAdd(text) }
    }

    fun delete(id: Long) {
        viewModelScope.launch { performDelete(id) }
    }

    fun move(item: ScratchpadItem, target: ScratchpadMoveTarget) {
        viewModelScope.launch {
            val affirmation = performMove(item, target)
            if (affirmation != null) _state.update { it.copy(priorityEdit = affirmation) }
        }
    }

    fun dismissPriorityEdit() {
        _state.update { it.copy(priorityEdit = null) }
    }

    fun setPriority(affirmation: Affirmation, text: String, priority: AffirmationPriority) {
        viewModelScope.launch { performSetPriority(affirmation, text, priority) }
    }

    internal suspend fun reload() {
        _state.update { it.copy(isLoading = true, errorMessage = null) }
        try {
            setItems(api.listScratchpad())
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun performAdd(text: String) {
        try {
            val created = api.createScratchpadItem(text)
            setItems(listOf(created) + _state.value.items)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    internal suspend fun performDelete(id: Long) {
        try {
            api.deleteScratchpadItem(id)
            setItems(_state.value.items.filterNot { it.id == id })
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    internal suspend fun performMove(item: ScratchpadItem, target: ScratchpadMoveTarget): Affirmation? =
        try {
            val result = api.moveScratchpadItem(item.id, target)
            setItems(_state.value.items.filterNot { it.id == item.id })
            result.affirmation
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
            null
        }

    internal suspend fun performSetPriority(affirmation: Affirmation, text: String, priority: AffirmationPriority) {
        try {
            api.updateAffirmation(affirmation.id, text, priority)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        }
    }

    private fun setItems(items: List<ScratchpadItem>) {
        _state.update { it.copy(items = items) }
        _unclassifiedCount.value = items.size
    }

    private fun fail(e: Exception) {
        _state.update { it.copy(errorMessage = e.message ?: e.toString()) }
    }
}
