// mock-exception: EXT — FCM 전달(Google 서버 → 기기)은 Firebase 프로젝트 자격증명·실 등록 토큰 경계라 CI 안에서 못 한다; 가짜 FCM 수신부가 기록한 백엔드 메시지를 RemoteMessage 로 앱의 PocketAideMessagingService.onMessageReceived 에 넣고, 알림 표시·탭·액션 버튼 이후는 실경로 (docs/e2e-mocking-policy.md)
package com.dlddu.pocketaide.e2e

import android.app.NotificationManager
import android.content.Context
import android.content.ContextWrapper
import androidx.test.platform.app.InstrumentationRegistry
import androidx.test.uiautomator.By
import androidx.test.uiautomator.BySelector
import androidx.test.uiautomator.UiDevice
import androidx.test.uiautomator.UiObject2
import androidx.test.uiautomator.Until
import com.dlddu.pocketaide.push.PocketAideMessagingService
import com.google.firebase.messaging.RemoteMessage
import okhttp3.OkHttpClient
import okhttp3.Request
import org.json.JSONArray
import org.json.JSONObject
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import java.util.UUID
import java.util.concurrent.TimeUnit

object Fcm {
    const val FIXTURE_DEVICE = "e2e-fcm-fixture"
    private const val RECEIVER = "http://localhost:5559"
    private const val SENDER = "pocket-aide-e2e@fcm.googleapis.com"
    private val http = OkHttpClient.Builder().callTimeout(20, TimeUnit.SECONDS).build()

    fun registerDevice(api: BackendApi): String {
        val device = "e2e-fcm-" + UUID.randomUUID().toString().replace("-", "")
        val reply = api.call("POST", "/api/device-tokens", JSONObject().put("token", device).put("platform", "android"))
        assertEquals("The Android device token should register: ${reply.body}", 201, reply.status)
        return device
    }

    fun messages(device: String): List<JSONObject> {
        val request = Request.Builder().url("$RECEIVER/e2e/received?token=$device").build()
        val body = http.newCall(request).execute().use { it.body?.string().orEmpty() }
        val rows = runCatching { JSONArray(body) }.getOrDefault(JSONArray())
        return (0 until rows.length()).map { rows.getJSONObject(it).getJSONObject("message") }
    }

    fun await(device: String, timeoutMillis: Long = 30_000, match: (JSONObject) -> Boolean): JSONObject? {
        val deadline = System.currentTimeMillis() + timeoutMillis
        do {
            messages(device).firstOrNull(match)?.let { return it }
            Thread.sleep(1_000)
        } while (System.currentTimeMillis() < deadline)
        return null
    }

    fun approval(device: String, approvalId: String, timeoutMillis: Long = 30_000): JSONObject? =
        await(device, timeoutMillis) { data(it)["approval_id"] == approvalId }

    fun titled(device: String, title: String, timeoutMillis: Long = 60_000): JSONObject? =
        await(device, timeoutMillis) { title(it) == title }

    fun data(message: JSONObject): Map<String, String> {
        val data = message.optJSONObject("data") ?: return emptyMap()
        return data.keys().asSequence().associateWith { data.getString(it) }
    }

    fun title(message: JSONObject): String? =
        message.optJSONObject("notification")?.optString("title") ?: data(message)["title"]

    fun body(message: JSONObject): String? =
        message.optJSONObject("notification")?.optString("body") ?: data(message)["body"]

    fun deliver(message: JSONObject) {
        val context = InstrumentationRegistry.getInstrumentation().targetContext
        val builder = RemoteMessage.Builder(SENDER).setMessageId(UUID.randomUUID().toString())
        data(message).forEach { (key, value) -> builder.addData(key, value) }
        message.optJSONObject("notification")?.let { notification ->
            builder.addData("gcm.n.e", "1")
            builder.addData("gcm.n.title", notification.optString("title"))
            builder.addData("gcm.n.body", notification.optString("body"))
        }
        message.optJSONObject("android")?.optJSONObject("notification")?.optString("channel_id")?.let {
            builder.addData("gcm.n.android_channel_id", it)
        }
        val service = PocketAideMessagingService()
        ContextWrapper::class.java.getDeclaredMethod("attachBaseContext", Context::class.java)
            .apply { isAccessible = true }
            .invoke(service, context)
        service.onMessageReceived(builder.build())
    }
}

object Shade {
    private const val SYSTEM_UI = "com.android.systemui"
    private val device: UiDevice get() = UiDevice.getInstance(InstrumentationRegistry.getInstrumentation())

    fun clearAppNotifications() {
        val context = InstrumentationRegistry.getInstrumentation().targetContext
        context.getSystemService(NotificationManager::class.java)?.cancelAll()
    }

    fun find(text: String, timeoutMillis: Long = 20_000): UiObject2? = find(By.pkg(SYSTEM_UI).text(text), timeoutMillis)

    fun findContaining(text: String, timeoutMillis: Long = 20_000): UiObject2? =
        find(By.pkg(SYSTEM_UI).textContains(text), timeoutMillis)

    private fun find(selector: BySelector, timeoutMillis: Long): UiObject2? {
        if (!device.hasObject(By.res(NOTIFICATION_STACK))) device.openNotification()
        if (device.wait(Until.hasObject(selector), timeoutMillis / 2) == true) return device.findObject(selector)
        device.openNotification()
        device.wait(Until.hasObject(selector), timeoutMillis / 2)
        return device.findObject(selector)
    }

    fun tap(title: String) {
        val notification = find(title)
        assertNotNull("The notification '$title' should be in the shade", notification)
        notification!!.click()
        device.waitForIdle()
    }

    fun hasAction(title: String, action: String): Boolean {
        expand(title, action)
        return device.hasObject(By.pkg(SYSTEM_UI).text(action))
    }

    fun tapAction(title: String, action: String) {
        expand(title, action)
        val button = device.findObject(By.pkg(SYSTEM_UI).text(action))
        assertNotNull("The notification '$title' should offer $action", button)
        button!!.click()
        device.waitForIdle()
    }

    fun close() {
        device.executeShellCommand("cmd statusbar collapse")
        device.waitForIdle()
    }

    private fun expand(title: String, action: String) {
        val titleView = find(title)
        assertNotNull("The notification '$title' should be in the shade", titleView)
        if (device.hasObject(By.pkg(SYSTEM_UI).text(action))) return
        var node: UiObject2? = titleView
        repeat(MAX_ANCESTORS) {
            val expander = node?.findObject(By.res(EXPAND_BUTTON))
            if (expander != null) {
                expander.click()
                device.wait(Until.hasObject(By.pkg(SYSTEM_UI).text(action)), 5_000)
                return
            }
            node = node?.parent ?: return
        }
    }

    private const val MAX_ANCESTORS = 10
    private const val NOTIFICATION_STACK = "com.android.systemui:id/notification_stack_scroller"
    private const val EXPAND_BUTTON = "android:id/expand_button"
}
