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
    // How many requests a person left on this task are still open. Absent when
    // none ever arrived, so "nothing waiting" and "nothing to say" are not the
    // same answer. The parent's own progress is what moves these.
    val requests: ChatTaskRequests? = null,
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
 * One ready-made follow-up. [label] says what it asks for and [request] is the
 * sentence that is actually sent — the panel puts the request in the editable
 * field before anything goes, so a label can never promise what the message
 * does not say.
 */
data class ChatTaskSuggestion(val label: String, val request: String)

/**
 * Follow-ups worth offering for one task, grounded in what the ledger records.
 * A recorded reason or next action is quoted verbatim; when the ledger has
 * none, the request says it has none and asks — it never supplies a reason
 * nobody wrote, and it never states a next step as settled.
 *
 * None of these instructs the session directly. Each is a message the user
 * sends into their own conversation, which the session reads and decides on,
 * so a tool approval or a human merge gate still happens where it did before.
 */
fun chatTaskSuggestions(task: ChatTask): List<ChatTaskSuggestion> {
    val reason = task.reason.trim()
    val next = task.nextAction.trim()
    val out = mutableListOf<ChatTaskSuggestion>()
    when (task.status) {
        ChatTaskStatus.BLOCKED -> out += if (reason.isNotEmpty()) {
            ChatTaskSuggestion("차단 해제 방법 받기", "이 작업의 차단 사유가 원장에 「$reason」로 적혀 있어. 이걸 풀 방법을 알려 줘.")
        } else {
            ChatTaskSuggestion("차단 사유 물어보기", "원장에 차단 사유가 적혀 있지 않아. 무엇 때문에 막혀 있는지 확인해서 알려 줘.")
        }

        ChatTaskStatus.REVIEW -> {
            out += if (reason.isNotEmpty()) {
                ChatTaskSuggestion("검토할 내용 받기", "이 작업의 검토 사유가 원장에 「$reason」로 적혀 있어. 내가 무엇을 보고 결정해야 하는지 알려 줘.")
            } else {
                ChatTaskSuggestion("검토할 내용 받기", "원장에 검토 사유가 적혀 있지 않아. 내가 무엇을 보고 결정해야 하는지 정리해서 알려 줘.")
            }
            out += ChatTaskSuggestion("진행해도 된다고 전하기", "검토 내용을 확인했어. 이 작업을 이대로 진행해 줘.")
            out += ChatTaskSuggestion("잠시 멈추라고 전하기", "이 작업은 아직 진행하지 말고, 내가 다시 말할 때까지 기다려 줘.")
        }

        ChatTaskStatus.RUNNING -> {
            out += ChatTaskSuggestion("지금까지 진행 내용 받기", "이 작업을 지금까지 어디까지 진행했는지 정리해서 알려 줘.")
            out += ChatTaskSuggestion("계속 진행하라고 전하기", "이 작업을 계속 진행해 줘.")
        }

        ChatTaskStatus.PLANNED -> {
            out += ChatTaskSuggestion("시작하라고 전하기", "이 작업을 시작해 줘.")
            out += ChatTaskSuggestion("무엇부터 할지 받기", "이 작업을 시작하면 무엇부터 하게 되는지 알려 줘.")
        }

        ChatTaskStatus.COMPLETED -> {
            out += ChatTaskSuggestion("결과 요약 받기", "이 작업의 결과를 요약해서 알려 줘.")
            out += ChatTaskSuggestion("남은 일 확인하기", "이 작업에서 더 남은 일이 있는지 알려 줘.")
        }

        // A status this app does not know. Ask what it is rather than name one.
        else -> out += ChatTaskSuggestion(
            "상태 확인하기",
            "원장에 적힌 이 작업의 상태 값을 앱이 알지 못해. 지금 어떤 상태인지 확인해서 알려 줘.",
        )
    }
    // The ledger's own next action, quoted. A completed task gets this only when
    // one is actually recorded — asking "what is next" about finished work would
    // invent a step the ledger does not have.
    if (next.isNotEmpty()) {
        out += ChatTaskSuggestion("적힌 다음 행동 진행", "원장에 적힌 다음 행동은 「$next」야. 이걸 진행해 줘.")
    } else if (task.status != ChatTaskStatus.COMPLETED) {
        out += ChatTaskSuggestion("다음 행동 받기", "원장에 다음 행동이 적혀 있지 않아. 다음에 무엇을 할지 정해서 알려 줘.")
    }
    return out
}

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
    INVALID("요청을 저장할 수 없는 형식이에요."),
    UNREACHABLE("요청을 저장하지 못했어요. 잠시 뒤 다시 시도해요."),
    ;

    companion object {
        fun from(e: Exception): ChatTaskRequestError = when (e) {
            is ApiException.BadStatus -> when {
                e.status == 404 -> NOT_FOUND
                e.status == 501 -> UNSUPPORTED
                e.status == 409 && e.body.contains("task_request_unavailable") -> UNAVAILABLE
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
 * Stores one request on [taskId]'s ledger row. [id] is the app's own key for
 * this request, so sending it again after an answer the app never saw stores
 * nothing twice. The backend derives the session and who asked from the signed
 * in user — neither is in this body.
 */
suspend fun ApiClient.storeChatTaskRequest(taskId: String, id: String, text: String): ChatTaskRequestStored =
    post(
        "/api/chat/tasks/$taskId/requests",
        ChatTaskRequestPayload(id = id, text = text),
        ChatTaskRequestPayload.serializer(),
        ChatTaskRequestStored.serializer(),
    )

@Serializable
private data class ChatTaskRequestPayload(val id: String, val text: String)
