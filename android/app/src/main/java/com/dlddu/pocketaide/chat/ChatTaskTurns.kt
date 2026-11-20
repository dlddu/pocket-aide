package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * One task's turns, which is what the detail pane shows — and which is that
 * task's whole history.
 *
 * There used to be a second pane under these, built on the ledger's history
 * route. Both the route and the rows behind it are gone (platform AC-N10): a
 * past management event was never a turn, and dressing one up as a turn would
 * put words in a person's mouth. So the turns are the history, for a task
 * carried over from the ledger exactly as for one written here — the pane has
 * no second source to offer and does not imply one is missing.
 *
 * The turn is also what the old ledger shapes could not hold: a task's
 * description and completion criteria, and a turn's instruction, the session's
 * own final answer, the parent's explanation, its question and the artifacts.
 * A pane built on the ledger page alone could say a task existed and that
 * somebody asked something, but not what was asked, what was done, or what
 * came back.
 */
@Serializable
data class ChatTaskTurn(
    val id: Long = 0,
    /** The person's own words. */
    val request: String? = null,
    /** What the parent passed to the session doing the work. */
    val instruction: String? = null,
    /**
     * That session's final answer, verbatim. Never a summary the parent wrote
     * — that is [explanation]. Showing one as the other would put words in a
     * session's mouth.
     */
    val response: String? = null,
    /** The parent's own account of the result, written for the person. */
    val explanation: String? = null,
    /** What the parent is asking on this turn. */
    val question: String? = null,
    val choices: List<ChatTaskTurnChoice> = emptyList(),
    val artifacts: List<String> = emptyList(),
    @SerialName("requested_at") val requestedAt: String = "",
    @SerialName("started_at") val startedAt: String? = null,
    @SerialName("ended_at") val endedAt: String? = null,
    /**
     * Where the turn came from: empty for one the session wrote,
     * `user_request` for one carried over from the retired ledger, and
     * `app_request` for one a person wrote through this app. The pane uses it
     * to avoid presenting a carried-over exchange as something written here —
     * the ledger's own words were a stored request, never an exchange the
     * session had.
     */
    val origin: String = "",
) {
    /** A turn nobody has finished is still somebody's to answer. */
    val isOpen: Boolean get() = endedAt == null

    /** A turn carried over from the ledger, rather than written here. */
    val isCarriedOver: Boolean get() = origin == ORIGIN_LEDGER

    companion object {
        const val ORIGIN_LEDGER = "user_request"
        const val ORIGIN_APP = "app_request"
    }
}

/**
 * One sentence a turn offers for the next one.
 *
 * There is no id: what a person picks becomes the **next turn's request**, and
 * which turn it was offered on is what decides whether the pick is still
 * current. Tapping one fills the request box — it does not store and does not
 * run anything.
 */
@Serializable
data class ChatTaskTurnChoice(
    val label: String = "",
    @SerialName("request_text") val requestText: String = "",
)

/**
 * The task those turns belong to, in the agreed shape.
 *
 * A null [description] or [completionCriteria] is "not recorded" — which is
 * what a task carried over from a ledger that never kept the field says — and
 * is deliberately not an empty string, so the pane can say nothing is recorded
 * rather than showing a blank as if it were the answer.
 */
@Serializable
data class ChatTaskTurnTask(
    val id: String = "",
    val title: String = "",
    val description: String? = null,
    @SerialName("completion_criteria") val completionCriteria: String? = null,
    val status: String = ChatTaskStatus.UNKNOWN,
    val version: Long = 0,
    @SerialName("session_id") val sessionId: String? = null,
    @SerialName("created_at") val createdAt: String? = null,
    @SerialName("updated_at") val updatedAt: String? = null,
)

/** One page of turns, oldest first, with the task they are about. */
@Serializable
data class ChatTaskTurnPage(
    @SerialName("task_id") val taskId: String = "",
    @SerialName("task_version") val taskVersion: Long = 0,
    val task: ChatTaskTurnTask = ChatTaskTurnTask(),
    val turns: List<ChatTaskTurn> = emptyList(),
    @SerialName("next_cursor") val nextCursor: Long = 0,
)

/**
 * Why a task's turns could not be read. Each is a different sentence because
 * the person's next move differs, and none of them may be shown as "this task
 * has no work" — the turns are still on the slot.
 *
 * This replaced ChatTaskHistoryError, which named the ledger's history route.
 * The classes are the same four because the platform's answers are: a task
 * this slot does not have, a session with no live pod, a deployment that keeps
 * no work records, and anything else.
 */
enum class ChatTaskTurnsError(val message: String) {
    NOT_FOUND("이 작업이 목록에서 사라졌어요."),
    UNAVAILABLE("채팅 세션이 깨어 있지 않아 진행 내용을 읽지 못했어요."),
    UNSUPPORTED("이 대화의 세션은 작업 기록을 두지 않아요."),
    UNREACHABLE("진행 내용을 불러오지 못했어요. 다시 시도해요."),
    ;

    companion object {
        fun from(e: Exception): ChatTaskTurnsError = when (e) {
            is ApiException.BadStatus -> when (e.status) {
                404 -> NOT_FOUND
                409 -> UNAVAILABLE
                501 -> UNSUPPORTED
                else -> UNREACHABLE
            }
            else -> UNREACHABLE
        }
    }
}

/** One page of [taskId]'s turns, oldest first. */
suspend fun ApiClient.chatTaskTurns(taskId: String, afterTurnId: Long, limit: Int): ChatTaskTurnPage {
    val query = buildList {
        if (afterTurnId > 0) add("after_turn_id=$afterTurnId")
        if (limit > 0) add("limit=$limit")
    }.joinToString("&")
    val path = "/api/chat/tasks/$taskId/turns" + if (query.isEmpty()) "" else "?$query"
    return get(path, ChatTaskTurnPage.serializer())
}
