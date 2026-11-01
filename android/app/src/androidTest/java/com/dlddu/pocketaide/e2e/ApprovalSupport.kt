package com.dlddu.pocketaide.e2e

import androidx.compose.ui.test.junit4.ComposeTestRule
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performScrollTo
import androidx.test.core.app.ActivityScenario
import com.dlddu.pocketaide.MainActivity
import org.json.JSONObject
import org.junit.Assert.assertEquals

object AppAccount {
    fun signOut(compose: ComposeTestRule) {
        compose.openTab("PrMonitor")
        compose.tap("prmonitor.settings.button")
        compose.tap("prmonitor.settings.signOut")
        compose.tap("prmonitor.settings.signOut.confirm", timeoutMillis = 10_000)
        compose.waitForTag(Tags.SIGN_IN, timeoutMillis = 20_000)
    }

    fun switchUser(compose: ComposeTestRule, scenario: ActivityScenario<MainActivity>, subject: String): ActivityScenario<MainActivity> {
        signOut(compose)
        scenario.close()
        AppSession.closeAll()
        BackendApi.setNextAppLoginSubject(subject)
        return AppSession.launchSignedIn(compose)
    }

    fun restoreDefaultUser(compose: ComposeTestRule) {
        BackendApi.setNextAppLoginSubject("")
        AppSession.closeAll()
        val signedIn = AppSession.launchSignedIn(compose)
        signOut(compose)
        signedIn.close()
        AppSession.closeAll()
        AppSession.launchSignedIn(compose).close()
        AppSession.closeAll()
    }
}

fun ComposeTestRule.tapScrolled(tag: String, timeoutMillis: Long = 15_000) {
    val node = waitForTag(tag, timeoutMillis)
    runCatching { node.performScrollTo() }
    waitForIdle()
    node(tag).performClick()
}

object ApprovalUi {
    fun openTab(compose: ComposeTestRule) {
        compose.openTab("Approval")
        compose.waitForTag("approvals.title")
    }

    fun openKeys(compose: ComposeTestRule) {
        openTab(compose)
        compose.tap("approvals.settings.button")
        compose.waitForTag("approval.keys.title")
    }

    fun openHistory(compose: ComposeTestRule) {
        compose.tap("approvals.history.button")
        compose.waitForTag("approvals.history.title")
    }

    fun back(compose: ComposeTestRule) {
        compose.tap("approvals.back")
        compose.waitForTag("approvals.title")
    }

    fun refreshPending(compose: ComposeTestRule) {
        compose.openTab("Affirmations")
        compose.waitForTagGone("approvals.title")
        openTab(compose)
    }

    fun waitForRow(compose: ComposeTestRule, externalId: String) {
        val tag = "approvals.row.$externalId"
        if (!E2e.eventually(5_000) { compose.exists(tag) }) refreshPending(compose)
        compose.waitForTag(tag)
    }

    fun issue(compose: ComposeTestRule, name: String) {
        compose.tapScrolled("approval.keys.issue")
        compose.type("approval.keys.issue.name", name)
        compose.tap("approval.keys.issue.submit")
    }

    fun registerServiceAccount(compose: ComposeTestRule, name: String, namespace: String, account: String) {
        compose.tapScrolled("approval.keys.register-sa")
        compose.type("approval.keys.sa.name", name)
        compose.type("approval.keys.sa.namespace", namespace)
        compose.type("approval.keys.sa.account", account)
        compose.tap("approval.keys.sa.submit")
    }

    fun secretIsNowhere(compose: ComposeTestRule, secret: String): Boolean = !compose.anyTextContains(secret)
}

object ApprovalKeysApi {
    fun keys(api: BackendApi): List<JSONObject> {
        val reply = api.call("GET", "/api/approval-keys")
        assertEquals("The caller key list should load: ${reply.body}", 200, reply.status)
        return reply.items("keys")
    }

    fun key(api: BackendApi, name: String): JSONObject? = keys(api).firstOrNull { it.optString("name") == name }

    fun revokeAll(api: BackendApi, name: String) {
        keys(api).filter { it.optString("name") == name }.forEach { api.call("POST", "/api/approval-keys/${it.getLong("id")}/revoke") }
    }

    fun autoResponse(api: BackendApi): String? = api.call("GET", "/api/approval-settings").json?.optString("auto_response")

    fun setAutoResponse(api: BackendApi, mode: String): BackendReply =
        api.call("PATCH", "/api/approval-settings", JSONObject().put("auto_response", mode))
}

fun ApprovalCaller.createReply(externalId: String, context: String, timeoutSeconds: Int? = null): BackendReply {
    val body = JSONObject().put("externalId", externalId).put("context", context).put("requesterName", "Deploy Bot")
    if (timeoutSeconds != null) body.put("timeoutSeconds", timeoutSeconds)
    return BackendApi.perform("POST", "/api/external/approvals", "Bearer $key", body)
}
