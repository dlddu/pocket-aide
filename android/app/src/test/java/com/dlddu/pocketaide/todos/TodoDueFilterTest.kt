package com.dlddu.pocketaide.todos

import org.junit.Assert.assertEquals
import org.junit.Test
import java.time.Clock
import java.time.DayOfWeek
import java.time.LocalDate
import java.time.ZoneId
import java.time.temporal.WeekFields

class TodoDueFilterTest {
    private val zone: ZoneId = ZoneId.of("Asia/Seoul")
    private val clock: Clock = Clock.fixed(LocalDate.of(2026, 10, 14).atStartOfDay(zone).toInstant(), zone)
    private val mondayWeek: WeekFields = WeekFields.of(DayOfWeek.MONDAY, 1)

    private val items = listOf(
        TodoItem(id = 1, title = "지난 마감", dueDate = "2026-10-13"),
        TodoItem(id = 2, title = "오늘 마감", dueDate = "2026-10-14"),
        TodoItem(id = 3, title = "이번 주 일요일", dueDate = "2026-10-18"),
        TodoItem(id = 4, title = "다음 주", dueDate = "2026-10-19"),
        TodoItem(id = 5, title = "날짜 없음"),
        TodoItem(id = 6, title = "완료된 오늘 마감", dueDate = "2026-10-14", completedAt = 10),
    )

    private fun ids(filter: TodoDueFilter, weekFields: WeekFields = mondayWeek): List<Long> =
        items.filter { filter.includes(it, clock, weekFields) }.map { it.id }

    @Test
    fun allKeepsEverything() {
        assertEquals(listOf(1L, 2L, 3L, 4L, 5L, 6L), ids(TodoDueFilter.ALL))
    }

    @Test
    fun todayKeepsItemsDueTodayOrEarlier() {
        assertEquals(listOf(1L, 2L, 6L), ids(TodoDueFilter.TODAY))
    }

    @Test
    fun thisWeekKeepsItemsDueByTheLastDayOfTheWeek() {
        assertEquals(listOf(1L, 2L, 3L, 6L), ids(TodoDueFilter.THIS_WEEK))
    }

    @Test
    fun thisWeekFollowsTheLocaleFirstWeekday() {
        val sundayWeek = WeekFields.of(DayOfWeek.SUNDAY, 1)
        assertEquals(listOf(1L, 2L, 6L), ids(TodoDueFilter.THIS_WEEK, sundayWeek))
    }

    @Test
    fun noDateKeepsOnlyUndatedItems() {
        assertEquals(listOf(5L), ids(TodoDueFilter.NO_DATE))
    }

    @Test
    fun deadlineCountIsOpenItemsWithDueDate() {
        assertEquals(4, TodoDeadline.openCount(items))
        assertEquals(0, TodoDeadline.openCount(emptyList()))
    }

    @Test
    fun visibleItemsCombineSearchAndDueFilter() {
        val state = TodoListUiState(items = items, query = "마감", dueFilter = TodoDueFilter.TODAY)
        assertEquals(listOf(1L, 2L, 6L), state.visibleItems(clock, mondayWeek).map { it.id })
        assertEquals(5, state.openCount)
        assertEquals(1, state.doneCount)
        assertEquals(4, state.deadlineCount)
    }
}
