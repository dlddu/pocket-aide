package com.dlddu.pocketaide.todos

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.dlddu.pocketaide.api.ApiClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.Clock
import java.time.temporal.WeekFields
import java.util.Locale

data class TodoListUiState(
    val items: List<TodoItem> = emptyList(),
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val query: String = "",
    val dueFilter: TodoDueFilter = TodoDueFilter.ALL,
    val errorMessage: String? = null,
) {
    val openCount: Int get() = items.count { !it.isDone }
    val doneCount: Int get() = items.size - openCount
    val deadlineCount: Int get() = TodoDeadline.openCount(items)

    fun visibleItems(
        clock: Clock = Clock.systemDefaultZone(),
        weekFields: WeekFields = WeekFields.of(Locale.getDefault()),
    ): List<TodoItem> =
        TodoSearch.filter(items, query).filter { dueFilter.includes(it, clock, weekFields) }
}

class TodoListViewModel(
    val area: TodoArea,
    private val api: ApiClient,
) : ViewModel() {
    private val _state = MutableStateFlow(TodoListUiState())
    val state: StateFlow<TodoListUiState> = _state.asStateFlow()

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

    fun setQuery(query: String) {
        _state.update { it.copy(query = query) }
    }

    fun setDueFilter(filter: TodoDueFilter) {
        _state.update { it.copy(dueFilter = filter) }
    }

    fun add(draft: TodoDraft) {
        viewModelScope.launch { performAdd(draft) }
    }

    fun update(id: Long, draft: TodoDraft) {
        viewModelScope.launch { performUpdate(id, draft) }
    }

    fun toggleDone(item: TodoItem) {
        viewModelScope.launch { performToggleDone(item) }
    }

    fun delete(id: Long) {
        viewModelScope.launch { performDelete(id) }
    }

    internal suspend fun reload() {
        _state.update { it.copy(isLoading = true, errorMessage = null) }
        try {
            val items = api.listTodos(area)
            _state.update { it.copy(items = items) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun performAdd(draft: TodoDraft) {
        try {
            api.createTodo(area, draft)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
            return
        }
        reload()
    }

    internal suspend fun performUpdate(id: Long, draft: TodoDraft) {
        try {
            api.updateTodo(area, id, draft)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(e)
            return
        }
        reload()
    }

    internal suspend fun performToggleDone(item: TodoItem) {
        val draft = TodoDraft.from(item)
        performUpdate(item.id, draft.copy(done = !draft.done))
    }

    internal suspend fun performDelete(id: Long) {
        try {
            api.deleteTodo(area, id)
            _state.update { state -> state.copy(items = state.items.filterNot { it.id == id }) }
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
