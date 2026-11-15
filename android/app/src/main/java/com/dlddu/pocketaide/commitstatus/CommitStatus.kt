package com.dlddu.pocketaide.commitstatus

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.approvals.ApprovalDecision
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class CommitStatusRequestStatus {
    @SerialName("PENDING")
    PENDING,

    @SerialName("APPROVED")
    APPROVED,

    @SerialName("REJECTED")
    REJECTED,
}

@Serializable
data class CommitStatusRequest(
    val id: String,
    val status: CommitStatusRequestStatus = CommitStatusRequestStatus.PENDING,
    val owner: String,
    val repo: String,
    val sha: String,
    val state: String,
    val context: String,
    val description: String = "",
    @SerialName("target_url") val targetUrl: String = "",
    val requester: String = "",
    val reason: String = "",
    @SerialName("session_id") val sessionId: String = "",
    @SerialName("session_name") val sessionName: String = "",
    @SerialName("key_name") val keyName: String = "",
    @SerialName("key_revoked") val keyRevoked: Boolean = false,
    @SerialName("github_status_id") val githubStatusId: Long? = null,
    @SerialName("last_error") val lastError: String = "",
    @SerialName("created_at") val createdAt: Long,
    @SerialName("processed_at") val processedAt: Long? = null,
)

@Serializable
data class CommitStatusRequestList(
    val items: List<CommitStatusRequest> = emptyList(),
    val history: List<CommitStatusRequest> = emptyList(),
)

sealed interface CommitStatusDecisionOutcome {
    data class Decided(val item: CommitStatusRequest) : CommitStatusDecisionOutcome
    data class Conflict(val item: CommitStatusRequest) : CommitStatusDecisionOutcome
    data class WriteFailed(val item: CommitStatusRequest, val message: String) : CommitStatusDecisionOutcome
    data class Unavailable(val message: String) : CommitStatusDecisionOutcome
}

object CommitStatusFormatting {
    const val FAILED_DECISION_NOTICE = "결정되지 않았습니다. 연결을 확인한 뒤 다시 시도하세요"
    const val UNAVAILABLE_NOTICE = "커밋 상태를 쓸 GitHub App 이 설정되지 않아 승인할 수 없습니다"
    const val HISTORY_HEADER = "최근 처리"

    fun title(item: CommitStatusRequest): String = "${item.context} → ${item.state}"

    fun shortSha(sha: String): String = sha.take(7)

    fun target(item: CommitStatusRequest): String = "${item.owner}/${item.repo}@${shortSha(item.sha)}"

    fun statusLabel(status: CommitStatusRequestStatus): String = when (status) {
        CommitStatusRequestStatus.PENDING -> "대기 중"
        CommitStatusRequestStatus.APPROVED -> "승인됨 — 기록함"
        CommitStatusRequestStatus.REJECTED -> "거절됨"
    }

    fun pendingHeader(count: Int): String = "대기 · ${count}건"

    fun confirmTitle(decision: ApprovalDecision): String =
        if (decision == ApprovalDecision.APPROVE) "커밋 상태를 기록할까요?" else "거절할까요?"

    fun confirmMessage(decision: ApprovalDecision, item: CommitStatusRequest): String =
        if (decision == ApprovalDecision.APPROVE) {
            "${target(item)} 에 ${title(item)} 을(를) GitHub 에 기록합니다. 확정하면 되돌릴 수 없습니다."
        } else {
            "${title(item)} 요청을 거절합니다. GitHub 에는 아무것도 기록하지 않습니다."
        }

    fun conflictNotice(item: CommitStatusRequest): String = "이미 처리된 요청입니다 — ${statusLabel(item.status)}"

    fun writeFailedNotice(message: String): String = "GitHub 에 기록하지 못했습니다 — $message"

    fun newestFirst(items: List<CommitStatusRequest>): List<CommitStatusRequest> =
        items.sortedWith(compareByDescending<CommitStatusRequest> { it.createdAt }.thenByDescending { it.id })
}

object CommitStatusNotificationAction {
    const val CATEGORY = "COMMIT_STATUS_REQUEST"
    const val APPROVE_ACTION = "COMMIT_STATUS_APPROVE"
    const val REJECT_ACTION = "COMMIT_STATUS_REJECT"

    fun title(decision: ApprovalDecision): String = if (decision == ApprovalDecision.APPROVE) "승인" else "거절"

    fun action(decision: ApprovalDecision): String =
        if (decision == ApprovalDecision.APPROVE) APPROVE_ACTION else REJECT_ACTION

    fun decision(action: String?): ApprovalDecision? = when (action) {
        APPROVE_ACTION -> ApprovalDecision.APPROVE
        REJECT_ACTION -> ApprovalDecision.REJECT
        else -> null
    }

    fun resultNotice(result: Result<CommitStatusDecisionOutcome>): String {
        val outcome = result.getOrNull() ?: return CommitStatusFormatting.FAILED_DECISION_NOTICE
        return when (outcome) {
            is CommitStatusDecisionOutcome.Decided ->
                "${CommitStatusFormatting.statusLabel(outcome.item.status)} — ${CommitStatusFormatting.title(outcome.item)}"
            is CommitStatusDecisionOutcome.Conflict -> CommitStatusFormatting.conflictNotice(outcome.item)
            is CommitStatusDecisionOutcome.WriteFailed -> CommitStatusFormatting.writeFailedNotice(outcome.message)
            is CommitStatusDecisionOutcome.Unavailable -> CommitStatusFormatting.UNAVAILABLE_NOTICE
        }
    }
}

@Serializable
internal class CommitStatusEmptyPayload

@Serializable
internal data class CommitStatusErrorResponse(val error: String = "", val request: CommitStatusRequest? = null)

suspend fun ApiClient.listCommitStatusRequests(): CommitStatusRequestList {
    val list = get("/api/commit-status-requests", CommitStatusRequestList.serializer())
    return list.copy(items = CommitStatusFormatting.newestFirst(list.items))
}

suspend fun ApiClient.commitStatusRequest(id: String): CommitStatusRequest =
    get("/api/commit-status-requests/$id", CommitStatusRequest.serializer())

suspend fun ApiClient.decideCommitStatusRequest(id: String, decision: ApprovalDecision): CommitStatusDecisionOutcome =
    try {
        CommitStatusDecisionOutcome.Decided(
            post(
                "/api/commit-status-requests/$id/${decision.wire}",
                CommitStatusEmptyPayload(),
                CommitStatusEmptyPayload.serializer(),
                CommitStatusRequest.serializer(),
            ),
        )
    } catch (e: ApiException.BadStatus) {
        if (e.status !in listOf(409, 502, 503)) throw e
        val body = try {
            ApiClient.json.decodeFromString(CommitStatusErrorResponse.serializer(), e.body)
        } catch (decode: IllegalArgumentException) {
            throw ApiException.Decoding(decode)
        }
        val item = body.request
        if (e.status == 503) {
            CommitStatusDecisionOutcome.Unavailable(body.error)
        } else if (item == null) {
            throw e
        } else if (e.status == 409) {
            CommitStatusDecisionOutcome.Conflict(item)
        } else {
            CommitStatusDecisionOutcome.WriteFailed(item, body.error)
        }
    }
