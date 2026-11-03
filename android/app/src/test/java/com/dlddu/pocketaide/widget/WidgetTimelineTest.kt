package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.affirmations.AffirmationPriority
import com.dlddu.pocketaide.affirmations.RotationSelector
import com.dlddu.pocketaide.affirmations.SeededRng
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.File
import java.time.Duration
import java.time.Instant

class WidgetTimelineTest {
    private fun aff(id: Long, priority: AffirmationPriority) =
        Affirmation(id = id, text = "t$id", priority = priority, createdAt = 0, updatedAt = 0)

    private val pool = listOf(
        aff(1, AffirmationPriority.HIGH),
        aff(2, AffirmationPriority.NORMAL),
        aff(3, AffirmationPriority.LOW),
    )

    @Test
    fun refreshIntervalIsThirtyMinutes() {
        assertEquals(Duration.ofMinutes(30), WidgetTimeline.REFRESH_INTERVAL)
    }

    @Test
    fun providerRefreshesEveryIntervalRegardlessOfOutcome() {
        val info = File("src/main/res/xml/pocket_aide_widget_info.xml").readText()
        val period = Regex("updatePeriodMillis=\"(\\d+)\"").find(info)?.groupValues?.get(1)?.toLong()
        assertEquals(WidgetTimeline.REFRESH_INTERVAL.toMillis(), period)
    }

    @Test
    fun slotStartAlignsToInterval() {
        val start = Instant.ofEpochSecond(1_800_000_000L)
        assertEquals(start, WidgetTimeline.slotStart(start))
        assertEquals(start, WidgetTimeline.slotStart(start.plusSeconds(29 * 60 + 59)))
        assertEquals(start.plusSeconds(30 * 60), WidgetTimeline.slotStart(start.plusSeconds(30 * 60)))
    }

    @Test
    fun pickUsesSameSelectorAndSeedAsApp() {
        val at = Instant.ofEpochSecond(1_800_000_000L + 600)
        val expected = RotationSelector().pick(pool, SeededRng(1_800_000_000uL))
        assertEquals(WidgetAffirmationState.Loaded(expected!!), WidgetTimeline.pick(pool, at))
    }

    @Test
    fun sameSlotKeepsSamePick() {
        val start = Instant.ofEpochSecond(1_800_000_000L)
        for (slot in 0 until 48) {
            val slotStart = start.plus(WidgetTimeline.REFRESH_INTERVAL.multipliedBy(slot.toLong()))
            assertEquals(
                WidgetTimeline.pick(pool, slotStart),
                WidgetTimeline.pick(pool, slotStart.plusSeconds(17 * 60)),
            )
        }
    }

    @Test
    fun rotationFollowsPriorityWeights() {
        val start = Instant.ofEpochSecond(1_800_000_000L)
        val counts = (0 until 3000).map { slot ->
            val state = WidgetTimeline.pick(pool, start.plus(WidgetTimeline.REFRESH_INTERVAL.multipliedBy(slot.toLong())))
            (state as WidgetAffirmationState.Loaded).affirmation.id
        }.groupingBy { it }.eachCount()
        assertTrue(counts.getValue(1) > counts.getValue(2))
        assertTrue(counts.getValue(2) > counts.getValue(3))
    }

    @Test
    fun emptyPoolIsEmptyState() {
        assertEquals(WidgetAffirmationState.Empty, WidgetTimeline.pick(emptyList(), Instant.EPOCH))
    }
}
