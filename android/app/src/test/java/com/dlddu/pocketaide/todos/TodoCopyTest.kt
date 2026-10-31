package com.dlddu.pocketaide.todos

import org.junit.Assert.assertEquals
import org.junit.Test

class TodoCopyTest {
    @Test
    fun summaryReadsLikeANotebookForPersonal() {
        assertEquals("3개 남음 · 2개 완료", TodoArea.PERSONAL.summary(open = 3, done = 2, deadline = 1))
    }

    @Test
    fun summaryReadsLikeAStatusBoardForWork() {
        assertEquals("12 OPEN · 8 DONE · DEADLINE 3", TodoArea.WORK.summary(open = 12, done = 8, deadline = 3))
        assertEquals("0 OPEN · 0 DONE", TodoArea.WORK.summary(open = 0, done = 0, deadline = 0))
    }

    @Test
    fun dueFilterChipTitles() {
        assertEquals(
            listOf("전체 · 11", "오늘", "이번 주", "날짜 없음"),
            TodoDueFilter.entries.map { it.chipTitle(11) },
        )
    }

    @Test
    fun personalMetaHidesHighPriorityBecauseTheTagShowsIt() {
        val item = TodoItem(id = 1, title = "a", memo = "임시공간에서 이동됨", dueDate = "2026-10-14", priority = TodoPriority.HIGH)
        assertEquals(listOf("2026-10-14", "메모: 임시공간에서 이동됨"), TodoRowMetaText.visibleParts(TodoArea.PERSONAL, item))
        assertEquals(
            listOf("2026-10-14", "우선", "메모: 임시공간에서 이동됨"),
            TodoRowMetaText.labelParts(TodoArea.PERSONAL, item),
        )
    }

    @Test
    fun workMetaRanksPriority() {
        val item = TodoItem(id = 1, title = "a", priority = TodoPriority.NORMAL)
        assertEquals(listOf("P2"), TodoRowMetaText.visibleParts(TodoArea.WORK, item))
        assertEquals(listOf("P1"), TodoRowMetaText.visibleParts(TodoArea.WORK, item.copy(priority = TodoPriority.HIGH)))
        assertEquals(listOf("P3"), TodoRowMetaText.visibleParts(TodoArea.WORK, item.copy(priority = TodoPriority.LOW)))
    }

    @Test
    fun areaTitlesAndPrompts() {
        assertEquals("내 일", TodoArea.PERSONAL.screenTitle)
        assertEquals("회사", TodoArea.WORK.screenTitle)
        assertEquals("개인 영역만 검색…", TodoArea.PERSONAL.searchPrompt)
        assertEquals("회사 영역만 검색…", TodoArea.WORK.searchPrompt)
    }
}
