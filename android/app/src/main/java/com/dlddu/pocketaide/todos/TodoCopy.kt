package com.dlddu.pocketaide.todos

val TodoArea.screenTitle: String
    get() = when (this) {
        TodoArea.PERSONAL -> "내 일"
        TodoArea.WORK -> "회사"
    }

val TodoArea.searchPrompt: String
    get() = when (this) {
        TodoArea.PERSONAL -> "개인 영역만 검색…"
        TodoArea.WORK -> "회사 영역만 검색…"
    }

fun TodoArea.summary(open: Int, done: Int, deadline: Int): String = when (this) {
    TodoArea.PERSONAL -> "${open}개 남음 · ${done}개 완료"
    TodoArea.WORK ->
        if (deadline > 0) "$open OPEN · $done DONE · DEADLINE $deadline" else "$open OPEN · $done DONE"
}

fun TodoDueFilter.chipTitle(total: Int): String = when (this) {
    TodoDueFilter.ALL -> "전체 · $total"
    TodoDueFilter.TODAY -> "오늘"
    TodoDueFilter.THIS_WEEK -> "이번 주"
    TodoDueFilter.NO_DATE -> "날짜 없음"
}

object TodoRowMetaText {
    fun labelParts(area: TodoArea, item: TodoItem): List<String> {
        val parts = mutableListOf<String>()
        item.dueDate?.let { parts.add(it) }
        item.priority?.let { parts.add(if (area == TodoArea.WORK) "P${workRank(it)}" else it.displayName) }
        if (item.memo.isNotEmpty()) parts.add("메모: ${item.memo}")
        return parts
    }

    fun visibleParts(area: TodoArea, item: TodoItem): List<String> {
        val parts = labelParts(area, item)
        if (area != TodoArea.PERSONAL || item.priority != TodoPriority.HIGH) return parts
        return parts.filter { it != TodoPriority.HIGH.displayName }
    }

    private fun workRank(priority: TodoPriority): Int = when (priority) {
        TodoPriority.HIGH -> 1
        TodoPriority.NORMAL -> 2
        TodoPriority.LOW -> 3
    }
}
