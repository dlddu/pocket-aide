package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class WebhookPushSettingsTest : E2eTest() {
    private class Stage(val name: String, val outcomes: String?, val pushesSuccess: Boolean, val pushesFailure: Boolean) {
        val enabled: Boolean get() = outcomes != null
    }

    private fun event(repo: String, number: Int, conclusion: String, token: String) = WebhookEvent(
        repo = repo,
        pullRequest = number,
        commit = Webhooks.commit(number, token),
        branch = "e2e/settings-$number",
        workflow = "Push settings",
        conclusion = conclusion,
    )

    @Test
    fun pushArrivalFollowsNotificationSettings() {
        val token = Webhooks.runToken()
        val api = BackendApi.signIn()
        restoreSettings(api)
        try {
            val device = Fcm.registerDevice(api)
            run(Stage("failure", "failure", pushesSuccess = false, pushesFailure = true), token, device, api)
            run(Stage("success", "success", pushesSuccess = true, pushesFailure = false), token, device, api)
            run(Stage("off", null, pushesSuccess = false, pushesFailure = false), token, device, api)
        } finally {
            restoreSettings(api)
        }
    }

    private fun run(stage: Stage, token: String, device: String, api: BackendApi) {
        val scenario = PrMonitor.open(compose)
        compose.tap("prmonitor.settings.button")
        compose.waitForTag("prmonitor.settings.enabled")
        compose.waitUntil(10_000) { compose.isEnabled("prmonitor.settings.enabled") }
        if (compose.isOn("prmonitor.settings.enabled") != stage.enabled) compose.tap("prmonitor.settings.enabled")
        compose.waitUntil(10_000) { compose.isOn("prmonitor.settings.enabled") == stage.enabled }
        stage.outcomes?.let { outcomes ->
            if (!compose.isSelected("filter.pill.$outcomes")) compose.tap("filter.pill.$outcomes")
            compose.waitUntil(10_000) { compose.isSelected("filter.pill.$outcomes") }
        }
        waitForSettings(api, stage.enabled, stage.outcomes)
        compose.tap("prmonitor.sheet.done")
        compose.waitForTagGone("prmonitor.settings.enabled")
        scenario.close()

        val repo = "dlddu/push-settings-$token-${stage.name}"
        val passed = event(repo, 91, "success", token)
        val failed = event(repo, 92, "failure", token)
        Webhooks.send(listOf(passed, failed))
        Webhooks.awaitItems(api, repo, 2)

        val expected = buildList {
            if (stage.pushesSuccess) add("CI 통과 — $repo #91")
            if (stage.pushesFailure) add("CI 실패 — $repo #92")
        }
        assertPushes(device, repo, expected)
        assertHistoryShows(listOf(passed, failed))
    }

    private fun restoreSettings(api: BackendApi) {
        api.call("PATCH", "/api/notification-settings", JSONObject().put("enabled", true).put("outcomes", "both"))
    }

    private fun waitForSettings(api: BackendApi, enabled: Boolean, outcomes: String?) {
        var last = ""
        val saved = E2e.eventually(10_000) {
            val reply = api.call("GET", "/api/notification-settings")
            last = reply.body
            val json = reply.json
            json != null && json.optBoolean("enabled") == enabled && (outcomes == null || json.optString("outcomes") == outcomes)
        }
        assertTrue("The sheet should save enabled=$enabled outcomes=${outcomes ?: "-"}, server has $last", saved)
    }

    private fun pushTitles(device: String, repo: String): List<String> =
        Fcm.messages(device).mapNotNull(Fcm::title).filter { it.contains(repo) }.sorted()

    private fun assertPushes(device: String, repo: String, expected: List<String>) {
        E2e.eventually(15_000) { pushTitles(device, repo).size >= expected.size }
        Thread.sleep(3_000)
        assertEquals("Pushes sent for $repo", expected.sorted(), pushTitles(device, repo))
    }

    private fun assertHistoryShows(events: List<WebhookEvent>) {
        val scenario = PrMonitor.open(compose)
        events.forEach {
            assertTrue(
                "The history should list ${it.repo} #${it.pullRequest} whatever the settings",
                E2e.eventually(15_000) { PrMonitor.reveal(compose, it.groupKey) },
            )
        }
        scenario.close()
    }
}
