package com.dlddu.pocketaide.auth

import android.content.Context
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone

@Serializable
data class SignOutReason(val detail: String, val atMillis: Long) {
    fun message(timeZone: TimeZone = TimeZone.getDefault()): String {
        val format = SimpleDateFormat("M월 d일 a h:mm", Locale.KOREAN).apply { this.timeZone = timeZone }
        return "${format.format(Date(atMillis))}\n\n$detail"
    }

    companion object {
        private const val MAX_DETAIL_LENGTH = 200
        private val json = Json { ignoreUnknownKeys = true }

        fun refreshHttpFailure(status: Int, body: String, atMillis: Long): SignOutReason {
            val verdict = if (status in 400..499) "토큰 갱신 거절" else "토큰 갱신 실패"
            return SignOutReason("$verdict — HTTP $status${suffix(oauthError(body))}", atMillis)
        }

        fun refreshFailure(message: String, atMillis: Long): SignOutReason =
            SignOutReason("토큰 갱신 실패 — ${oneLine(message)}", atMillis)

        fun unauthorized(stored: SignOutReason?, request: String, apiBody: String, atMillis: Long): SignOutReason {
            val api = "API 응답 — $request → HTTP 401${suffix(oneLine(apiBody))}"
            return SignOutReason(listOfNotNull(stored?.detail, api).joinToString("\n"), atMillis)
        }

        fun tokenUnreadable(error: Throwable, atMillis: Long): SignOutReason =
            SignOutReason(
                "저장된 로그인 토큰을 읽지 못해 지웠어요 — ${oneLine("${error::class.simpleName}: ${error.message}")}",
                atMillis,
            )

        fun tokenMissing(atMillis: Long): SignOutReason =
            SignOutReason("앱을 시작할 때 저장된 로그인 토큰이 없었어요. 지워진 이유는 기록되지 않았어요.", atMillis)

        private fun oauthError(body: String): String {
            val parsed = runCatching {
                val fields = json.parseToJsonElement(body).jsonObject
                val error = fields["error"]?.jsonPrimitive?.contentOrNull
                val description = fields["error_description"]?.jsonPrimitive?.contentOrNull
                when {
                    error == null -> null
                    description.isNullOrBlank() -> error
                    else -> "$error: $description"
                }
            }.getOrNull()
            return oneLine(parsed ?: body)
        }

        private fun suffix(text: String): String = if (text.isEmpty()) "" else ": $text"

        private fun oneLine(text: String): String {
            val flat = text.trim().replace(Regex("\\s+"), " ")
            return if (flat.length <= MAX_DETAIL_LENGTH) flat else flat.take(MAX_DETAIL_LENGTH) + "…"
        }
    }
}

interface SignOutReasonStore {
    fun load(): SignOutReason?
    fun save(reason: SignOutReason)
    fun clear()
    fun hadSession(): Boolean
    fun markSession()
    fun forgetSession()
}

class PrefsSignOutReasonStore(context: Context) : SignOutReasonStore {
    private val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)

    override fun load(): SignOutReason? {
        val stored = prefs.getString(KEY_REASON, null) ?: return null
        return runCatching { json.decodeFromString(SignOutReason.serializer(), stored) }.getOrNull()
    }

    override fun save(reason: SignOutReason) {
        prefs.edit().putString(KEY_REASON, json.encodeToString(SignOutReason.serializer(), reason)).commit()
    }

    override fun clear() {
        prefs.edit().remove(KEY_REASON).commit()
    }

    override fun hadSession(): Boolean = prefs.getBoolean(KEY_HAD_SESSION, false)

    override fun markSession() {
        if (!hadSession()) prefs.edit().putBoolean(KEY_HAD_SESSION, true).commit()
    }

    override fun forgetSession() {
        prefs.edit().remove(KEY_HAD_SESSION).commit()
    }

    private companion object {
        const val PREFS = "pocketaide.auth"
        const val KEY_REASON = "signOutReason"
        const val KEY_HAD_SESSION = "hadSession"
        val json = Json { ignoreUnknownKeys = true }
    }
}

class InMemorySignOutReasonStore(initial: SignOutReason? = null, hadSession: Boolean = false) : SignOutReasonStore {
    @Volatile
    private var reason: SignOutReason? = initial

    @Volatile
    private var session: Boolean = hadSession

    override fun load(): SignOutReason? = reason

    override fun save(reason: SignOutReason) {
        this.reason = reason
    }

    override fun clear() {
        reason = null
    }

    override fun hadSession(): Boolean = session

    override fun markSession() {
        session = true
    }

    override fun forgetSession() {
        session = false
    }
}
