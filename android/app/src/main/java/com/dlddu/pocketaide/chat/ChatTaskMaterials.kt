package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * What a person needs in order to review one task, in one small answer
 * (PRD-AI채팅 AC10, 검토 자료).
 *
 * A task in review is opened to decide something, and that takes the pull
 * requests the work produced, what has already been settled, and what is still
 * being asked. All three live in the task's turns, and the turns are the long
 * part — which is why the detail pane keeps them closed (AC10). So the backend
 * folds them once per task and answers with the few short facts, and this is
 * that answer. Nothing here reads the exchanges, and nothing in the task list
 * does either.
 *
 * The agreed turn has no column for "this was decided" and none for an
 * alternative's advantages and drawbacks, and that schema is not the app's to
 * extend. So the backend reads those two out of what the author *marked* in the
 * turn's own `response` or the parent's `explanation`, and nowhere else —
 * [tradeoffsRecorded] and [decisionsRecorded] say whether anything was found.
 * The app manufactures neither: a parent's `explanation` is its account of a
 * result, not a decision, and a choice's label is the words on a button, not a
 * trade-off. A request that repeats an offered sentence is not a press and not
 * an acceptance.
 */
@Serializable
data class ChatTaskMaterials(
    @SerialName("task_id") val taskId: String = "",
    // The platform session these were read from. An answer that names another
    // session belongs to a conversation that has since moved, and is dropped
    // rather than shown under the task now open.
    @SerialName("session_id") val sessionId: String = "",
    // What the work produced, newest sighting first. One entry per distinct
    // thing, however many turns named it.
    val artifacts: List<ChatTaskArtifact> = emptyList(),
    // The alternatives an author weighed, each with the advantages and
    // drawbacks they wrote for it, newest statement first.
    val tradeoffs: List<ChatTaskTradeoff> = emptyList(),
    // What an author recorded as settled, in their own words.
    val decisions: List<ChatTaskDecision> = emptyList(),
    // What the task's latest turn is still asking. An earlier question with
    // turns after it is history, and is deliberately not here.
    val open: List<ChatTaskOpenPoint> = emptyList(),
    // Whether any alternative's advantages and drawbacks were found written
    // down. Read rather than assumed: it is what lets the screen say *why* it
    // shows none instead of leaving a silence that reads as "there were none".
    @SerialName("tradeoffs_recorded") val tradeoffsRecorded: Boolean = false,
    // The same for settled points, for the same reason.
    @SerialName("decisions_recorded") val decisionsRecorded: Boolean = false,
    // The newest turn these materials account for.
    @SerialName("folded_through") val foldedThrough: Long = 0,
    // The fold stopped at its page bound, so this is a prefix. Saying so keeps
    // a partial gathering from reading as "this is every pull request there is".
    val truncated: Boolean = false,
) {
    /** Whether the records gave this task anything at all to review. */
    val hasAny: Boolean get() =
        artifacts.isNotEmpty() || tradeoffs.isNotEmpty() || decisions.isNotEmpty() || open.isNotEmpty()
}

/**
 * One thing the work produced, as the records name it.
 *
 * [ref] is the recorded string verbatim and is always there — it is the fact.
 * [url] is set only when the ref is itself an absolute address, so a row
 * without one is shown as the identifier it is rather than as a dead tap. [repo]
 * and [number] are read out of the ref itself, and [title] only ever comes from
 * a record that already ties that exact repository and number to a title.
 */
@Serializable
data class ChatTaskArtifact(
    val ref: String = "",
    val url: String = "",
    val title: String = "",
    val repo: String = "",
    val number: Int = 0,
    @SerialName("turn_id") val turnId: Long = 0,
) {
    /** Whether a tap has somewhere to go. */
    val openable: Boolean get() = url.isNotBlank()

    /**
     * The line to show. The recorded title when there is one, otherwise the
     * identifier — `repo#number` when the ref says both, and the ref itself
     * when it does not. A title is never invented from the number.
     */
    val label: String get() = when {
        title.isNotBlank() -> title
        repo.isNotBlank() && number > 0 -> "$repo#$number"
        else -> ref
    }

    /** Whether [label] is the recorded title rather than an identifier. */
    val titled: Boolean get() = title.isNotBlank()
}

/**
 * One settled point, in the author's own words.
 *
 * [text] is verbatim — nothing summarises it. [turnId] and [source] are where
 * it was written, so a reader can open that exchange and read the whole thing
 * rather than taking this line's word for it.
 */
@Serializable
data class ChatTaskDecision(
    val text: String = "",
    @SerialName("turn_id") val turnId: Long = 0,
    val source: String = "",
)

/**
 * One alternative as its author weighed it.
 *
 * [pros] and [cons] are their own words. An empty one means that side was not
 * written, and the screen says so for that side rather than leaving a blank
 * that reads as "no drawbacks". [alternative] is the name the author gave the
 * option, which is what makes two of these comparable.
 */
@Serializable
data class ChatTaskTradeoff(
    val alternative: String = "",
    val pros: List<String> = emptyList(),
    val cons: List<String> = emptyList(),
    @SerialName("turn_id") val turnId: Long = 0,
    val source: String = "",
) {
    /** Whether the author wrote both sides of this one. */
    val bothSides: Boolean get() = pros.isNotEmpty() && cons.isNotEmpty()
}

/**
 * Something the task's latest turn is still asking: the question, and the
 * sentences offered as answers. Each sentence is a sentence to send, not a
 * permission.
 */
@Serializable
data class ChatTaskOpenPoint(
    @SerialName("turn_id") val turnId: Long = 0,
    val question: String = "",
    val choices: List<ChatTaskChoice> = emptyList(),
)

/**
 * Why the review materials could not be read. The same four answers the
 * exchanges have, because the platform's refusals mean the same things — and
 * none of them is "this task produced nothing", which would read as a finished
 * review.
 */
enum class ChatTaskMaterialsError(val message: String) {
    NOT_FOUND("이 작업이 목록에서 사라졌어요."),
    UNAVAILABLE("채팅 세션이 깨어 있지 않아 검토 자료를 읽지 못했어요."),
    UNSUPPORTED("이 대화의 세션은 작업 기록을 두지 않아요."),
    UNREACHABLE("검토 자료를 불러오지 못했어요. 다시 시도해요."),
    ;

    companion object {
        fun from(e: Exception): ChatTaskMaterialsError = when (e) {
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

/** The review materials for one task. One call, and only when a task is open. */
suspend fun ApiClient.chatTaskMaterials(taskId: String): ChatTaskMaterials =
    get("/api/chat/tasks/$taskId/materials", ChatTaskMaterials.serializer())
