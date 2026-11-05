package com.dlddu.pocketaide.api

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class SseParserTest {
    private fun parse(vararg lines: String): List<SseEvent> {
        val parser = SseParser()
        return lines.mapNotNull { parser.feed(it) }
    }

    @Test
    fun dataOnlyBlockIsAMessageEvent() {
        assertEquals(listOf(SseEvent("message", """{"delta":"안"}""")), parse("""data: {"delta":"안"}""", ""))
    }

    @Test
    fun namedEventCarriesItsName() {
        assertEquals(listOf(SseEvent("done", "{}")), parse("event: done", "data: {}", ""))
    }

    @Test
    fun eventNameDoesNotLeakIntoTheNextBlock() {
        val events = parse("event: error", "data: a", "", "data: b", "")
        assertEquals(listOf(SseEvent("error", "a"), SseEvent("message", "b")), events)
    }

    @Test
    fun multipleDataLinesJoinWithNewline() {
        assertEquals(listOf(SseEvent("message", "a\nb")), parse("data: a", "data: b", ""))
    }

    @Test
    fun commentsAndBlocksWithoutDataAreSkipped() {
        assertEquals(emptyList<SseEvent>(), parse(": keep-alive", "", "event: ping", ""))
    }

    @Test
    fun incompleteBlockEmitsNothingUntilBlankLine() {
        assertNull(SseParser().feed("data: x"))
    }
}
