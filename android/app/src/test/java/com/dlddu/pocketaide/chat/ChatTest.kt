package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.api.SseEvent
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test
import java.time.LocalDate
import java.time.ZoneId
import java.time.ZoneOffset

class ChatTest {
    private val utc: ZoneId = ZoneOffset.UTC

    @Test
    fun deltaEventDecodes() {
        assertEquals(ChatStreamEvent.Delta("안녕"), ChatStreamEvents.decode(SseEvent("message", """{"delta":"안녕"}""")))
    }

    @Test
    fun doneEventDecodesTheStoredMessage() {
        val event = ChatStreamEvents.decode(
            SseEvent("done", """{"message":{"id":7,"role":"assistant","content":"답","created_at":5}}"""),
        )
        assertEquals(ChatStreamEvent.Done(ChatMessage(7, ChatRole.ASSISTANT, "답", 5)), event)
    }

    @Test
    fun doneWithNullMessageDecodes() {
        assertEquals(ChatStreamEvent.Done(null), ChatStreamEvents.decode(SseEvent("done", """{"message":null}""")))
    }

    @Test
    fun errorEventDecodes() {
        assertEquals(ChatStreamEvent.Failed("reply failed"), ChatStreamEvents.decode(SseEvent("error", """{"message":"reply failed"}""")))
    }

    @Test
    fun unknownOrMalformedEventsAreIgnored() {
        assertNull(ChatStreamEvents.decode(SseEvent("ping", "{}")))
        assertNull(ChatStreamEvents.decode(SseEvent("message", "not json")))
    }

    @Test
    fun serviceUnavailableMapsToUnconfiguredCopy() {
        assertEquals(ChatCopy.UNCONFIGURED, ChatCopy.errorMessage(ApiException.BadStatus(503, "chat is not configured")))
    }

    @Test
    fun blankTitleFallsBackToNewConversation() {
        assertEquals("새 대화", ChatCopy.sessionTitle(ChatSession(1)))
        assertEquals("주제", ChatCopy.sessionTitle(ChatSession(1, "주제")))
    }

    @Test
    fun dayLabelsAreRelativeToToday() {
        val today = LocalDate.of(2026, 11, 5)
        val noon = today.atStartOfDay(utc).toEpochSecond() + 12 * 3600
        assertEquals("오늘", ChatTimes.dayLabel(noon, today, utc))
        assertEquals("어제", ChatTimes.dayLabel(noon - 86_400, today, utc))
        assertEquals("11월 1일", ChatTimes.dayLabel(noon - 4 * 86_400, today, utc))
        assertEquals("오늘 · 12:00", ChatTimes.stamp(noon, today, utc))
    }
}
