package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.api.isAuthPoolExhausted
import com.dlddu.pocketaide.sessions.AgentSessionCheckpoint
import com.dlddu.pocketaide.sessions.AgentSessionFormatting
import java.time.Instant
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// PRD-1 AC12: whether the codex session answering this conversation is on, and
// turning it back on when it is not. The backend asks the platform about that
// one session — the one this user's conversation points at — and never reads
// its transcript, so this can be polled without restoring an archived session
// (PRD-11 AC5).

/**
 * What the chat tab says about the session answering the conversation.
 *
 * [UNKNOWN] is a real answer and not a parsing accident: the backend sends it
 * for a platform state it cannot place, and for a turn-on the platform took
 * without a state coming back. An answer this app does not recognise lands
 * here too, which is the same honest "we do not know".
 */
enum class ChatSessionPowerState(val wire: String, val label: String) {
    /** No session yet — the first message opens one (AC4). */
    ABSENT("none", "세션 없음"),

    /** The workload is up: answering, or idling with its environment still there. */
    ON("on", "세션 켜짐"),

    /** Archived — the environment was handed back and a restore brings it back. */
    OFF("off", "세션 꺼짐"),

    /** The platform no longer has the session. The conversation is untouched. */
    GONE("gone", "세션 없음"),

    UNKNOWN("unknown", "세션 상태 알 수 없음"),
    ;

    /**
     * Whether the screen offers to turn the session on. Only an archived
     * session is offered: there is nothing to restore in the others, and a
     * button that would have to invent a session is not this feature.
     */
    val canPowerOn: Boolean get() = this == OFF

    companion object {
        fun from(wire: String): ChatSessionPowerState =
            entries.firstOrNull { it.wire == wire } ?: UNKNOWN
    }
}

@Serializable
data class ChatSessionPower(
    /** The backend's word, kept as it arrived so an unknown one is not lost. */
    val power: String = "",
    /** The platform's own state, when it had one. */
    val state: String? = null,
    @SerialName("last_access") val lastAccess: String? = null,
    /** The archive's own detail, in the sessions tab's model — one shape for
     * one platform field (PRD-11 AC5). */
    val checkpoint: AgentSessionCheckpoint? = null,
    /**
     * True only when this answer came from a turn-on that actually restored the
     * session, so "켰어요" is never said about one that was already on.
     */
    val restored: Boolean = false,
) {
    val powerState: ChatSessionPowerState get() = ChatSessionPowerState.from(power)

    /**
     * The line under the state, in the words the sessions tab uses for the same
     * platform states (PRD-11 AC5) — and, where the platform gave us nothing to
     * show, what happens next instead of a blank.
     */
    fun detail(now: Instant = Instant.now()): String = when (powerState) {
        ChatSessionPowerState.ABSENT -> "첫 메시지를 보내면 세션이 열려요"
        ChatSessionPowerState.ON ->
            if (lastAccess.isNullOrBlank()) {
                "환경이 떠 있어요"
            } else {
                "환경이 떠 있어요 · 마지막 사용 ${AgentSessionFormatting.relative(lastAccess, now)}"
            }
        ChatSessionPowerState.OFF -> buildList {
            checkpoint?.createdAt?.takeIf { it.isNotBlank() }?.let {
                add("아카이브 ${AgentSessionFormatting.relative(it, now)}")
            }
            AgentSessionFormatting.size(checkpoint?.sizeBytes)?.let { add(it) }
            add("환경 반납됨 · 켜면 이 대화가 그대로 이어져요")
        }.joinToString(" · ")
        ChatSessionPowerState.GONE -> "세션이 사라졌어요 · 대화 기록은 그대로이고 다음 메시지가 새 세션을 열어요"
        ChatSessionPowerState.UNKNOWN ->
            if (state.isNullOrBlank()) "상태를 확인하지 못했어요" else "플랫폼이 「$state」라고 알려 왔어요"
    }

    /**
     * What to say after a turn-on, which is not always "turned it on": the
     * session may have been on already, or the platform may have taken the
     * restore without a state coming back.
     */
    val powerOnNotice: String get() = when {
        !restored -> "세션이 이미 켜져 있었어요."
        powerState == ChatSessionPowerState.ON -> "세션을 켰어요. 이어서 보내면 같은 대화로 이어져요."
        else -> "켜기를 접수했어요. 켜졌는지는 아직 확인하지 못했어요 — 상태를 다시 읽어 주세요."
    }
}

/**
 * The refusals the chat tab explains differently. The backend's `code` is what
 * tells 409s apart, so a session the platform is busy with is never shown as
 * "send a message first".
 */
enum class ChatPowerError(val message: String) {
    NO_SESSION("아직 세션이 없어요. 먼저 메시지를 한 번 보내 주세요."),
    GONE("그 세션은 이미 사라졌어요. 대화 기록은 그대로이고, 다음 메시지가 새 세션을 열어요."),
    BUSY("플랫폼이 이 세션을 다루는 중이에요. 잠시 뒤 다시 시도해 주세요."),
    SLOTS_BUSY("로그인 슬롯이 모두 사용 중이에요. 다른 세션이 끝난 뒤 다시 시도해 주세요."),
    UNREACHABLE("세션 플랫폼에 연결하지 못했어요. 잠시 뒤 다시 시도해 주세요."),
    FAILED("세션 상태를 처리하지 못했어요."),
    ;

    companion object {
        fun from(e: Exception): ChatPowerError = when (e) {
            is ApiException.BadStatus -> when {
                e.isAuthPoolExhausted() -> SLOTS_BUSY
                e.status == 404 -> GONE
                e.status == 409 -> if (e.body.contains("chat_power_no_session")) NO_SESSION else BUSY
                e.status == 502 || e.status == 503 || e.status == 504 -> UNREACHABLE
                else -> FAILED
            }
            is ApiException.Transport -> UNREACHABLE
            else -> FAILED
        }
    }
}

/**
 * Reads whether the session answering this conversation is on. Safe to call on
 * a schedule: the backend reads the session's record and not its transcript, so
 * an archived session stays archived (PRD-11 AC5).
 */
suspend fun ApiClient.chatSessionPower(): ChatSessionPower =
    get("/api/chat/session", ChatSessionPower.serializer())

/**
 * Turns that same session back on — the explicit restore, which is the only
 * thing that may restore it (PRD-11 「복원 트리거」). It carries no session id:
 * the backend restores the one answering the signed in user's conversation, so
 * no tap can open a second session or reach somebody else's.
 */
suspend fun ApiClient.powerOnChatSession(): ChatSessionPower =
    post(
        "/api/chat/session/power",
        ChatEmptyPayload(),
        ChatEmptyPayload.serializer(),
        ChatSessionPower.serializer(),
    )
