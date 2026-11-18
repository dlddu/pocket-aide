package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// The chat tab's wide layout shows the task ledger of the codex session that
// answers this conversation. The backend reads it straight from the session's
// own SQLite ledger through a read-only platform route — no model summarises
// it, and nothing here writes.

@Serializable
data class ChatTask(
    val id: String,
    val title: String = "",
    val status: String = ChatTaskStatus.UNKNOWN,
    val stage: String = "",
    // Why the task stands where it does now, and what comes next. The ledger
    // records these only sometimes, so empty is normal and the panel says
    // "not recorded" rather than leaving a blank.
    val reason: String = "",
    @SerialName("next_action") val nextAction: String = "",
)

@Serializable
data class ChatTaskList(
    // The platform session the rows were read from. A reply that names another
    // session than the one asked for is a reply for a conversation that has
    // since moved on, and is dropped rather than shown.
    @SerialName("session_id") val sessionId: String = "",
    val tasks: List<ChatTask> = emptyList(),
    val counts: Map<String, Int> = emptyMap(),
    // The backend could not walk the whole ledger, so this is a prefix. Saying
    // so keeps a partial list from reading as "this is all the work".
    val truncated: Boolean = false,
)

object ChatTaskStatus {
    const val BLOCKED = "blocked"
    const val REVIEW = "review"
    const val RUNNING = "running"
    const val PLANNED = "planned"
    const val COMPLETED = "completed"
    const val UNKNOWN = "unknown"

    /** Display order: what needs a person first, then what waits, then done. */
    val order = listOf(BLOCKED, REVIEW, RUNNING, PLANNED, COMPLETED, UNKNOWN)

    fun label(status: String): String = when (status) {
        BLOCKED -> "차단"
        REVIEW -> "검토"
        RUNNING -> "진행"
        PLANNED -> "예정"
        COMPLETED -> "완료"
        else -> "알 수 없음"
    }

    /** Whether this status is one a person is expected to act on. */
    fun needsAttention(status: String): Boolean = status == BLOCKED || status == REVIEW
}

/**
 * Why the panel cannot show a list. Each is a different sentence on purpose —
 * "no tasks" is an empty list, never one of these.
 */
enum class ChatTaskError(val message: String) {
    UNSUPPORTED("이 대화의 세션은 작업 목록을 두지 않아요."),
    UNAVAILABLE("채팅 세션이 깨어 있지 않아 작업 목록을 읽지 못했어요. 메시지를 보내면 다시 읽어요."),
    UNREACHABLE("작업 목록을 불러오지 못했어요. 잠시 뒤 다시 시도해요."),
    ;

    companion object {
        fun from(e: Exception): ChatTaskError = when (e) {
            is ApiException.BadStatus -> when (e.status) {
                // The backend's own distinction: 501 = this session keeps no
                // ledger, 409 = it keeps one but is not awake to be read.
                501 -> UNSUPPORTED
                409 -> UNAVAILABLE
                else -> UNREACHABLE
            }
            else -> UNREACHABLE
        }
    }
}

suspend fun ApiClient.chatTasks(): ChatTaskList = get("/api/chat/tasks", ChatTaskList.serializer())
