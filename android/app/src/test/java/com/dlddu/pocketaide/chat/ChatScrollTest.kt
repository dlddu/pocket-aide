package com.dlddu.pocketaide.chat

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ChatScrollTest {
    private val slack = 48

    private fun atEnd(index: Int, end: Int, total: Int = 3, viewportEnd: Int = 1000) =
        isWithinListEnd(
            lastVisibleIndex = index,
            lastVisibleEndPx = end,
            totalCount = total,
            viewportEndPx = viewportEnd,
            slackPx = slack,
        )

    @Test
    fun theEndOfTheLastMessageAtTheViewportBottomIsTheEnd() {
        assertTrue(atEnd(index = 2, end = 1000))
    }

    @Test
    fun aLastMessageStillGrowingPastTheFoldIsNotTheEnd() {
        assertFalse(atEnd(index = 2, end = 4200))
    }

    @Test
    fun aGapUpToTheSlackCountsAsTheEndSoAFlingThatStopsShortKeepsFollowing() {
        assertTrue(atEnd(index = 2, end = 1048))
        assertFalse(atEnd(index = 2, end = 1049))
    }

    @Test
    fun readingAnEarlierMessageIsNotTheEndEvenWhenItEndsInsideTheViewport() {
        assertFalse(atEnd(index = 0, end = 400))
        assertFalse(atEnd(index = 1, end = 900))
    }

    @Test
    fun theOnlyMessageIsTheLastOne() {
        assertTrue(atEnd(index = 0, end = 300, total = 1))
    }

    private fun history(vararg ids: Long) =
        ids.map { ChatMessage(id = it, role = ChatMessage.ROLE_USER, content = "메시지 $it") }

    @Test
    fun theAnchoredMessageIsFoundWhereItNowSitsAndNotWhereItSat() {
        // The visit left on message 7. An index would send the reader back to
        // the same slot, which holds a different message once a sync has
        // brought history the visit never saw.
        val anchor = ChatScrollAnchor(messageId = 7, offset = 40)
        assertEquals(2, anchoredIndex(history(5, 6, 7), anchor))
        assertEquals(6, anchoredIndex(history(1, 2, 3, 4, 5, 6, 7), anchor))
    }

    @Test
    fun aDeletedAnchorMessageHasNoPlaceSoTheVisitLandsAtTheEnd() {
        assertEquals(-1, anchoredIndex(history(1, 2, 4), ChatScrollAnchor(messageId = 3, offset = 0)))
    }

    @Test
    fun aConversationThatStartedOverHasNoneOfTheAnchoredIdsSoTheVisitLandsAtTheEnd() {
        assertEquals(-1, anchoredIndex(history(90, 91), ChatScrollAnchor(messageId = 3, offset = 0)))
    }

    @Test
    fun nothingAnchoredIsTheEnd() {
        assertEquals(-1, anchoredIndex(history(1, 2, 3), null))
        assertEquals(-1, anchoredIndex(emptyList(), ChatScrollAnchor(messageId = 1, offset = 0)))
    }
}
