package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull

// The chat tab's wide layout shows the work records of the codex session that
// answers this conversation. The backend reads them straight from the session's
// own SQLite through a read-only platform route — no model summarises them, and
// nothing here writes.
//
// These used to be the ledger's shapes. The ledger is gone (platform AC-N10)
// and the shapes went with it: a stage, a free-text reason and a next action
// were the parent's own bookkeeping rather than anything a person reads, and
// the ledger page had nowhere to put a task's description, its completion
// criteria, or the turn a choice belongs to. What the panel shows now comes
// from the agreed task and turn records.

@Serializable
data class ChatTask(
    val id: String,
    val title: String = "",
    val status: String = ChatTaskStatus.UNKNOWN,
    // What the task is for and how it will be known to be done. The ledger page
    // had no column for either, so the panel could not show what a task was
    // about. Empty is normal and the panel says "not recorded" rather than
    // leaving a blank that reads like an answer.
    val description: String = "",
    @SerialName("completion_criteria") val completionCriteria: String = "",
    // How many exchanges this task has had, so the row can say it has history
    // without being opened.
    @SerialName("turns_total") val turnsTotal: Long = 0,
    // Whether this task was carried over from the retired ledger. The panel
    // uses it to avoid presenting a carried-over task as one written here.
    val migrated: Boolean = false,
    // The open turn on this task, when one is open. Absent when nothing is
    // waiting, so "nothing waiting" and "nothing to say" are not the same
    // answer.
    val waiting: ChatTaskWaiting? = null,
    // What the session offers to say next. Empty means it offers none — the
    // screen says so and keeps free text; the app never makes up a choice.
    val choices: List<ChatTaskChoice> = emptyList(),
    // The turn [choices] were offered on. It goes back with a pick, and the
    // platform refuses one from a turn the task has moved past.
    @SerialName("choice_turn_id") val choiceTurnId: Long = 0,
)

/**
 * The open turn on a task: what was asked, what the session is asking back,
 * and whether anybody has picked it up. [running] is a turn the parent has
 * started and not finished; otherwise nobody has taken it yet.
 *
 * This replaced the ledger's request counts. A person's words are the task's
 * next turn now, so "something is waiting" is "its latest turn has not ended"
 * rather than a tally of rows beside the task — and there is no request
 * lifecycle to poll, because the turn ending is what settles it.
 */
@Serializable
data class ChatTaskWaiting(
    @SerialName("turn_id") val turnId: Long = 0,
    val request: String = "",
    val question: String = "",
    val running: Boolean = false,
)

/**
 * One sentence the session offers for this task. The button shows [label] and
 * fills the request box with [requestText].
 *
 * There is no id and no basis version any more: what makes a pick current is
 * the turn it was offered on still being the task's latest, which the task
 * carries once as [ChatTask.choiceTurnId]. A version beside each choice was a
 * second way of saying the same thing that could disagree with the first.
 *
 * It is a sentence to send, not a permission: picking one approves no tool and
 * no merge, and nothing is sent until the person stores it.
 */
@Serializable
data class ChatTaskChoice(
    val label: String = "",
    @SerialName("request_text") val requestText: String = "",
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
 * What the panel is showing of the ledger. A display choice only: picking a
 * status or hiding completed work changes no ledger row, sends nothing and
 * deletes nothing — it narrows what is drawn and nothing else.
 */
data class ChatTaskFilter(
    // null means every status; the panel's "전체" pill.
    val status: String? = null,
    val hideCompleted: Boolean = false,
) {
    /** The rows this choice draws, in the order the ledger gave them. */
    fun select(tasks: List<ChatTask>): List<ChatTask> = tasks.filter { shows(it) }

    fun shows(task: ChatTask): Boolean {
        if (status != null && task.status != status) return false
        return !(hideCompleted && task.status == ChatTaskStatus.COMPLETED)
    }

    /** Picking the status already picked goes back to 전체, so a pill is its own way out. */
    fun toggled(status: String): ChatTaskFilter = copy(status = if (this.status == status) null else status)

    /**
     * Why there is no row to draw, or null when there is one. A ledger with no
     * tasks, a status that has none and work the user chose to hide are three
     * different answers, and none of them may be shown as another — the last
     * two especially, because the work is still there.
     */
    fun emptiness(tasks: List<ChatTask>): ChatTaskEmptiness? {
        if (tasks.isEmpty()) return ChatTaskEmptiness.NO_TASKS
        if (select(tasks).isNotEmpty()) return null
        // Rows of the chosen status exist, so hiding completed work is what
        // emptied the list; otherwise the session simply has none of this status.
        val ofStatus = tasks.filter { status == null || it.status == status }
        return if (ofStatus.isEmpty()) ChatTaskEmptiness.FILTERED else ChatTaskEmptiness.COMPLETED_HIDDEN
    }
}

/** Which of the three "nothing to show" answers the panel owes the user. */
enum class ChatTaskEmptiness { NO_TASKS, FILTERED, COMPLETED_HIDDEN }

/**
 * The turn a stored request became.
 *
 * It is not a chat message: it is written to the session's own records and
 * stays there until the parent works it, so leaving one never interrupts a
 * reply being written and never puts words in the conversation.
 *
 * There is no request status any more. The ledger kept a row beside the task
 * with a lifecycle of its own (pending, claimed, reconciling, processed,
 * failed) that the app polled; the turn replaced all of it, and whether a
 * request has been dealt with is whether its turn has ended — which the task
 * list already says through [ChatTask.waiting]. Showing a status nothing
 * writes would have been worse than showing none.
 */
@Serializable
data class ChatTaskTurnRef(
    val id: Long = 0,
    val request: String = "",
    @SerialName("ended_at") val endedAt: String? = null,
    val origin: String = "",
) {
    /** Whether the parent has finished with this turn. */
    val settled: Boolean get() = endedAt != null
}

/** What the backend answers when a request is stored. */
@Serializable
data class ChatTaskRequestStored(
    val turn: ChatTaskTurnRef = ChatTaskTurnRef(),
    // True when this id had already stored the same words: the app's retry of
    // an answer it never saw, not a second request.
    val replayed: Boolean = false,
)

/**
 * Why a request could not be stored. Each is a different sentence because the
 * person's next move differs: wait, retry, or nothing to do here. None of them
 * says the request was stored.
 */
enum class ChatTaskRequestError(val message: String) {
    NOT_FOUND("이 작업이 목록에서 사라졌어요. 목록을 다시 불러요."),
    UNAVAILABLE("채팅 세션이 깨어 있지 않아 저장하지 못했어요. 잠시 뒤 다시 시도해요."),
    UNSUPPORTED("이 대화의 세션은 작업 요청을 받지 않아요."),
    CONFLICT("같은 요청 번호로 다른 내용이 이미 저장돼 있어요. 다시 쓰면 새 번호로 저장해요."),
    STALE_CHOICE("작업이 한 번 더 진행돼 고른 선택지가 지금 묻는 것이 아니에요. 쓴 내용은 그대로 두었으니, 새 선택지를 보거나 그대로 다시 저장해요."),
    INVALID("요청을 저장할 수 없는 형식이에요."),
    UNREACHABLE("요청을 저장하지 못했어요. 잠시 뒤 다시 시도해요."),
    ;

    companion object {
        fun from(e: Exception): ChatTaskRequestError = when (e) {
            is ApiException.BadStatus -> when {
                e.status == 404 -> NOT_FOUND
                e.status == 501 -> UNSUPPORTED
                e.status == 409 && e.body.contains("task_request_unavailable") -> UNAVAILABLE
                // The screen is out of date, not the words: its own sentence,
                // and the draft is kept.
                e.status == 409 && e.body.contains("task_request_stale_choice") -> STALE_CHOICE
                e.status == 409 -> CONFLICT
                e.status == 400 -> INVALID
                else -> UNREACHABLE
            }
            else -> UNREACHABLE
        }
    }
}

/**
 * The words that are stored. The free text is the person's own, trimmed and
 * otherwise untouched — the task is named by the route, not by a line pasted
 * into the message, so nothing is sent that the person did not write.
 */
fun chatTaskRequestText(request: String): String = request.trim()

/**
 * Whether a draft still carries the pick it was filled from.
 *
 * Provenance holds only while the draft is the session's sentence word for
 * word. Edit one character and the words are the person's own: the pick and its
 * basis version are dropped and the request is stored as free text, so the
 * session is never told that a sentence it wrote came back when it did not.
 * The same rule decides what is sent, which keeps the two from disagreeing.
 */
fun chatTaskPickHolds(draft: String, pickedText: String): Boolean =
    pickedText.isNotEmpty() && draft == pickedText

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

/**
 * Stores one request on [taskId], where it becomes that task's next turn. [id]
 * is the app's own key for this request, so sending it again after an answer
 * the app never saw stores nothing twice. The backend derives the session and
 * who asked from the signed in user — neither is in this body.
 */
suspend fun ApiClient.storeChatTaskRequest(
    taskId: String,
    id: String,
    text: String,
    choiceTurnId: Long = 0,
): ChatTaskRequestStored =
    post(
        "/api/chat/tasks/$taskId/requests",
        ChatTaskRequestPayload(id = id, text = text, choiceTurnId = choiceTurnId),
        ChatTaskRequestPayload.serializer(),
        ChatTaskRequestStored.serializer(),
    )

@Serializable
private data class ChatTaskRequestPayload(
    val id: String,
    val text: String,
    // The turn whose offered sentences the person picked from. The platform
    // refuses a pick from a turn the task has moved past, so an open screen
    // cannot act on a question that is no longer being asked. Free text sends
    // zero and is never refused for staleness.
    @SerialName("choice_turn_id") val choiceTurnId: Long = 0,
)
