package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.contentOrNull

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
    // How many requests a person left on this task are still open. Absent when
    // none ever arrived, so "nothing waiting" and "nothing to say" are not the
    // same answer. The parent's own progress is what moves these.
    val requests: ChatTaskRequests? = null,
    // What the session itself offers to say about this task, written against
    // the version this row carries. Empty means it offers none — the screen
    // says so and keeps free text; the app never makes up a choice.
    val choices: List<ChatTaskChoice> = emptyList(),
)

/**
 * One sentence the session offers for this task. The button shows [label] and
 * fills the request box with [requestText]; [id] goes back with the stored
 * request so the session knows which of its own sentences came back.
 *
 * It is a sentence to send, not a permission: picking one approves no tool and
 * no merge, and nothing is sent until the person stores it.
 */
@Serializable
data class ChatTaskChoice(
    val id: String,
    val label: String = "",
    @SerialName("request_text") val requestText: String = "",
    @SerialName("basis_version") val basisVersion: Long = 0,
    @SerialName("created_at") val createdAt: String = "",
)

/**
 * The requests waiting on one task, as the ledger counts them. Pending is
 * nobody has taken it yet, claimed is the session is on it, and reconciling is
 * a claim that lapsed after something had already been handed over — the
 * session has to find out what landed before sending anything again.
 */
@Serializable
data class ChatTaskRequests(
    val pending: Int = 0,
    val claimed: Int = 0,
    val reconciling: Int = 0,
    @SerialName("last_seq") val lastSeq: Long = 0,
) {
    /** What the session has not settled yet — the number the panel shows. */
    val open: Int get() = pending + claimed + reconciling
}

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
        if (tasks.isEmpty()) return ChatTaskEmptiness.LEDGER
        if (select(tasks).isNotEmpty()) return null
        // Rows of the chosen status exist, so hiding completed work is what
        // emptied the list; otherwise the ledger simply has none of this status.
        val ofStatus = tasks.filter { status == null || it.status == status }
        return if (ofStatus.isEmpty()) ChatTaskEmptiness.FILTERED else ChatTaskEmptiness.COMPLETED_HIDDEN
    }
}

/** Which of the three "nothing to show" answers the panel owes the user. */
enum class ChatTaskEmptiness { LEDGER, FILTERED, COMPLETED_HIDDEN }

/**
 * One request stored on a task's ledger row.
 *
 * It is not a chat message: it is written to the session's own task ledger and
 * stays there until the session picks it up, so leaving one never interrupts a
 * reply being written and never puts words in the conversation. [status] is the
 * ledger's, and only the session moves it — the app shows it and never infers
 * it.
 */
@Serializable
data class ChatTaskRequest(
    val id: String,
    @SerialName("task_id") val taskId: String = "",
    val seq: Long = 0,
    val text: String = "",
    val status: String = ChatTaskRequestStatus.PENDING,
    val result: String = "",
) {
    /** Whether the session has finished with this request, either way. */
    val settled: Boolean
        get() = status == ChatTaskRequestStatus.PROCESSED || status == ChatTaskRequestStatus.FAILED
}

/** What the backend answers when a request is stored. */
@Serializable
data class ChatTaskRequestStored(
    val request: ChatTaskRequest,
    // True when this id had already stored the same words: the app's retry of
    // an answer it never saw, not a second request.
    val replayed: Boolean = false,
)

object ChatTaskRequestStatus {
    const val PENDING = "pending"
    const val CLAIMED = "claimed"
    const val RECONCILING = "reconciling"
    const val PROCESSED = "processed"
    const val FAILED = "failed"

    /** What the person is told the request is doing now. */
    fun label(status: String): String = when (status) {
        PENDING -> "저장됨 · 확인 대기"
        CLAIMED -> "확인 중"
        RECONCILING -> "확인 중 · 결과 대조"
        PROCESSED -> "처리됨"
        FAILED -> "처리 실패"
        else -> "알 수 없음"
    }
}

/**
 * Why a request could not be stored. Each is a different sentence because the
 * person's next move differs: wait, retry, or nothing to do here. None of them
 * says the request was stored.
 */
enum class ChatTaskRequestError(val message: String) {
    NOT_FOUND("이 작업이 원장에서 사라졌어요. 목록을 다시 불러요."),
    UNAVAILABLE("채팅 세션이 깨어 있지 않아 저장하지 못했어요. 잠시 뒤 다시 시도해요."),
    UNSUPPORTED("이 대화의 세션은 작업 요청을 받지 않아요."),
    CONFLICT("같은 요청 번호로 다른 내용이 이미 저장돼 있어요. 다시 쓰면 새 번호로 저장해요."),
    STALE_CHOICE("작업이 바뀌어 고른 선택지가 더 이상 유효하지 않아요. 쓴 내용은 그대로 두었으니, 새 선택지를 보거나 그대로 다시 저장해요."),
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
 * One line of a task's history. The session writes these as it works and the
 * app shows them in order — it never fills a gap with the task's current state.
 */
@Serializable
data class ChatTaskEvent(
    val seq: Long = 0,
    @SerialName("task_id") val taskId: String = "",
    val at: String = "",
    val kind: String = "",
    val actor: String = "",
    val summary: String = "",
    // The ledger's own bounded JSON, so a value is not always a string: a
    // task_created line carries its version as a number and a published-choices
    // line carries its labels as an array. Decoding it as strings made every
    // history that has such a line fail to read at all, which the screen could
    // only report as "could not load".
    val detail: Map<String, JsonElement> = emptyMap(),
    @SerialName("request_id") val requestId: String = "",
) {
    /** The person's own words, when this line is a request they left. */
    val text: String get() = detailText("text")

    /** What the session wrote back, when this line settles a request. */
    val result: String get() = detailText("result")

    private fun detailText(key: String): String =
        (detail[key] as? JsonPrimitive)?.contentOrNull.orEmpty()
}

@Serializable
data class ChatTaskEventPage(
    val events: List<ChatTaskEvent> = emptyList(),
    @SerialName("next_cursor") val nextCursor: Long = 0,
)

object ChatTaskEventKind {
    /** What each line is, said in the words a person reads. */
    fun label(kind: String): String = when (kind) {
        "task_created" -> "작업 등록"
        "task_updated" -> "작업 변경"
        "task_deleted" -> "작업 삭제"
        "request_stored" -> "내 요청"
        "request_claimed" -> "요청 접수"
        "request_dispatched" -> "처리 시작"
        "request_settled" -> "요청 종료"
        "choices_published" -> "선택지 갱신"
        "choices_cleared" -> "선택지 비움"
        else -> "기록"
    }
}

/** Why a task's history could not be read. */
enum class ChatTaskHistoryError(val message: String) {
    NOT_FOUND("이 작업이 원장에서 사라졌어요."),
    UNAVAILABLE("채팅 세션이 깨어 있지 않아 이력을 읽지 못했어요."),
    UNSUPPORTED("이 대화의 세션은 작업 이력을 두지 않아요."),
    UNREACHABLE("작업 이력을 불러오지 못했어요. 다시 시도해요."),
    ;

    companion object {
        fun from(e: Exception): ChatTaskHistoryError = when (e) {
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

suspend fun ApiClient.chatTasks(): ChatTaskList = get("/api/chat/tasks", ChatTaskList.serializer())

/** One page of [taskId]'s history, oldest first. */
suspend fun ApiClient.chatTaskEvents(taskId: String, afterSeq: Long, limit: Int): ChatTaskEventPage {
    val query = buildList {
        if (afterSeq > 0) add("after_seq=$afterSeq")
        if (limit > 0) add("limit=$limit")
    }.joinToString("&")
    val path = "/api/chat/tasks/$taskId/events" + if (query.isEmpty()) "" else "?$query"
    return get(path, ChatTaskEventPage.serializer())
}

/**
 * Stores one request on [taskId]'s ledger row. [id] is the app's own key for
 * this request, so sending it again after an answer the app never saw stores
 * nothing twice. The backend derives the session and who asked from the signed
 * in user — neither is in this body.
 */
suspend fun ApiClient.storeChatTaskRequest(
    taskId: String,
    id: String,
    text: String,
    choiceId: String,
    choiceBasisVersion: Long = 0,
): ChatTaskRequestStored =
    post(
        "/api/chat/tasks/$taskId/requests",
        ChatTaskRequestPayload(
            id = id,
            text = text,
            choiceId = choiceId,
            choiceBasisVersion = choiceBasisVersion,
        ),
        ChatTaskRequestPayload.serializer(),
        ChatTaskRequestStored.serializer(),
    )

@Serializable
private data class ChatTaskRequestPayload(
    val id: String,
    val text: String,
    @SerialName("choice_id") val choiceId: String = "",
    // The task version the screen saw that choice on. The slot refuses a pick
    // whose basis no longer holds, so an open screen cannot act on a sentence
    // the session wrote about an older state.
    @SerialName("choice_basis_version") val choiceBasisVersion: Long = 0,
)
