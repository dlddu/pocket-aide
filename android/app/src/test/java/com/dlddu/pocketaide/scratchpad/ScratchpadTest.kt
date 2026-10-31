package com.dlddu.pocketaide.scratchpad

import com.dlddu.pocketaide.affirmations.AffirmationPriority
import com.dlddu.pocketaide.api.ApiClient
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.time.Clock
import java.time.ZoneId
import java.time.ZonedDateTime

class ScratchpadTest {
    private val zone: ZoneId = ZoneId.of("Asia/Seoul")

    private fun epoch(day: Int, hour: Int, minute: Int = 0): Long =
        ZonedDateTime.of(2026, 9, day, hour, minute, 0, 0, zone).toEpochSecond()

    private val clock: Clock = Clock.fixed(ZonedDateTime.of(2026, 9, 29, 15, 0, 0, 0, zone).toInstant(), zone)

    @Test
    fun groupsByDayNewestFirst() {
        val items = listOf(
            ScratchpadItem(id = 1, text = "어제 밤", capturedAt = epoch(28, 22, 51)),
            ScratchpadItem(id = 2, text = "오늘 아침", source = ScratchpadSource.SHORTCUT, capturedAt = epoch(29, 9, 13)),
            ScratchpadItem(id = 3, text = "오늘 오후", capturedAt = epoch(29, 14, 8)),
            ScratchpadItem(id = 4, text = "지난주", capturedAt = epoch(22, 10)),
        )
        val sections = ScratchpadSections.group(items, clock)
        assertEquals(listOf("오늘", "어제", "9월 22일"), sections.map { it.title })
        assertEquals(listOf(listOf(3L, 2L), listOf(1L), listOf(4L)), sections.map { section -> section.items.map { it.id } })
    }

    @Test
    fun sameCaptureTimeOrdersByIdDescending() {
        val items = listOf(
            ScratchpadItem(id = 1, text = "a", capturedAt = epoch(29, 9)),
            ScratchpadItem(id = 2, text = "b", capturedAt = epoch(29, 9)),
        )
        assertEquals(listOf(2L, 1L), ScratchpadSections.group(items, clock).single().items.map { it.id })
    }

    @Test
    fun emptyListHasNoSections() {
        assertTrue(ScratchpadSections.group(emptyList(), clock).isEmpty())
    }

    @Test
    fun timeLabelUsesTwentyFourHourClock() {
        val item = ScratchpadItem(id = 1, text = "x", capturedAt = epoch(28, 22, 51))
        assertEquals("22:51", ScratchpadSections.timeLabel(item, zone))
    }

    @Test
    fun decodesServerItemMetadata() {
        val json = """{"id":7,"text":"retention fact-check","source":"shortcut","captured_at":1700000000,"created_at":1700000001}"""
        val item = ApiClient.json.decodeFromString(ScratchpadItem.serializer(), json)
        assertEquals(ScratchpadSource.SHORTCUT, item.source)
        assertEquals(1_700_000_000L, item.capturedAt)
        assertEquals("숏컷 · 음성", item.source.displayName)
    }

    @Test
    fun decodesMoveResultForEachTarget() {
        val todoJson =
            """{"target":"work","todo":{"id":3,"title":"메일","memo":"","due_date":null,"priority":null,"completed_at":null,"created_at":1,"updated_at":1}}"""
        val todo = ApiClient.json.decodeFromString(ScratchpadMoveResult.serializer(), todoJson)
        assertEquals(ScratchpadMoveTarget.WORK, todo.target)
        assertEquals("메일", todo.todo?.title)
        assertNull(todo.affirmation)

        val affJson =
            """{"target":"affirmation","affirmation":{"id":9,"text":"매일 1%","priority":"normal","created_at":1,"updated_at":1}}"""
        val aff = ApiClient.json.decodeFromString(ScratchpadMoveResult.serializer(), affJson)
        assertEquals(AffirmationPriority.NORMAL, aff.affirmation?.priority)
        assertNull(aff.todo)

        val routineJson = """{"target":"routine","routine":{"id":4,"name":"아침","cadence":"daily","steps":[]}}"""
        val routine = ApiClient.json.decodeFromString(ScratchpadMoveResult.serializer(), routineJson)
        assertEquals(ScratchpadMoveTarget.ROUTINE, routine.target)
        assertTrue(routine.routine != null)
    }

    @Test
    fun moveTargetChipLabels() {
        assertEquals(
            listOf("→ 개인", "→ 회사", "→ 다짐", "→ 루틴"),
            ScratchpadMoveTarget.entries.map { it.chipLabel },
        )
    }

    @Test
    fun widgetCountTextShowsUnclassifiedCount() {
        assertEquals("미분류 3개", WidgetScratchpad.countText(3))
        assertEquals("미분류 1개", WidgetScratchpad.countText(1))
    }

    @Test
    fun widgetCountTextMarksEmptyScratchpad() {
        assertEquals("정리할 항목 없음", WidgetScratchpad.countText(0))
    }
}
