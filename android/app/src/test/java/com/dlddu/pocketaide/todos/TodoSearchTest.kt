package com.dlddu.pocketaide.todos

import com.dlddu.pocketaide.api.ApiClient
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.boolean
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.time.LocalDate

class TodoSearchTest {
    private val items = listOf(
        TodoItem(id = 1, title = "치과 예약 잡기"),
        TodoItem(id = 2, title = "Retention fact-check", memo = "임시공간에서 이동됨"),
        TodoItem(id = 3, title = "책장 정리", completedAt = 10),
    )

    @Test
    fun emptyQueryReturnsEverything() {
        assertEquals(listOf(1L, 2L, 3L), TodoSearch.filter(items, "  ").map { it.id })
    }

    @Test
    fun matchesTitleCaseInsensitively() {
        assertEquals(listOf(2L), TodoSearch.filter(items, "retention").map { it.id })
    }

    @Test
    fun matchesMemo() {
        assertEquals(listOf(2L), TodoSearch.filter(items, "임시공간").map { it.id })
    }

    @Test
    fun noMatchReturnsEmpty() {
        assertTrue(TodoSearch.filter(items, "분기 리뷰").isEmpty())
    }

    @Test
    fun dueDateRoundTrip() {
        val date = TodoDueDate.date("2026-10-02")
        assertEquals(LocalDate.of(2026, 10, 2), date)
        assertEquals("2026-10-02", date?.let { TodoDueDate.string(it) })
        assertNull(TodoDueDate.date("10/02"))
    }

    @Test
    fun draftEncodingOmitsUnsetOptionals() {
        val raw = ApiClient.json.encodeToString(TodoDraft.serializer(), TodoDraft(title = "x"))
        val obj: JsonObject = ApiClient.json.parseToJsonElement(raw).jsonObject
        assertEquals("x", obj.getValue("title").jsonPrimitive.content)
        assertFalse(obj.getValue("done").jsonPrimitive.boolean)
        assertFalse(obj.containsKey("due_date"))
        assertFalse(obj.containsKey("priority"))
    }

    @Test
    fun draftEncodingUsesWireNames() {
        val raw = ApiClient.json.encodeToString(
            TodoDraft.serializer(),
            TodoDraft(title = "x", dueDate = "2026-10-02", priority = TodoPriority.HIGH, done = true),
        )
        assertEquals(
            """{"title":"x","memo":"","due_date":"2026-10-02","priority":"high","done":true}""",
            raw,
        )
    }

    @Test
    fun draftFromItemCarriesDoneState() {
        val draft = TodoDraft.from(TodoItem(id = 9, title = "a", memo = "m", dueDate = "2026-01-01", completedAt = 5))
        assertEquals(TodoDraft(title = "a", memo = "m", dueDate = "2026-01-01", priority = null, done = true), draft)
    }

    @Test
    fun decodesServerItem() {
        val json =
            """{"id":3,"title":"메일","memo":"","due_date":null,"priority":"low","completed_at":null,"created_at":1,"updated_at":2}"""
        val item = ApiClient.json.decodeFromString(TodoItem.serializer(), json)
        assertEquals(TodoPriority.LOW, item.priority)
        assertNull(item.dueDate)
        assertFalse(item.isDone)
        assertEquals(2L, item.updatedAt)
    }
}
