package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * One task's turns, which is what the detail pane shows as the work itself.
 *
 * Why this exists beside [ChatTaskEvent]: the task row and the history page are
 * the old ledger's shapes, and those have nowhere to put what the agreed model
 * holds — a task's description and completion criteria, and a turn's
 * instruction, the session's own final answer, the parent's explanation, its
 * question and the artifacts. A pane built on them alone can say a task exists
 * and that somebody asked something, but not what was asked, what was done, or
 * what came back.
 *
 * The two are not alternatives. A task carried over from the ledger keeps its
 * ledger history and still shows it, separately and labelled as the older
 * record; a task created after the cutover has no ledger history and none is
 * invented for it — its turns *are* its history.
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
     * `user_request` for one carried over from the old ledger, `app_request`
     * for one a person wrote through this app after the cutover.
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
    /**
     * Whether this task also has an older ledger history to offer under the
     * turns. False for a task created after the cutover, which has none — the
     * pane then does not imply one is missing.
     */
    @SerialName("ledger_history") val ledgerHistory: Boolean = false,
)

/** One page of [taskId]'s turns, oldest first. */
suspend fun ApiClient.chatTaskTurns(taskId: String, afterTurnId: Long, limit: Int): ChatTaskTurnPage {
    val query = buildList {
        if (afterTurnId > 0) add("after_turn_id=$afterTurnId")
        if (limit > 0) add("limit=$limit")
    }.joinToString("&")
    val path = "/api/chat/tasks/$taskId/turns" + if (query.isEmpty()) "" else "?$query"
    return get(path, ChatTaskTurnPage.serializer())
}
