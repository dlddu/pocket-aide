package com.dlddu.pocketaide.chat

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
}
