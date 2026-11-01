package com.dlddu.pocketaide.prmonitor

import com.dlddu.pocketaide.api.ApiClient
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class HistoryGroupingTest {
    private fun item(
        id: Long,
        createdAt: Long,
        prNumber: Int? = null,
        sha: String = "",
        conclusion: String = "success",
        acked: Long? = null,
        repo: String = "dlddu/pocket-aide",
    ) = NotificationHistoryItem(
        id = id,
        repoFullName = repo,
        prNumber = prNumber,
        headSha = sha,
        conclusion = conclusion,
        acknowledgedAt = acked,
        createdAt = createdAt,
    )

    @Test
    fun groupsByPrThenShaThenRow() {
        val groups = HistoryGrouping.group(
            listOf(
                item(1, 10, prNumber = 42, sha = "aaa"),
                item(2, 20, prNumber = 42, sha = "bbb"),
                item(3, 30, sha = "ccc"),
                item(4, 40, sha = "ccc"),
                item(5, 50),
            ),
        )

        assertEquals(listOf("row:5", "sha:dlddu/pocket-aide:ccc", "pr:dlddu/pocket-aide:42"), groups.map { it.id })
        assertEquals(listOf(4L, 3L), groups[1].items.map { it.id })
        assertEquals("bbb", groups[2].headSha)
    }

    @Test
    fun unacknowledgedGroupsComeFirst() {
        val groups = HistoryGrouping.group(
            listOf(
                item(1, 100, prNumber = 1, acked = 200),
                item(2, 10, prNumber = 2),
            ),
        )

        assertEquals(listOf("pr:dlddu/pocket-aide:2", "pr:dlddu/pocket-aide:1"), groups.map { it.id })
        assertFalse(groups[0].allAcknowledged)
        assertTrue(groups[1].allAcknowledged)
    }

    @Test
    fun countsOutcomes() {
        val group = HistoryGrouping.group(
            listOf(
                item(1, 1, prNumber = 7, conclusion = "success"),
                item(2, 2, prNumber = 7, conclusion = "timed_out"),
                item(3, 3, prNumber = 7, conclusion = "in_progress"),
                item(4, 4, prNumber = 7, conclusion = "neutral", acked = 9),
            ),
        ).single()

        assertEquals(1, group.successCount)
        assertEquals(1, group.failureCount)
        assertEquals(1, group.inProgressCount)
        assertEquals(3, group.unacknowledgedCount)
        assertEquals(4L, group.latestCreatedAt)
        assertTrue(group.contains(3))
        assertFalse(group.contains(null))
    }

    @Test
    fun verdictAndOutcomeFollowIos() {
        assertEquals("CI 통과", CiConclusion.verdictLabel("SUCCESS"))
        assertEquals("CI 시작", CiConclusion.verdictLabel("queued"))
        assertEquals("CI 타임아웃", CiConclusion.verdictLabel("timed_out"))
        assertEquals("CI neutral", CiConclusion.verdictLabel("neutral"))
        assertEquals(CiOutcome.FAILURE, CiConclusion.outcome("neutral"))
        assertEquals(CiOutcome.IN_PROGRESS, CiConclusion.outcome("waiting"))
    }

    @Test
    fun rowTitleFallsBackThroughPrAndBranch() {
        val base = item(1, 1)
        assertEquals("dlddu/pocket-aide · #3 제목", PrMonitorText.rowTitle(base.copy(prNumber = 3, prTitle = "제목")))
        assertEquals("dlddu/pocket-aide · #3", PrMonitorText.rowTitle(base.copy(prNumber = 3, prTitle = "")))
        assertEquals("dlddu/pocket-aide · main", PrMonitorText.rowTitle(base.copy(headBranch = "main")))
        assertEquals("dlddu/pocket-aide", PrMonitorText.rowTitle(base))
        assertEquals(" @abcdef1", PrMonitorText.shortShaSuffix("abcdef1234"))
    }

    @Test
    fun decodesHistoryItemWithMissingOptionalFields() {
        val decoded = ApiClient.json.decodeFromString(
            NotificationHistoryItem.serializer(),
            """{"id":1,"repo_full_name":"a/b","workflow_name":"ci","head_branch":"main","head_sha":"abc","conclusion":"failure","created_at":5}""",
        )

        assertNull(decoded.prNumber)
        assertNull(decoded.acknowledgedAt)
        assertFalse(decoded.isAcknowledged)
    }
}
