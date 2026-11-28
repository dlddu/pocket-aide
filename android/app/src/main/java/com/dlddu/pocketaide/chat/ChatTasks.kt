package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

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
    // without being opened — and so the detail pane can tell whether the
    // exchanges it already read are still current without reading them again.
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
    // What that same turn said, so the pane can put the words the choices
    // answer directly above them. Null when the session offers no choices, and
    // null from a backend that does not send it yet — the pane then falls back
    // to what it showed before rather than losing the question.
    @SerialName("choice_turn") val choiceTurn: ChatTaskChoiceTurn? = null,
)

/**
 * What the turn that offered [ChatTask.choices] said: the words that were sent,
 * the parent's account of what came back, and the question those choices are
 * answers to.
 *
 * It comes from the same read of the same turn as the choices and
 * [ChatTask.choiceTurnId], so the sentences on screen and the words above them
 * cannot be two different turns. The app never fills a blank one from
 * elsewhere: a field the records do not have is left out, because a turn that
 * only asked is the ordinary case and borrowing another turn's words would put
 * them in this turn's mouth.
 *
 * There is no response here on purpose. The session's verbatim answer is the
 * long part, it is already there when somebody opens the exchanges, and it is
 * not what a person needs in order to read the question in front of them.
 */
@Serializable
data class ChatTaskChoiceTurn(
    val request: String = "",
    val explanation: String = "",
    val question: String = "",
) {
    /** Whether the records gave this turn anything to show above its choices. */
    val hasAny: Boolean get() = request.isNotBlank() || explanation.isNotBlank() || question.isNotBlank()
}

/**
 * The open turn on a task: what was asked, what the session is asking back,
 * and whether anybody has picked it up. [running] is a turn the parent has
 * started and not finished; otherwise nobody has taken it yet.
 *
 * This replaced the ledger's request counts. A person's words are the task's
 * next turn now, so "something is waiting" is "its latest turn has not ended"
 * rather than a tally of rows beside the task — and there is no request
 * lifecycle to poll, because the turn ending is what settles it.
 *
 * [request] is still what the route answers, and so still read here. The pane
 * draws the words from the task's exchanges instead — that turn's own request,
 * read only when somebody opens them — so nothing on screen depends on this
 * copy. It is kept because the field is the route's shape, not the pane's.
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
) {
    /**
     * The same list with every status one this version draws, and the counts
     * added up under those names.
     *
     * The backend folds the ledger's two names for a judgment already, so this
     * is for the answer of a pod that has not rolled yet — and for a newer one
     * that starts counting a status this build has no chip for. Both are the
     * same mistake if left alone: a count under a name nothing is drawn for,
     * which makes the chips disagree with the list and with each other.
     *
     * Counts are remapped rather than recounted from [tasks]: they are the
     * whole ledger's, including the tail a truncated walk never sent, so
     * counting the rows on hand would quietly shrink them. Keys that fold
     * together are summed, so the total still matches what the ledger holds.
     * Every drawn status keeps a key at zero, which is what lets the panel tell
     * "none of this status" from "no such status".
     */
    fun normalised(): ChatTaskList {
        val folded = ChatTaskStatus.order.associateWith { 0 }.toMutableMap()
        counts.forEach { (status, count) ->
            val key = ChatTaskStatus.normalise(status)
            folded[key] = (folded[key] ?: 0) + count
        }
        return copy(
            tasks = tasks.map { it.copy(status = ChatTaskStatus.normalise(it.status)) },
            counts = folded,
        )
    }
}

object ChatTaskStatus {
    const val REVIEW = "review"
    const val RUNNING = "running"
    const val PLANNED = "planned"
    const val COMPLETED = "completed"
    const val UNKNOWN = "unknown"

    /**
     * The ledger's other name for work that is the person's to decide, which is
     * [REVIEW] said twice: a task whose missing prerequisite is an approval, a
     * permission or a decision only the person can give. Records written before
     * the two became one status still hold it, a backend pod that has not rolled
     * yet still sends it, and a phone can still have it saved as the status it
     * was last filtering by — so it is read here and answered as [REVIEW].
     *
     * It is not in [order] and has no label on purpose: there is one chip for
     * one judgment, and a second one would split the same work into two counts.
     */
    private const val LEGACY_BLOCKED = "blocked"

    /** Display order: what needs a person first, then what waits, then done. */
    val order = listOf(REVIEW, RUNNING, PLANNED, COMPLETED, UNKNOWN)

    fun label(status: String): String = when (status) {
        REVIEW -> "검토"
        RUNNING -> "진행"
        PLANNED -> "예정"
        COMPLETED -> "완료"
        else -> "알 수 없음"
    }

    /** Whether this status is one a person is expected to act on. */
    fun needsAttention(status: String): Boolean = status == REVIEW

    /**
     * One task status as this version draws it. The backend already folds the
     * ledger's two names into one, so this is for the answer of a pod that has
     * not rolled yet: without it the same task would be counted under a chip
     * the panel no longer has and so be drawn nowhere.
     *
     * Only [LEGACY_BLOCKED] folds. A status this version does not know is
     * [UNKNOWN], never a judgment — telling somebody a decision is theirs on no
     * evidence is worse than saying the record is not understood here.
     */
    fun normalise(status: String): String {
        val candidate = status.trim().lowercase()
        if (candidate == LEGACY_BLOCKED) return REVIEW
        return if (candidate in order) candidate else UNKNOWN
    }

    /**
     * A saved status to filter by, as this version can draw it: [REVIEW] for
     * either of the ledger's two names for a judgment, and null — 전체 — for a
     * status this version has no chip for.
     *
     * A phone keeps the chip it was last filtering by across a rotation, a
     * fold and a process death. Left alone, a saved `blocked` would narrow the
     * list to a status no task carries any more and no chip can be pressed to
     * undo, so the panel would look empty with the work still there.
     */
    fun normaliseFilter(status: String?): String? {
        val candidate = status?.trim()?.lowercase() ?: return null
        if (candidate == LEGACY_BLOCKED) return REVIEW
        return candidate.takeIf { it in order }
    }
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
 * What the task list already says about a task's exchanges, as one value.
 *
 * The detail pane reads the exchanges themselves only when somebody opens
 * them, so it needs a cheap way to tell "the page I am holding is still the
 * current one" from "something happened and I should read again". Every field
 * here is one the panel's own poll already carries, so comparing them costs no
 * request at all — the panel is the signal, and no second one is invented.
 *
 * What each field catches: [turnsTotal] a new exchange, [latestTurnId] and
 * [running] the open one being taken up, [status] the task moving, and
 * [choiceTurnId] the session offering different sentences. A turn finishing
 * drops the open turn, which moves [latestTurnId] too.
 *
 * [none] is what a pane holding nothing has. Its [turnsTotal] is -1 so it can
 * never equal a real task's — a task with no exchanges yet has 0, and reading
 * "already current" from that would leave the pane empty for good.
 */
data class ChatTaskHistorySignal(
    val turnsTotal: Long = -1,
    val latestTurnId: Long = 0,
    val running: Boolean = false,
    val status: String = "",
    val choiceTurnId: Long = 0,
) {
    companion object {
        val none = ChatTaskHistorySignal()

        fun of(task: ChatTask): ChatTaskHistorySignal = ChatTaskHistorySignal(
            turnsTotal = task.turnsTotal,
            latestTurnId = task.waiting?.turnId ?: 0,
            running = task.waiting?.running ?: false,
            status = task.status,
            choiceTurnId = task.choiceTurnId,
        )
    }
}

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

/**
 * The task list, with every status folded to one this version draws
 * ([ChatTaskList.normalised]). The fold is here rather than in the panel so
 * there is one door the statuses come through: a screen, a count and a saved
 * filter cannot then disagree about what a record means.
 */
suspend fun ApiClient.chatTasks(): ChatTaskList =
    get("/api/chat/tasks", ChatTaskList.serializer()).normalised()

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
