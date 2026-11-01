package com.dlddu.pocketaide.todos

import com.dlddu.pocketaide.api.ApiClient
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.time.Clock
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.time.format.DateTimeParseException
import java.time.temporal.WeekFields
import java.util.Locale

@Serializable
enum class TodoArea(val wire: String) {
    @SerialName("personal")
    PERSONAL("personal"),

    @SerialName("work")
    WORK("work"),
}

@Serializable
enum class TodoPriority(val wire: String, val displayName: String) {
    @SerialName("high")
    HIGH("high", "우선"),

    @SerialName("normal")
    NORMAL("normal", "보통"),

    @SerialName("low")
    LOW("low", "낮음"),
}

@Serializable
data class TodoItem(
    val id: Long,
    val title: String,
    val memo: String = "",
    @SerialName("due_date") val dueDate: String? = null,
    val priority: TodoPriority? = null,
    @SerialName("completed_at") val completedAt: Long? = null,
    @SerialName("created_at") val createdAt: Long = 0,
    @SerialName("updated_at") val updatedAt: Long = 0,
) {
    val isDone: Boolean get() = completedAt != null
}

@Serializable
data class TodoDraft(
    val title: String,
    val memo: String = "",
    @SerialName("due_date") val dueDate: String? = null,
    val priority: TodoPriority? = null,
    val done: Boolean = false,
) {
    companion object {
        fun from(item: TodoItem): TodoDraft =
            TodoDraft(
                title = item.title,
                memo = item.memo,
                dueDate = item.dueDate,
                priority = item.priority,
                done = item.isDone,
            )
    }
}

object TodoDueDate {
    private val formatter: DateTimeFormatter = DateTimeFormatter.ISO_LOCAL_DATE

    fun string(date: LocalDate): String = date.format(formatter)

    fun date(string: String): LocalDate? =
        try {
            LocalDate.parse(string, formatter)
        } catch (e: DateTimeParseException) {
            null
        }
}

object TodoSearch {
    fun filter(items: List<TodoItem>, query: String): List<TodoItem> {
        val needle = query.trim()
        if (needle.isEmpty()) return items
        return items.filter {
            it.title.contains(needle, ignoreCase = true) || it.memo.contains(needle, ignoreCase = true)
        }
    }
}

enum class TodoDueFilter(val wire: String) {
    ALL("all"),
    TODAY("today"),
    THIS_WEEK("thisWeek"),
    NO_DATE("noDate"),
    ;

    fun includes(
        item: TodoItem,
        clock: Clock = Clock.systemDefaultZone(),
        weekFields: WeekFields = WeekFields.of(Locale.getDefault()),
    ): Boolean = when (this) {
        ALL -> true
        NO_DATE -> item.dueDate == null
        TODAY -> {
            val due = item.dueDate
            due != null && due <= TodoDueDate.string(LocalDate.now(clock))
        }
        THIS_WEEK -> {
            val due = item.dueDate
            val lastDay = LocalDate.now(clock).with(weekFields.dayOfWeek(), 7L)
            due != null && due <= TodoDueDate.string(lastDay)
        }
    }
}

object TodoDeadline {
    fun openCount(items: List<TodoItem>): Int = items.count { !it.isDone && it.dueDate != null }
}

@Serializable
internal data class TodosListResponse(val items: List<TodoItem>)

suspend fun ApiClient.listTodos(area: TodoArea): List<TodoItem> =
    get("/api/todos/${area.wire}", TodosListResponse.serializer()).items

suspend fun ApiClient.createTodo(area: TodoArea, draft: TodoDraft): TodoItem =
    post("/api/todos/${area.wire}", draft, TodoDraft.serializer(), TodoItem.serializer())

suspend fun ApiClient.updateTodo(area: TodoArea, id: Long, draft: TodoDraft): TodoItem =
    patch("/api/todos/${area.wire}/$id", draft, TodoDraft.serializer(), TodoItem.serializer())

suspend fun ApiClient.deleteTodo(area: TodoArea, id: Long) {
    delete("/api/todos/${area.wire}/$id")
}
