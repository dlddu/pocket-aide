package com.dlddu.pocketaide.e2e

import androidx.compose.ui.semantics.SemanticsNode
import androidx.compose.ui.semantics.SemanticsProperties
import androidx.compose.ui.semantics.getOrNull
import androidx.compose.ui.test.hasAnyAncestor
import androidx.compose.ui.test.hasScrollToNodeAction
import androidx.compose.ui.test.hasTestTag
import androidx.compose.ui.test.junit4.ComposeTestRule
import androidx.compose.ui.test.performScrollToIndex
import androidx.compose.ui.test.performScrollToNode
import androidx.test.core.app.ActivityScenario
import com.dlddu.pocketaide.MainActivity
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import java.util.UUID
import java.util.concurrent.TimeUnit
import kotlin.random.Random

data class WebhookEvent(
    val repo: String,
    val pullRequest: Int?,
    val commit: String,
    val branch: String,
    val workflow: String,
    val conclusion: String,
) {
    val groupKey: String get() = if (pullRequest != null) "pr:$repo:$pullRequest" else "sha:$repo:$commit"

    val pushTitle: String
        get() {
            val verdict = when (conclusion) {
                "success" -> "CI 통과"
                "failure" -> "CI 실패"
                else -> "CI $conclusion"
            }
            return if (pullRequest != null) "$verdict — $repo #$pullRequest" else "$repo — $conclusion"
        }

    fun envelope(): JSONObject {
        val runId = Random.nextLong(100_000_000, 999_999_999)
        val pulls = JSONArray()
        if (pullRequest != null) {
            pulls.put(
                JSONObject()
                    .put("id", runId + 1)
                    .put("number", pullRequest)
                    .put("url", "${Webhooks.GITHUB_API}/repos/$repo/pulls/$pullRequest")
                    .put("head", JSONObject().put("ref", branch).put("sha", commit))
                    .put("base", JSONObject().put("ref", "main").put("sha", "0".repeat(40))),
            )
        }
        val run = JSONObject()
            .put("id", runId)
            .put("name", workflow)
            .put("head_branch", branch)
            .put("head_sha", commit)
            .put("event", if (pullRequest == null) "push" else "pull_request")
            .put("status", "completed")
            .put("conclusion", conclusion)
            .put("run_number", 1)
            .put("html_url", "${Webhooks.GITHUB_WEB}/$repo/actions/runs/$runId")
            .put("pull_requests", pulls)
        val repository = JSONObject()
            .put("id", runId + 2)
            .put("name", repo.substringAfterLast('/'))
            .put("full_name", repo)
            .put("html_url", "${Webhooks.GITHUB_WEB}/$repo")
        return JSONObject()
            .put("action", "completed")
            .put("workflow_run", run)
            .put("repository", repository)
            .put("sender", JSONObject().put("login", "dlddu"))
    }
}

object Webhooks {
    const val GITHUB_API = "https://api.github.com"
    const val GITHUB_WEB = "https://github.com"
    private const val SQS_ENDPOINT = "http://127.0.0.1:4566/"
    private const val QUEUE_URL = "http://localhost:4566/123456789012/pocket-aide-e2e"
    private const val AUTHORIZATION = "AWS4-HMAC-SHA256 Credential=test/20260101/us-east-1/sqs/aws4_request, " +
        "SignedHeaders=host;x-amz-target, Signature="
    private val http = OkHttpClient.Builder().callTimeout(20, TimeUnit.SECONDS).build()
    private val amzJson = "application/x-amz-json-1.0".toMediaType()

    fun runToken(): String = UUID.randomUUID().toString().take(8).lowercase()

    fun commit(seed: Int, token: String): String = (seed.toString() + token.repeat(5)).take(40)

    fun send(events: List<WebhookEvent>) {
        events.forEach {
            send(it)
            Thread.sleep(1_500)
        }
    }

    fun send(event: WebhookEvent) {
        val message = JSONObject()
            .put("QueueUrl", QUEUE_URL)
            .put("MessageBody", event.envelope().toString())
            .put(
                "MessageAttributes",
                JSONObject().put("x-github-event", JSONObject().put("DataType", "String").put("StringValue", "workflow_run")),
            )
        val request = Request.Builder()
            .url(SQS_ENDPOINT)
            .post(message.toString().toRequestBody(amzJson))
            .header("X-Amz-Target", "AmazonSQS.SendMessage")
            .header("Authorization", AUTHORIZATION + "0".repeat(64))
            .build()
        val (status, body) = http.newCall(request).execute().use { it.code to it.body?.string().orEmpty() }
        assertEquals("SendMessage for ${event.workflow} should be accepted: $body", 200, status)
    }

    fun history(api: BackendApi): List<JSONObject> {
        val all = mutableListOf<JSONObject>()
        var before: Long? = null
        repeat(20) {
            val reply = api.call("GET", "/api/notification-history?limit=100" + (before?.let { "&before=$it" } ?: ""))
            assertEquals("The history list should load: ${reply.body}", 200, reply.status)
            val page = reply.items()
            all += page
            if (page.size < 100) return all
            before = page.last().getLong("id")
        }
        return all
    }

    fun items(api: BackendApi, repo: String): List<JSONObject> = history(api).filter { it.optString("repo_full_name") == repo }

    fun awaitItems(api: BackendApi, repo: String, count: Int): List<JSONObject> {
        var found = items(api, repo)
        val deadline = System.currentTimeMillis() + 60_000
        while (found.size < count && System.currentTimeMillis() < deadline) {
            Thread.sleep(1_000)
            found = items(api, repo)
        }
        assertEquals("The consumer should store $count history items for $repo", count, found.size)
        return found
    }
}

object PrMonitor {
    const val GROUP_PREFIX = "prmonitor.group."
    private val rowAck = Regex("prmonitor\\.row\\.\\d+\\.ack\\.button")
    private val rowStatus = Regex("prmonitor\\.row\\.\\d+\\.status")

    fun groupTag(key: String) = GROUP_PREFIX + key

    fun open(compose: ComposeTestRule): ActivityScenario<MainActivity> {
        val scenario = AppSession.launchSignedIn(compose)
        compose.openTab("PrMonitor")
        compose.waitForTag("prmonitor.screen")
        return scenario
    }

    fun reopenTab(compose: ComposeTestRule) {
        compose.openTab("Affirmations")
        compose.waitForTagGone("prmonitor.screen")
        compose.openTab("PrMonitor")
        compose.waitForTag("prmonitor.screen")
    }

    fun openUntil(compose: ComposeTestRule, group: String, label: String): ActivityScenario<MainActivity> {
        val scenario = open(compose)
        var attempts = 0
        while (!waitForLabel(compose, group, label, 10_000) && attempts < 6) {
            attempts += 1
            reopenTab(compose)
        }
        assertTrue("Group $group should read $label: ${labels(compose, group)}", labels(compose, group).contains(label))
        return scenario
    }

    fun waitForLabel(compose: ComposeTestRule, group: String, label: String, timeoutMillis: Long = 10_000): Boolean =
        E2e.eventually(timeoutMillis) { reveal(compose, group) && labels(compose, group).contains(label) }

    fun reveal(compose: ComposeTestRule, group: String): Boolean {
        val tag = groupTag(group)
        scrollIntoView(compose, tag)
        return compose.exists(tag)
    }

    fun scrollIntoView(compose: ComposeTestRule, tag: String) {
        val list = compose.onAllNodes(hasScrollToNodeAction() and hasAnyAncestor(hasTestTag("prmonitor.screen")))
        if (list.fetchSemanticsNodes().isEmpty()) return
        runCatching { list[0].performScrollToNode(hasTestTag(tag)) }
        compose.waitForIdle()
    }

    fun scrollToTop(compose: ComposeTestRule) {
        val list = compose.onAllNodes(hasScrollToNodeAction() and hasAnyAncestor(hasTestTag("prmonitor.screen")))
        if (list.fetchSemanticsNodes().isNotEmpty()) runCatching { list[0].performScrollToIndex(0) }
        compose.waitForIdle()
    }

    fun labels(compose: ComposeTestRule, group: String): List<String> = compose.textsUnder(groupTag(group))

    fun hasLabel(compose: ComposeTestRule, group: String, label: String): Boolean = labels(compose, group).contains(label)

    fun itemAcknowledgeButtons(compose: ComposeTestRule, group: String): List<String> =
        compose.tagsUnder(groupTag(group)).filter(rowAck::matches)

    fun acknowledgeAllButtons(compose: ComposeTestRule, group: String): Int =
        if (compose.exists(groupTag(group) + ".ack-all.button")) 1 else 0

    fun itemStatuses(compose: ComposeTestRule, group: String): List<String> =
        compose.tagsUnder(groupTag(group)).filter(rowStatus::matches).map { compose.textOf(it) }

    fun acknowledgedItems(compose: ComposeTestRule, group: String): Int = itemStatuses(compose, group).count { it.endsWith("· 확인됨") }

    fun badge(compose: ComposeTestRule, group: String): String =
        if (compose.exists(groupTag(group) + ".badge")) compose.textOf(groupTag(group) + ".badge") else ""

    fun tapFirstItemAcknowledge(compose: ComposeTestRule, group: String) {
        val button = itemAcknowledgeButtons(compose, group).firstOrNull()
        assertTrue("Group $group should offer 확인 on an item", button != null)
        scrollIntoView(compose, button!!)
        compose.tap(button)
    }

    fun tapAcknowledgeAll(compose: ComposeTestRule, group: String) {
        val tag = groupTag(group) + ".ack-all.button"
        scrollIntoView(compose, tag)
        compose.tap(tag)
    }

    fun groupsOf(compose: ComposeTestRule, repo: String): Set<String> {
        val pattern = Regex("^" + Regex.escape(GROUP_PREFIX) + "((?:pr|sha):" + Regex.escape(repo) + ":[0-9a-z]+)$")
        return compose.onAllNodes(hasTagMatching(pattern), useUnmergedTree = true).fetchSemanticsNodes()
            .mapNotNull { node -> node.config.getOrNull(SemanticsProperties.TestTag) }
            .mapNotNull { pattern.find(it)?.groupValues?.get(1) }
            .toSet()
    }

    fun top(compose: ComposeTestRule, tag: String): Float =
        compose.onAllNodesWithTagUnmerged(tag).firstOrNull()?.boundsInRoot?.top ?: Float.POSITIVE_INFINITY

    fun unreadCount(compose: ComposeTestRule): Int? {
        if (!E2e.eventually(10_000) { compose.exists("prmonitor.unread.badge") }) return null
        return compose.textsUnder("prmonitor.unread.badge").firstNotNullOfOrNull { it.trim().toIntOrNull() }
    }

    fun waitForUnread(compose: ComposeTestRule, expected: Int): Int? {
        var value = unreadCount(compose)
        val deadline = System.currentTimeMillis() + 15_000
        while (value != expected && System.currentTimeMillis() < deadline) {
            Thread.sleep(1_000)
            value = unreadCount(compose)
        }
        return value
    }
}

fun ComposeTestRule.onAllNodesWithTagUnmerged(tag: String): List<SemanticsNode> =
    onAllNodes(hasTestTag(tag), useUnmergedTree = true).fetchSemanticsNodes()
