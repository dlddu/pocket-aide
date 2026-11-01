package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookHistoryGroupingTest : E2eTest() {
    private val workflows = listOf("Build", "Lint", "Deploy", "Docs")
    private val clock = Regex("^[0-9]{2}:[0-9]{2}$")

    private fun events(repo: String, token: String): List<WebhookEvent> {
        val first = Webhooks.commit(1, token)
        val direct = Webhooks.commit(2, token)
        val second = Webhooks.commit(3, token)
        return listOf(
            WebhookEvent(repo, 1, first, "feature/one", "Build", "success"),
            WebhookEvent(repo, null, direct, "main", "Deploy", "success"),
            WebhookEvent(repo, 2, second, "feature/two", "Build", "success"),
            WebhookEvent(repo, 1, first, "feature/one", "Lint", "failure"),
            WebhookEvent(repo, null, direct, "main", "Docs", "success"),
        )
    }

    private fun groupOrder(sent: List<WebhookEvent>): List<String> =
        sent.reversed().map { it.groupKey }.distinct()

    private fun title(lead: WebhookEvent): String =
        if (lead.pullRequest != null) "${lead.repo} · #${lead.pullRequest}" else "${lead.repo} · ${lead.branch} @${lead.commit.take(7)}"

    private fun assertHeader(labels: List<String>, members: List<WebhookEvent>, group: String) {
        val lead = members.first()
        val passed = members.count { it.conclusion == "success" }
        val failed = members.size - passed
        assertTrue("Group $group should be titled ${title(lead)}: $labels", labels.contains(title(lead)))
        assertTrue("Group $group should count ${members.size} items: $labels", labels.contains("CI ${members.size}건"))
        assertTrue("Group $group should count ${members.size} unacknowledged: $labels", labels.contains("미확인 ${members.size}"))
        assertEquals("Group $group should summarise $passed passed: $labels", passed > 0, labels.contains("통과 $passed"))
        assertEquals("Group $group should summarise $failed failed: $labels", failed > 0, labels.contains("실패 $failed"))
        assertTrue("Group $group should show its latest event time: $labels", labels.any { it.startsWith("최근 ") })
        assertEquals(
            "Only the PR-less group should read PR 없음 · ${lead.branch} 직접 푸시: $labels",
            lead.pullRequest == null,
            labels.contains("PR 없음 · ${lead.branch} 직접 푸시"),
        )
    }

    private fun normalized(labels: List<String>, repo: String, token: String): List<String> =
        labels
            .filter { !it.startsWith("최근 ") && !clock.matches(it) && !it.startsWith("확인 ") }
            .map { it.replace(repo, "R").replace(token.take(6), "T") }
            .sorted()

    private fun composition(sent: List<WebhookEvent>, repo: String, token: String): List<List<String>> {
        val order = groupOrder(sent)
        val top = order.first()
        val count = sent.count { it.groupKey == top }
        val scenario = PrMonitor.openUntil(compose, top, "CI ${count}건")
        val seen = mutableSetOf<String>()
        val result = sortedMapOf<String, List<String>>()
        for (group in order) {
            assertTrue("Group $group should be listed", PrMonitor.reveal(compose, group))
            val labels = PrMonitor.labels(compose, group)
            val members = sent.filter { it.groupKey == group }
            assertHeader(labels, members, group)
            assertEquals(
                "Group $group should list its items newest first: $labels",
                members.reversed().map { it.workflow },
                labels.filter { it in workflows },
            )
            seen += PrMonitor.groupsOf(compose, repo)
            result[group.replace(repo, "R").replace(token, "T")] = normalized(labels, repo, token)
        }
        assertEquals("The five events of $repo should form exactly these group cards", order.toSet(), seen)
        assertEquals("Two PRs and one PR-less commit should make three groups", 3, order.size)
        scenario.close()
        return result.values.toList()
    }

    @Test
    fun eventsOfTheSamePullRequestOrCommitShareOneGroupCardWhateverTheArrivalOrder() {
        val token = Webhooks.runToken()
        val repo = "dlddu/e2e-grouping-$token"
        val sent = events(repo, token)
        Webhooks.send(sent)
        val first = composition(sent, repo, token)

        val otherToken = Webhooks.runToken()
        val otherRepo = "dlddu/e2e-regrouped-$otherToken"
        val reordered = events(otherRepo, otherToken).reversed()
        Webhooks.send(reordered)
        val second = composition(reordered, otherRepo, otherToken)

        assertEquals("A different arrival order should give the same groups", first, second)
    }
}
