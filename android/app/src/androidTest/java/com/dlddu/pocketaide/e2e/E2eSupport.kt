// mock-exception: EXT — 실 OIDC IdP 대신 oidcmock 과 로그인 왕복(앱은 Custom Tab, 테스트 러너는 PKCE)을 하고, 그 토큰으로 실 백엔드 API 에 데이터를 시드한다 (docs/e2e-mocking-policy.md 허용목록)
package com.dlddu.pocketaide.e2e

import android.Manifest
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.util.Base64
import android.util.Log
import androidx.compose.ui.semantics.SemanticsNode
import androidx.compose.ui.semantics.SemanticsProperties
import androidx.compose.ui.semantics.getOrNull
import androidx.compose.ui.state.ToggleableState
import androidx.compose.ui.test.SemanticsMatcher
import androidx.compose.ui.test.SemanticsNodeInteraction
import androidx.compose.ui.test.hasScrollToNodeAction
import androidx.compose.ui.test.hasTestTag
import androidx.compose.ui.test.hasText
import androidx.compose.ui.test.junit4.ComposeTestRule
import androidx.compose.ui.test.junit4.createEmptyComposeRule
import androidx.compose.ui.test.onAllNodesWithTag
import androidx.compose.ui.test.onNodeWithTag
import androidx.compose.ui.test.performClick
import androidx.compose.ui.test.performScrollToNode
import androidx.compose.ui.test.performTextInput
import androidx.test.core.app.ActivityScenario
import androidx.test.platform.app.InstrumentationRegistry
import androidx.test.rule.GrantPermissionRule
import androidx.test.runner.lifecycle.ActivityLifecycleMonitorRegistry
import androidx.test.runner.lifecycle.Stage
import androidx.test.uiautomator.By
import androidx.test.uiautomator.UiDevice
import com.dlddu.pocketaide.MainActivity
import okhttp3.FormBody
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Rule
import java.security.MessageDigest
import java.security.SecureRandom
import java.util.UUID
import java.util.concurrent.TimeUnit
import java.util.regex.Pattern

abstract class E2eTest {
    @get:Rule(order = 0)
    val permissions: GrantPermissionRule = GrantPermissionRule.grant(*E2e.startupPermissions)

    @get:Rule(order = 1)
    val compose: ComposeTestRule = createEmptyComposeRule()

    protected fun launchSignedIn(): ActivityScenario<MainActivity> = AppSession.launchSignedIn(compose)

    protected fun relaunch(scenario: ActivityScenario<MainActivity>): ActivityScenario<MainActivity> {
        scenario.close()
        return AppSession.launchSignedIn(compose)
    }

    @After
    fun closeApp() {
        AppSession.closeAll()
    }
}

object E2e {
    val startupPermissions: Array<String> = buildList {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) add(Manifest.permission.POST_NOTIFICATIONS)
        add(Manifest.permission.READ_CALENDAR)
        add(Manifest.permission.ACCESS_COARSE_LOCATION)
    }.toTypedArray()

    fun uniqueToken(): String = UUID.randomUUID().toString().take(8)

    fun eventually(timeoutMillis: Long = 15_000, condition: () -> Boolean): Boolean {
        val deadline = System.currentTimeMillis() + timeoutMillis
        while (System.currentTimeMillis() < deadline) {
            if (runCatching(condition).getOrDefault(false)) return true
            Thread.sleep(250)
        }
        return runCatching(condition).getOrDefault(false)
    }
}

object AppSession {
    private const val TAG = "PocketAideE2e"
    private const val SIGN_IN_TIMEOUT_MILLIS = 120_000L
    private const val SIGN_IN_LABEL = "Sign in with OIDC"
    private const val CANCELLED_AFTER_MILLIS = 3_000L
    private const val MAX_SIGN_IN_ATTEMPTS = 3
    private const val CHROME = "com.android.chrome"
    private const val WARM_UP_URL = "${BackendApi.IDP}/.well-known/openid-configuration"
    private const val WARM_UP_MARKER = "authorization_endpoint"
    private const val WARM_UP_TIMEOUT_MILLIS = 60_000L
    private val browserButtons = Pattern.compile(
        "(?i)(accept & continue|accept and continue|use without an account|no thanks|no, thanks|continue|got it|skip|not now|open|ok)",
    )
    private val browserButtonIds = listOf(
        "com.android.chrome:id/terms_accept",
        "com.android.chrome:id/signin_fre_dismiss_button",
        "com.android.chrome:id/negative_button",
        "com.android.chrome:id/positive_button",
        "com.android.chrome:id/button_primary",
        "com.android.chrome:id/ack_button",
    )

    private val instrumentation get() = InstrumentationRegistry.getInstrumentation()
    private val device: UiDevice get() = UiDevice.getInstance(instrumentation)
    private val appPackage: String get() = instrumentation.targetContext.packageName

    private var browserWarmedUp = false

    fun launchSignedIn(compose: ComposeTestRule): ActivityScenario<MainActivity> {
        warmUpBrowser()
        var scenario = ActivityScenario.launch(MainActivity::class.java)
        compose.waitUntil(30_000) { compose.exists(Tags.SHELL) || compose.exists(Tags.SIGN_IN) }
        val signedInBefore = compose.exists(Tags.SHELL)
        var attempts = 0
        while (!compose.exists(Tags.SHELL)) {
            attempts += 1
            if (attempts > MAX_SIGN_IN_ATTEMPTS) {
                throw AssertionError("Sign-in kept returning to the login screen: texts=${visibleTexts()}")
            }
            compose.waitUntil(15_000) { compose.isEnabled(Tags.SIGN_IN) }
            compose.onNodeWithTag(Tags.SIGN_IN).performClick()
            if (passBrowser()) break
            Log.i(TAG, "sign-in returned to the login screen, retrying ($attempts)")
        }
        compose.waitUntil(30_000) { compose.exists(Tags.SHELL) }
        if (!signedInBefore) {
            finishAppActivities()
            scenario = ActivityScenario.launch(MainActivity::class.java)
            compose.waitUntil(30_000) { compose.exists(Tags.SHELL) }
        }
        return scenario
    }

    fun closeAll() {
        finishAppActivities()
    }

    private fun finishAppActivities() {
        val monitor = ActivityLifecycleMonitorRegistry.getInstance()
        val live = listOf(Stage.PRE_ON_CREATE, Stage.CREATED, Stage.STARTED, Stage.RESUMED, Stage.PAUSED, Stage.STOPPED)
        fun appActivities() = live.flatMap { monitor.getActivitiesInStage(it) }.filterIsInstance<MainActivity>()
        instrumentation.runOnMainSync { appActivities().forEach { it.finish() } }
        val gone = E2e.eventually(10_000) {
            var empty = false
            instrumentation.runOnMainSync { empty = appActivities().isEmpty() }
            empty
        }
        Log.i(TAG, "closed the sign-in activities before relaunching: allGone=$gone")
    }

    private fun warmUpBrowser() {
        if (browserWarmedUp) return
        browserWarmedUp = true
        val context = instrumentation.targetContext
        val intent = Intent(Intent.ACTION_VIEW, Uri.parse(WARM_UP_URL))
            .setPackage(CHROME)
            .addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        if (intent.resolveActivity(context.packageManager) == null) return
        context.startActivity(intent)
        val deadline = System.currentTimeMillis() + WARM_UP_TIMEOUT_MILLIS
        while (System.currentTimeMillis() < deadline) {
            val foreground = device.currentPackageName
            if (foreground == CHROME && device.hasObject(By.textContains(WARM_UP_MARKER))) break
            if (foreground != null && foreground != appPackage) dismissBrowserPrompt(foreground)
            device.waitForIdle(1_000)
            Thread.sleep(500)
        }
        Log.i(TAG, "browser warm-up finished: foreground=${device.currentPackageName}")
        device.pressHome()
    }

    private fun passBrowser(): Boolean {
        val deadline = System.currentTimeMillis() + SIGN_IN_TIMEOUT_MILLIS
        var backOnLoginSince = 0L
        while (System.currentTimeMillis() < deadline) {
            if (device.hasObject(By.pkg(appPackage).text(Tags.SHELL_LABEL))) return true
            val foreground = device.currentPackageName
            if (foreground != appPackage) {
                backOnLoginSince = 0L
                dismissBrowserPrompt(foreground)
            } else if (!device.hasObject(By.pkg(appPackage).text(SIGN_IN_LABEL))) {
                backOnLoginSince = 0L
            } else if (backOnLoginSince == 0L) {
                backOnLoginSince = System.currentTimeMillis()
            } else if (System.currentTimeMillis() - backOnLoginSince > CANCELLED_AFTER_MILLIS) {
                return false
            }
            device.waitForIdle(1_000)
            Thread.sleep(500)
        }
        throw AssertionError("Sign-in did not return to the app: foreground=${device.currentPackageName} texts=${visibleTexts()}")
    }

    private fun dismissBrowserPrompt(foreground: String?) {
        foreground ?: return
        browserButtonIds.firstNotNullOfOrNull { device.findObject(By.res(it)) }?.let {
            Log.i(TAG, "tapping browser control ${it.resourceName}")
            it.click()
            return
        }
        device.findObject(By.pkg(foreground).text(browserButtons).clickable(true))?.let {
            Log.i(TAG, "tapping browser button ${it.text}")
            it.click()
        }
    }

    private fun visibleTexts(): List<String> =
        device.findObjects(By.text(Pattern.compile(".+"))).mapNotNull { runCatching { it.text }.getOrNull() }.take(40)
}

object Tags {
    const val SHELL = "tab.Affirmations"
    const val SHELL_LABEL = "더 보기"
    const val SIGN_IN = "SignInButton"
}

fun ComposeTestRule.exists(tag: String): Boolean =
    onAllNodesWithTag(tag, useUnmergedTree = true).fetchSemanticsNodes().isNotEmpty()

fun ComposeTestRule.isEnabled(tag: String): Boolean =
    onAllNodesWithTag(tag, useUnmergedTree = true).fetchSemanticsNodes()
        .any { it.config.getOrNull(SemanticsProperties.Disabled) == null }

fun ComposeTestRule.node(tag: String): SemanticsNodeInteraction = onNodeWithTag(tag, useUnmergedTree = true)

fun ComposeTestRule.waitForTag(tag: String, timeoutMillis: Long = 15_000): SemanticsNodeInteraction {
    waitUntil(timeoutMillis) { exists(tag) }
    return node(tag)
}

fun ComposeTestRule.waitForTagGone(tag: String, timeoutMillis: Long = 15_000) {
    waitUntil(timeoutMillis) { !exists(tag) }
}

fun ComposeTestRule.tap(tag: String, timeoutMillis: Long = 15_000) {
    waitForTag(tag, timeoutMillis).performClick()
}

fun ComposeTestRule.type(tag: String, text: String) {
    waitForTag(tag).performTextInput(text)
}

fun ComposeTestRule.textOf(tag: String): String =
    node(tag).fetchSemanticsNode().config.getOrNull(SemanticsProperties.Text)?.joinToString("") { it.text } ?: ""

fun ComposeTestRule.waitForText(tag: String, expected: String, timeoutMillis: Long = 15_000) {
    val matched = runCatching { waitUntil(timeoutMillis) { exists(tag) && textOf(tag) == expected } }.isSuccess
    assertTrue("$tag should read '$expected' but read '${if (exists(tag)) textOf(tag) else "<missing>"}'", matched)
}

fun ComposeTestRule.mergedText(tag: String): String {
    val node = onAllNodesWithTag(tag).fetchSemanticsNodes().firstOrNull() ?: return ""
    val parts = node.config.getOrNull(SemanticsProperties.Text).orEmpty().map { it.text } +
        node.config.getOrNull(SemanticsProperties.ContentDescription).orEmpty()
    return parts.joinToString(" ")
}

fun ComposeTestRule.isSelected(tag: String): Boolean =
    onAllNodesWithTag(tag, useUnmergedTree = true).fetchSemanticsNodes()
        .any { it.config.getOrNull(SemanticsProperties.Selected) == true }

fun ComposeTestRule.isOn(tag: String): Boolean =
    onAllNodesWithTag(tag, useUnmergedTree = true).fetchSemanticsNodes()
        .any { it.config.getOrNull(SemanticsProperties.ToggleableState) == ToggleableState.On }

fun ComposeTestRule.subtree(tag: String): List<SemanticsNode> {
    val root = onAllNodesWithTag(tag, useUnmergedTree = true).fetchSemanticsNodes().firstOrNull() ?: return emptyList()
    val out = mutableListOf<SemanticsNode>()
    fun walk(node: SemanticsNode) {
        out += node
        node.children.forEach(::walk)
    }
    walk(root)
    return out
}

fun ComposeTestRule.textsUnder(tag: String): List<String> =
    subtree(tag)
        .mapNotNull { node -> node.config.getOrNull(SemanticsProperties.Text)?.let { node.boundsInRoot.top to it.joinToString("") { t -> t.text } } }
        .sortedBy { it.first }
        .map { it.second }

fun ComposeTestRule.tagsUnder(tag: String): List<String> =
    subtree(tag).mapNotNull { it.config.getOrNull(SemanticsProperties.TestTag) }

fun ComposeTestRule.anyTextContains(text: String): Boolean =
    onAllNodes(hasText(text, substring = true), useUnmergedTree = true).fetchSemanticsNodes().isNotEmpty()

fun ComposeTestRule.hasAnyText(text: String): Boolean =
    onAllNodes(hasText(text), useUnmergedTree = true).fetchSemanticsNodes().isNotEmpty()

fun ComposeTestRule.tagWithText(prefix: Regex, text: String): String? =
    onAllNodes(hasText(text) and hasTagMatching(prefix), useUnmergedTree = true).fetchSemanticsNodes()
        .firstNotNullOfOrNull { it.config.getOrNull(SemanticsProperties.TestTag) }

fun ComposeTestRule.scrollTo(tag: String) {
    val lists = onAllNodes(hasScrollToNodeAction()).fetchSemanticsNodes()
    if (lists.isEmpty() || exists(tag)) return
    runCatching { onAllNodes(hasScrollToNodeAction())[0].performScrollToNode(hasTestTag(tag)) }
}

fun ComposeTestRule.openTab(tab: String) {
    tap("tab.$tab")
}

fun ComposeTestRule.openMore(tab: String) {
    tap("tab.More")
    tap("more.$tab")
}

fun hasTagMatching(pattern: Regex): SemanticsMatcher =
    SemanticsMatcher("testTag matches $pattern") { node ->
        node.config.getOrNull(SemanticsProperties.TestTag)?.let(pattern::matches) ?: false
    }

class BackendReply(val status: Int, val body: String, val location: String?) {
    val json: JSONObject? get() = runCatching { JSONObject(body) }.getOrNull()

    fun items(key: String = "items"): List<JSONObject> {
        val array = json?.optJSONArray(key) ?: JSONArray()
        return (0 until array.length()).map { array.getJSONObject(it) }
    }
}

class BackendApi(private val token: String) {
    fun call(method: String, path: String, json: JSONObject? = null): BackendReply =
        perform(method, path, "Bearer $token", json)

    fun todos(area: String): List<JSONObject> = call("GET", "/api/todos/$area").items()

    fun todo(area: String, title: String): JSONObject? = todos(area).firstOrNull { it.optString("title") == title }

    fun scratchpad(): List<JSONObject> = call("GET", "/api/scratchpad").items()

    fun affirmations(): List<JSONObject> = call("GET", "/api/affirmations").items()

    fun approvalStatus(id: String): String? = call("GET", "/api/approvals/$id").json?.optString("status")

    fun issueApprovalKey(name: String): ApprovalCaller {
        val reply = call("POST", "/api/approval-keys", JSONObject().put("name", name))
        assertEquals("Issuing a caller key should succeed: ${reply.body}", 201, reply.status)
        val json = requireNotNull(reply.json)
        return ApprovalCaller(id = json.getLong("id"), key = json.getString("key"))
    }

    companion object {
        const val IDP = "http://localhost:5556"
        const val SERVER = "http://localhost:8080"
        const val CLIENT_ID = "pocket-aide-ios-dev"
        const val REDIRECT_URI = "pocketaide-dev://callback"

        private val http = OkHttpClient.Builder()
            .followRedirects(false)
            .callTimeout(20, TimeUnit.SECONDS)
            .build()
        private val jsonType = "application/json".toMediaType()

        fun signIn(subject: String? = null): BackendApi {
            val verifier = randomVerifier()
            val challenge = Base64.encodeToString(
                MessageDigest.getInstance("SHA-256").digest(verifier.toByteArray()),
                Base64.URL_SAFE or Base64.NO_PADDING or Base64.NO_WRAP,
            )
            val authorize = Uri.parse("$IDP/authorize").buildUpon()
                .appendQueryParameter("response_type", "code")
                .appendQueryParameter("client_id", CLIENT_ID)
                .appendQueryParameter("redirect_uri", REDIRECT_URI)
                .appendQueryParameter("scope", "openid profile email")
                .appendQueryParameter("state", UUID.randomUUID().toString())
                .appendQueryParameter("code_challenge", challenge)
                .appendQueryParameter("code_challenge_method", "S256")
                .apply { if (subject != null) appendQueryParameter("login_hint", subject) }
                .build()
                .toString()
            val redirect = send(Request.Builder().url(authorize).build())
            val code = redirect.location?.let { Uri.parse(it).getQueryParameter("code") }
            assertTrue("/authorize should redirect with a code: ${redirect.status} ${redirect.body}", !code.isNullOrEmpty())
            val form = FormBody.Builder()
                .add("grant_type", "authorization_code")
                .add("code", code!!)
                .add("code_verifier", verifier)
                .add("client_id", CLIENT_ID)
                .add("redirect_uri", REDIRECT_URI)
                .build()
            val reply = send(Request.Builder().url("$IDP/token").post(form).build())
            val token = reply.json?.optString("access_token").orEmpty()
            assertTrue("/token should issue an access token: ${reply.status} ${reply.body}", reply.status == 200 && token.isNotEmpty())
            return BackendApi(token)
        }

        fun setNextAppLoginSubject(subject: String) {
            val form = FormBody.Builder().add("sub", subject).build()
            val reply = send(Request.Builder().url("$IDP/e2e/next-login-subject").post(form).build())
            assertEquals("oidcmock should take the next app login subject: ${reply.body}", 204, reply.status)
        }

        fun perform(method: String, path: String, authorization: String?, json: JSONObject?): BackendReply {
            val body: RequestBody? = json?.toString()?.toRequestBody(jsonType)
                ?: if (method == "POST" || method == "PATCH" || method == "PUT") ByteArray(0).toRequestBody(null) else null
            val request = Request.Builder().url(SERVER + path).method(method, body).apply {
                if (authorization != null) header("Authorization", authorization)
            }.build()
            return send(request)
        }

        private fun send(request: Request): BackendReply =
            http.newCall(request).execute().use {
                BackendReply(it.code, it.body?.string().orEmpty(), it.header("Location"))
            }

        private fun randomVerifier(): String {
            val alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
            val random = SecureRandom()
            return (1..64).map { alphabet[random.nextInt(alphabet.length)] }.joinToString("")
        }
    }
}

class ApprovalCaller(val id: Long, val key: String) {
    fun create(externalId: String, context: String, timeoutSeconds: Int): String {
        val body = JSONObject()
            .put("externalId", externalId)
            .put("context", context)
            .put("requesterName", "Deploy Bot")
            .put("timeoutSeconds", timeoutSeconds)
        val reply = BackendApi.perform("POST", "/api/external/approvals", "Bearer $key", body)
        assertEquals("Creating an approval request should succeed: ${reply.body}", 201, reply.status)
        return requireNotNull(reply.json).getString("id")
    }

    fun view(id: String): JSONObject? = BackendApi.perform("GET", "/api/external/approvals/$id", "Bearer $key", null).json
}
