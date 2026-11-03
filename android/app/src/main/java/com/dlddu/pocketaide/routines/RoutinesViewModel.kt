package com.dlddu.pocketaide.routines

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
import java.time.LocalDate

data class RoutinesUiState(
    val today: List<RoutineDay> = emptyList(),
    val routines: List<Routine> = emptyList(),
    val dayKey: String = "",
    val isLoading: Boolean = false,
    val isRefreshing: Boolean = false,
    val errorMessage: String? = null,
    val canRetry: Boolean = false,
) {
    val restingToday: List<Routine>
        get() {
            val scheduled = today.map { it.id }.toSet()
            return routines.filter { it.id !in scheduled }
        }

    fun routine(id: Long): Routine? = routines.firstOrNull { it.id == id }
}

class RoutinesViewModel(
    private val api: ApiClient,
    private val clock: Clock = Clock.systemDefaultZone(),
) : ViewModel() {
    private val _state = MutableStateFlow(RoutinesUiState(dayKey = RoutineDayFormat.string(today())))
    val state: StateFlow<RoutinesUiState> = _state.asStateFlow()

    private var retryAction: (suspend () -> Unit)? = null

    fun today(): LocalDate = LocalDate.now(clock)

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

    fun retry() {
        val action = retryAction ?: return
        viewModelScope.launch { action() }
    }

    fun create(draft: RoutineDraft) {
        viewModelScope.launch { performCreate(draft) }
    }

    fun delete(id: Long) {
        viewModelScope.launch { performDelete(id) }
    }

    fun addStep(routineId: Long, title: String) {
        viewModelScope.launch { performAddStep(routineId, title) }
    }

    fun deleteStep(stepId: Long, routineId: Long) {
        viewModelScope.launch { performDeleteStep(stepId, routineId) }
    }

    fun toggle(step: RoutineDayStep, routine: RoutineDay) {
        viewModelScope.launch { performToggle(step, routine) }
    }

    internal suspend fun reload() {
        _state.update { it.copy(isLoading = true) }
        clearFailure()
        val key = RoutineDayFormat.string(today())
        try {
            val all = api.listRoutines()
            val scheduled = api.listRoutines(day = key)
            _state.update { it.copy(dayKey = key, routines = all, today = scheduled) }
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(RoutineFailureCopy.Action.LOAD, e) { reload() }
        } finally {
            _state.update { it.copy(isLoading = false) }
        }
    }

    internal suspend fun performCreate(draft: RoutineDraft) {
        try {
            api.createRoutine(draft)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(RoutineFailureCopy.Action.CREATE, e) { performCreate(draft) }
            return
        }
        reload()
    }

    internal suspend fun performDelete(id: Long) {
        try {
            api.deleteRoutine(id)
            _state.update { state ->
                state.copy(
                    routines = state.routines.filterNot { it.id == id },
                    today = state.today.filterNot { it.id == id },
                )
            }
            clearFailure()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(RoutineFailureCopy.Action.DELETE, e) { performDelete(id) }
        }
    }

    internal suspend fun performAddStep(routineId: Long, title: String) {
        try {
            api.addRoutineStep(routineId, title)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(RoutineFailureCopy.Action.ADD_STEP, e) { performAddStep(routineId, title) }
            return
        }
        reload()
    }

    internal suspend fun performDeleteStep(stepId: Long, routineId: Long) {
        try {
            api.deleteRoutineStep(routineId, stepId)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(RoutineFailureCopy.Action.DELETE_STEP, e) { performDeleteStep(stepId, routineId) }
            return
        }
        reload()
    }

    internal suspend fun performToggle(step: RoutineDayStep, routine: RoutineDay) {
        setStep(step, routine, checked = !step.checked)
    }

    private suspend fun setStep(step: RoutineDayStep, routine: RoutineDay, checked: Boolean) {
        try {
            val updated = api.setRoutineStep(routine.id, step.id, routine.day, checked)
            _state.update { state ->
                state.copy(today = state.today.map { if (it.id == updated.id) updated else it })
            }
            clearFailure()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(RoutineFailureCopy.Action.CHECK, e) { setStep(step, routine, checked) }
        }
    }

    suspend fun history(routineId: Long): List<RoutineHistoryDay> =
        try {
            api.routineHistory(routineId, _state.value.dayKey)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            fail(RoutineFailureCopy.Action.HISTORY, e, retry = null)
            emptyList()
        }

    private fun clearFailure() {
        retryAction = null
        _state.update { it.copy(errorMessage = null, canRetry = false) }
    }

    private fun fail(action: RoutineFailureCopy.Action, error: Exception, retry: (suspend () -> Unit)?) {
        retryAction = retry
        _state.update { it.copy(errorMessage = RoutineFailureCopy.message(action, error), canRetry = retry != null) }
    }
}
