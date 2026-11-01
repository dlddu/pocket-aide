package com.dlddu.pocketaide.affirmations

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class AffirmationSortTest {
    private fun aff(id: Long, priority: AffirmationPriority, createdAt: Long) =
        Affirmation(id = id, text = "t$id", priority = priority, createdAt = createdAt, updatedAt = createdAt)

    private val pool = listOf(
        aff(1, AffirmationPriority.LOW, 300),
        aff(2, AffirmationPriority.HIGH, 100),
        aff(3, AffirmationPriority.NORMAL, 400),
        aff(4, AffirmationPriority.HIGH, 200),
        aff(5, AffirmationPriority.NORMAL, 400),
    )

    @Test
    fun priorityPutsHighFirstThenNewest() {
        assertEquals(listOf(4L, 2L, 5L, 3L, 1L), AffirmationSortOrder.PRIORITY.sorted(pool).map { it.id })
    }

    @Test
    fun newestIgnoresPriority() {
        assertEquals(listOf(5L, 3L, 1L, 4L, 2L), AffirmationSortOrder.NEWEST.sorted(pool).map { it.id })
    }

    @Test
    fun emptyStaysEmpty() {
        assertTrue(AffirmationSortOrder.PRIORITY.sorted(emptyList()).isEmpty())
        assertTrue(AffirmationSortOrder.NEWEST.sorted(emptyList()).isEmpty())
    }

    @Test
    fun labelsMatchIos() {
        assertEquals(listOf("우선순위 순", "최신순"), AffirmationSortOrder.entries.map { it.label })
        assertEquals(listOf("priority", "newest"), AffirmationSortOrder.entries.map { it.wire })
    }
}
