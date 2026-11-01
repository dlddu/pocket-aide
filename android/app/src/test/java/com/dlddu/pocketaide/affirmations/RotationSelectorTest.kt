package com.dlddu.pocketaide.affirmations

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import kotlin.math.abs

class RotationSelectorTest {
    private fun aff(id: Long, priority: AffirmationPriority) =
        Affirmation(id = id, text = "t$id", priority = priority, createdAt = 0, updatedAt = 0)

    @Test
    fun seededRngIsSplitmix64() {
        assertEquals(0x910A_2DEC_8902_5CC1uL, SeededRng(1uL).next())
    }

    @Test
    fun multiplyFullKeepsHighBits() {
        assertEquals(ULong.MAX_VALUE - 1uL to 1uL, RandomSource.multiplyFull(ULong.MAX_VALUE, ULong.MAX_VALUE))
    }

    @Test
    fun emptyPoolReturnsNull() {
        assertNull(RotationSelector().pick(emptyList(), SeededRng(1uL)))
    }

    @Test
    fun singleItemAlwaysReturnsIt() {
        val pool = listOf(aff(1, AffirmationPriority.NORMAL))
        for (seed in 1..32) {
            assertEquals(1L, RotationSelector().pick(pool, SeededRng(seed.toULong()))?.id)
        }
    }

    @Test
    fun sameSeedIsDeterministic() {
        val pool = listOf(
            aff(1, AffirmationPriority.HIGH),
            aff(2, AffirmationPriority.NORMAL),
            aff(3, AffirmationPriority.LOW),
            aff(4, AffirmationPriority.HIGH),
        )
        val selector = RotationSelector()
        assertEquals(selector.pick(pool, SeededRng(42uL))?.id, selector.pick(pool, SeededRng(42uL))?.id)
    }

    @Test
    fun weightedDistribution() {
        val selector = RotationSelector()
        val pool = listOf(aff(1, AffirmationPriority.HIGH), aff(2, AffirmationPriority.NORMAL), aff(3, AffirmationPriority.LOW))
        val rng = SeededRng(12345uL)
        val counts = mutableMapOf(1L to 0, 2L to 0, 3L to 0)
        val trials = 6000
        repeat(trials) { selector.pick(pool, rng)?.let { counts[it.id] = counts.getValue(it.id) + 1 } }
        val expected = mapOf(1L to trials * 3.0 / 6, 2L to trials * 2.0 / 6, 3L to trials * 1.0 / 6)
        for ((id, want) in expected) {
            val drift = abs(counts.getValue(id) - want) / want
            assertTrue("id=$id got=${counts[id]} want=$want", drift < 0.10)
        }
    }

    @Test
    fun zeroWeightItemsAreSkipped() {
        val selector = RotationSelector(
            mapOf(AffirmationPriority.HIGH to 5, AffirmationPriority.NORMAL to 0, AffirmationPriority.LOW to 0),
        )
        val pool = listOf(
            aff(1, AffirmationPriority.HIGH),
            aff(2, AffirmationPriority.NORMAL),
            aff(3, AffirmationPriority.LOW),
            aff(4, AffirmationPriority.NORMAL),
        )
        val rng = SeededRng(7uL)
        repeat(200) { assertEquals(1L, selector.pick(pool, rng)?.id) }
    }

    @Test
    fun uiTestSeedRanksPrioritiesForEveryPoolOrder() {
        val high = AffirmationPriority.HIGH
        val normal = AffirmationPriority.NORMAL
        val low = AffirmationPriority.LOW
        val orders = listOf(
            listOf(high, normal, low), listOf(high, low, normal),
            listOf(normal, high, low), listOf(normal, low, high),
            listOf(low, high, normal), listOf(low, normal, high),
        )
        for (order in orders) {
            for (skipped in 0..3) {
                val counts = rotationCounts(RotationSelector(), order, skipped)
                assertTrue("order=$order skipped=$skipped counts=$counts", ranksByPriority(counts))
            }
        }
        val uniform = RotationSelector(mapOf(high to 1, normal to 1, low to 1))
        for (skipped in 0..3) {
            assertFalse(ranksByPriority(rotationCounts(uniform, listOf(low, normal, high), skipped)))
        }
    }

    private fun rotationCounts(
        selector: RotationSelector,
        order: List<AffirmationPriority>,
        skipped: Int,
    ): Map<AffirmationPriority, Int> {
        val pool = order.mapIndexed { index, priority -> aff(index + 1L, priority) }
        val rng = SeededRng(2009uL)
        repeat(skipped) { selector.pick(pool, rng) }
        val counts = AffirmationPriority.entries.associateWith { 0 }.toMutableMap()
        repeat(30) { selector.pick(pool, rng)?.let { counts[it.priority] = counts.getValue(it.priority) + 1 } }
        return counts
    }

    private fun ranksByPriority(counts: Map<AffirmationPriority, Int>): Boolean {
        val high = counts.getValue(AffirmationPriority.HIGH)
        val normal = counts.getValue(AffirmationPriority.NORMAL)
        val low = counts.getValue(AffirmationPriority.LOW)
        return high > normal && normal > low && low > 0
    }
}
