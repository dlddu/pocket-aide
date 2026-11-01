package com.dlddu.pocketaide.routines

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.time.DayOfWeek
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.time.format.DateTimeParseException
import kotlin.math.roundToInt

@Serializable
enum class RoutineCadence(val wire: String, val displayName: String) {
    @SerialName("daily")
    DAILY("daily", "매일"),

    @SerialName("weekdays")
    WEEKDAYS("weekdays", "특정 요일"),

    @SerialName("weekly")
    WEEKLY("weekly", "매주"),

    @SerialName("monthly")
    MONTHLY("monthly", "매월"),
}

object RoutineWeekdays {
    val symbols = listOf("일", "월", "화", "수", "목", "금", "토")
    val displayOrder = listOf(1, 2, 3, 4, 5, 6, 0)

    fun bit(index: Int): Int = 1 shl index

    fun contains(mask: Int, index: Int): Boolean = (mask and bit(index)) != 0

    fun index(dayOfWeek: DayOfWeek): Int = dayOfWeek.value % 7

    fun summary(cadence: RoutineCadence, weekdays: Int, monthDay: Int): String = when (cadence) {
        RoutineCadence.DAILY -> "매일"
        RoutineCadence.WEEKDAYS -> displayOrder.filter { contains(weekdays, it) }.joinToString("·") { symbols[it] }
        RoutineCadence.WEEKLY -> "매주 ${symbols[displayOrder.firstOrNull { contains(weekdays, it) } ?: 0]}요일"
        RoutineCadence.MONTHLY -> "매월 ${monthDay}일"
    }
}

object RoutineDayFormat {
    private val formatter: DateTimeFormatter = DateTimeFormatter.ISO_LOCAL_DATE

    fun string(date: LocalDate): String = date.format(formatter)

    fun date(day: String): LocalDate? =
        try {
            LocalDate.parse(day, formatter)
        } catch (e: DateTimeParseException) {
            null
        }

    fun title(day: String): String {
        val parts = day.split("-").mapNotNull { it.toIntOrNull() }
        if (parts.size != 3) return day
        return "${parts[1]}월 ${parts[2]}일"
    }
}

object RoutineSchedule {
    private const val LOOKAHEAD_DAYS = 366

    fun daysUntilNext(routine: Routine, after: String): Int? {
        val base = RoutineDayFormat.date(after) ?: return null
        for (offset in 1..LOOKAHEAD_DAYS) {
            val date = base.plusDays(offset.toLong())
            if (RoutineDayFormat.string(date) >= routine.startDay && matches(routine, date)) return offset
        }
        return null
    }

    internal fun matches(routine: Routine, date: LocalDate): Boolean = when (routine.cadence) {
        RoutineCadence.DAILY -> true
        RoutineCadence.WEEKDAYS, RoutineCadence.WEEKLY ->
            RoutineWeekdays.contains(routine.weekdays, RoutineWeekdays.index(date.dayOfWeek))
        RoutineCadence.MONTHLY -> date.dayOfMonth == minOf(routine.monthDay, date.lengthOfMonth())
    }
}

@Serializable
data class RoutineStep(val id: Long, val title: String, val position: Int)

@Serializable
data class Routine(
    val id: Long,
    val name: String,
    val cadence: RoutineCadence,
    val weekdays: Int = 0,
    @SerialName("month_day") val monthDay: Int = 0,
    @SerialName("start_day") val startDay: String,
    val steps: List<RoutineStep> = emptyList(),
    @SerialName("created_at") val createdAt: Long = 0,
) {
    val scheduleSummary: String get() = RoutineWeekdays.summary(cadence, weekdays, monthDay)
}

@Serializable
data class RoutineDayStep(val id: Long, val title: String, val position: Int, val checked: Boolean)

@Serializable
data class RoutineDay(
    val id: Long,
    val name: String,
    val cadence: RoutineCadence,
    val weekdays: Int = 0,
    @SerialName("month_day") val monthDay: Int = 0,
    val day: String,
    val steps: List<RoutineDayStep> = emptyList(),
    val done: Int,
    val total: Int,
    val completed: Boolean,
) {
    val scheduleSummary: String get() = RoutineWeekdays.summary(cadence, weekdays, monthDay)

    val progress: Float get() = if (total > 0) done.toFloat() / total else 0f

    val progressPercent: Int get() = percent(done, total)
}

@Serializable
data class RoutineHistoryDay(
    val day: String,
    val scheduled: Boolean,
    val done: Int,
    val total: Int,
    val completed: Boolean,
) {
    val statusLabel: String
        get() = when {
            !scheduled -> "쉬는 날"
            completed -> "완료"
            else -> "미완료 $done/$total"
        }

    val progressPercent: Int get() = percent(done, total)
}

internal fun percent(done: Int, total: Int): Int =
    if (total > 0) (done.toDouble() / total * 100).roundToInt() else 0

data class RoutineHistorySummary(val scheduledDays: Int, val completedDays: Int) {
    companion object {
        fun of(days: List<RoutineHistoryDay>): RoutineHistorySummary =
            RoutineHistorySummary(
                scheduledDays = days.count { it.scheduled },
                completedDays = days.count { it.completed },
            )
    }
}

object RoutineFailureCopy {
    enum class Action(val title: String) {
        LOAD("루틴을 불러오지 못했습니다."),
        CREATE("루틴을 저장하지 못했습니다."),
        DELETE("루틴을 삭제하지 못했습니다."),
        ADD_STEP("단계를 추가하지 못했습니다."),
        DELETE_STEP("단계를 삭제하지 못했습니다."),
        CHECK("단계 체크를 저장하지 못했습니다."),
        HISTORY("이력을 불러오지 못했습니다."),
    }

    private const val SERVER_ERROR_MIN = 500

    fun message(action: Action, error: Throwable): String = "${action.title} ${reason(error)}"

    private fun reason(error: Throwable): String = when {
        error is ApiException.Transport -> "네트워크 연결을 확인해 주세요."
        error is ApiException.BadStatus && error.status >= SERVER_ERROR_MIN ->
            "서버가 응답하지 않습니다. 잠시 후 다시 시도해 주세요."
        else -> "잠시 후 다시 시도해 주세요."
    }
}

@Serializable
data class RoutineDraft(
    val name: String,
    val cadence: RoutineCadence,
    val weekdays: Int,
    @SerialName("month_day") val monthDay: Int,
    @SerialName("start_day") val startDay: String,
    val steps: List<String>,
)

@Serializable
internal data class RoutineListResponse(val items: List<Routine>)

@Serializable
internal data class RoutineDayResponse(val items: List<RoutineDay>)

@Serializable
internal data class RoutineHistoryResponse(val items: List<RoutineHistoryDay>)

@Serializable
internal data class RoutineStepPayload(val title: String)

@Serializable
internal data class RoutineCheckPayload(val checked: Boolean)

suspend fun ApiClient.listRoutines(): List<Routine> =
    get("/api/routines", RoutineListResponse.serializer()).items

suspend fun ApiClient.createRoutine(draft: RoutineDraft): Routine =
    post("/api/routines", draft, RoutineDraft.serializer(), Routine.serializer())

suspend fun ApiClient.deleteRoutine(id: Long) {
    delete("/api/routines/$id")
}

suspend fun ApiClient.addRoutineStep(routineId: Long, title: String): Routine =
    post(
        "/api/routines/$routineId/steps",
        RoutineStepPayload(title = title),
        RoutineStepPayload.serializer(),
        Routine.serializer(),
    )

suspend fun ApiClient.deleteRoutineStep(routineId: Long, stepId: Long) {
    delete("/api/routines/$routineId/steps/$stepId")
}

suspend fun ApiClient.listRoutines(day: String): List<RoutineDay> =
    get("/api/routines/days/$day", RoutineDayResponse.serializer()).items

suspend fun ApiClient.setRoutineStep(routineId: Long, stepId: Long, day: String, checked: Boolean): RoutineDay =
    patch(
        "/api/routines/$routineId/days/$day/steps/$stepId",
        RoutineCheckPayload(checked = checked),
        RoutineCheckPayload.serializer(),
        RoutineDay.serializer(),
    )

suspend fun ApiClient.routineHistory(routineId: Long, endingOn: String): List<RoutineHistoryDay> =
    get("/api/routines/$routineId/history/$endingOn", RoutineHistoryResponse.serializer()).items
