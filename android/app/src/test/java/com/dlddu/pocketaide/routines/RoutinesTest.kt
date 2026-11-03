package com.dlddu.pocketaide.routines

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.scratchpad.ScratchpadMoveResult
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Test
import java.io.IOException
import java.time.LocalDate

class RoutinesTest {
    private fun routine(
        cadence: String,
        weekdays: Int = 0,
        monthDay: Int = 0,
        startDay: String = "2026-01-01",
    ): Routine = ApiClient.json.decodeFromString(
        Routine.serializer(),
        """{"id":1,"name":"r","cadence":"$cadence","weekdays":$weekdays,"month_day":$monthDay,"start_day":"$startDay","steps":[],"created_at":1}""",
    )

    @Test
    fun decodesRoutineWithSteps() {
        val json = """{"id":4,"name":"아침 루틴","cadence":"weekly","weekdays":1,"month_day":0,"start_day":"2026-10-01","steps":[{"id":1,"title":"물 한 컵","position":0},{"id":2,"title":"스트레칭","position":1}],"created_at":1700000000}"""
        val routine = ApiClient.json.decodeFromString(Routine.serializer(), json)
        assertEquals(RoutineCadence.WEEKLY, routine.cadence)
        assertEquals(listOf("물 한 컵", "스트레칭"), routine.steps.map { it.title })
        assertEquals("2026-10-01", routine.startDay)
        assertEquals("매주 일요일", routine.scheduleSummary)
    }

    @Test
    fun scheduleSummaries() {
        val monWedFri = RoutineWeekdays.bit(1) or RoutineWeekdays.bit(3) or RoutineWeekdays.bit(5)
        assertEquals("매일", RoutineWeekdays.summary(RoutineCadence.DAILY, 0, 0))
        assertEquals("월·수·금", RoutineWeekdays.summary(RoutineCadence.WEEKDAYS, monWedFri, 0))
        assertEquals(
            "토·일",
            RoutineWeekdays.summary(RoutineCadence.WEEKDAYS, RoutineWeekdays.bit(0) or RoutineWeekdays.bit(6), 0),
        )
        assertEquals("매월 15일", RoutineWeekdays.summary(RoutineCadence.MONTHLY, 0, 15))
    }

    @Test
    fun dayProgressAndCompletion() {
        val json = """{"id":4,"name":"저녁 정리","cadence":"daily","weekdays":0,"month_day":0,"day":"2026-10-01","steps":[{"id":1,"title":"책상","position":0,"checked":true},{"id":2,"title":"옷","position":1,"checked":false}],"done":1,"total":2,"completed":false}"""
        val day = ApiClient.json.decodeFromString(RoutineDay.serializer(), json)
        assertEquals(50, day.progressPercent)
        assertFalse(day.completed)
        assertEquals(listOf(1L), day.steps.filter { it.checked }.map { it.id })
        assertEquals("매일 · 1 / 2 완료", RoutinesCopy.cardSummary(day))
        assertEquals("50%", RoutinesCopy.progressLabel(day))
        assertEquals("대기", RoutinesCopy.progressLabel(day.copy(done = 0)))
        assertEquals("완료", RoutinesCopy.progressLabel(day.copy(done = 2, completed = true)))
    }

    @Test
    fun historyStatusAndSummary() {
        val json = """{"items":[{"day":"2026-09-29","scheduled":true,"done":2,"total":2,"completed":true},{"day":"2026-09-30","scheduled":true,"done":1,"total":2,"completed":false},{"day":"2026-10-01","scheduled":false,"done":0,"total":2,"completed":false}]}"""
        val days = ApiClient.json.decodeFromString(RoutineHistoryResponse.serializer(), json).items
        assertEquals(listOf("완료", "미완료 1/2", "쉬는 날"), days.map { it.statusLabel })
        assertEquals(listOf(100, 50, 0), days.map { it.progressPercent })
        assertEquals(RoutineHistorySummary(scheduledDays = 2, completedDays = 1), RoutineHistorySummary.of(days))
        assertEquals("10월 1일", RoutineDayFormat.title("2026-10-01"))
    }

    @Test
    fun daysUntilNextScheduledDay() {
        assertEquals(1, RoutineSchedule.daysUntilNext(routine("daily"), "2026-10-12"))
        assertEquals(6, RoutineSchedule.daysUntilNext(routine("weekly", weekdays = RoutineWeekdays.bit(0)), "2026-10-12"))
        assertEquals(7, RoutineSchedule.daysUntilNext(routine("weekly", weekdays = RoutineWeekdays.bit(1)), "2026-10-12"))
        assertEquals(
            2,
            RoutineSchedule.daysUntilNext(
                routine("weekdays", weekdays = RoutineWeekdays.bit(3) or RoutineWeekdays.bit(5)),
                "2026-10-12",
            ),
        )
        assertEquals(29, RoutineSchedule.daysUntilNext(routine("monthly", monthDay = 31), "2026-11-01"))
        assertEquals(20, RoutineSchedule.daysUntilNext(routine("monthly", monthDay = 1), "2026-10-12"))
        assertEquals(8, RoutineSchedule.daysUntilNext(routine("daily", startDay = "2026-10-20"), "2026-10-12"))
        assertNull(RoutineSchedule.daysUntilNext(routine("daily"), "bad"))
    }

    @Test
    fun dayFormatUsesCalendarDay() {
        assertEquals("2026-10-02", RoutineDayFormat.string(LocalDate.of(2026, 10, 2)))
        assertEquals(LocalDate.of(2026, 10, 2), RoutineDayFormat.date("2026-10-02"))
        assertNull(RoutineDayFormat.date("2026-13-40"))
    }

    @Test
    fun todayLabelUsesKoreanWeekday() {
        assertEquals("10월 12일 · 월요일", RoutinesCopy.todayLabel(LocalDate.of(2026, 10, 12)))
    }

    @Test
    fun draftEncodesWireKeys() {
        val draft = RoutineDraft(
            name = "월말 정산",
            cadence = RoutineCadence.MONTHLY,
            weekdays = 0,
            monthDay = 31,
            startDay = "2026-10-01",
            steps = listOf("카드 내역"),
        )
        val obj = ApiClient.json.parseToJsonElement(ApiClient.json.encodeToString(RoutineDraft.serializer(), draft)).jsonObject
        assertEquals("31", obj.getValue("month_day").jsonPrimitive.content)
        assertEquals("2026-10-01", obj.getValue("start_day").jsonPrimitive.content)
        assertEquals("monthly", obj.getValue("cadence").jsonPrimitive.content)
    }

    @Test
    fun moveResultCarriesRoutine() {
        val json = """{"target":"routine","routine":{"id":5,"name":"저녁 정리","cadence":"daily","weekdays":0,"month_day":0,"start_day":"2026-10-01","steps":[],"created_at":1}}"""
        val result = ApiClient.json.decodeFromString(ScratchpadMoveResult.serializer(), json)
        val routine = ApiClient.json.decodeFromJsonElement(Routine.serializer(), requireNotNull(result.routine))
        assertEquals("저녁 정리", routine.name)
    }

    @Test
    fun failureCopyIsReadableAndHidesRawError() {
        assertEquals(
            "단계 체크를 저장하지 못했습니다. 네트워크 연결을 확인해 주세요.",
            RoutineFailureCopy.message(RoutineFailureCopy.Action.CHECK, ApiException.Transport(IOException("offline"))),
        )
        val server = RoutineFailureCopy.message(
            RoutineFailureCopy.Action.CHECK,
            ApiException.BadStatus(503, "service unavailable"),
        )
        assertEquals("단계 체크를 저장하지 못했습니다. 서버가 응답하지 않습니다. 잠시 후 다시 시도해 주세요.", server)
        assertFalse(server.contains("503"))
        assertFalse(server.contains("service unavailable"))
        assertEquals(
            "루틴을 불러오지 못했습니다. 잠시 후 다시 시도해 주세요.",
            RoutineFailureCopy.message(RoutineFailureCopy.Action.LOAD, ApiException.BadStatus(404, "not found")),
        )
        assertEquals(
            "이력을 불러오지 못했습니다. 잠시 후 다시 시도해 주세요.",
            RoutineFailureCopy.message(RoutineFailureCopy.Action.HISTORY, IOException("timed out")),
        )
    }
}
