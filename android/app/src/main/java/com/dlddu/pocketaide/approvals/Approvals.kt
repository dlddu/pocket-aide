package com.dlddu.pocketaide.approvals

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Locale

@Serializable
enum class ApprovalStatus {
    @SerialName("PENDING")
    PENDING,

    @SerialName("APPROVED")
    APPROVED,

    @SerialName("REJECTED")
    REJECTED,

    @SerialName("EXPIRED")
    EXPIRED,
}

@Serializable
data class ApprovalRequest(
    val id: String,
    val status: ApprovalStatus = ApprovalStatus.PENDING,
    @SerialName("external_id") val externalId: String = "",
    val context: String,
    @SerialName("requester_name") val requesterName: String = "",
    @SerialName("session_id") val sessionId: String? = null,
    @SerialName("session_name") val sessionName: String? = null,
    val reason: String? = null,
    @SerialName("key_name") val keyName: String = "",
    @SerialName("key_revoked") val keyRevoked: Boolean = false,
    @SerialName("created_at") val createdAt: Long,
    @SerialName("expires_at") val expiresAt: Long? = null,
    @SerialName("processed_at") val processedAt: Long? = null,
    @SerialName("decision_mode") val decisionMode: String? = null,
) {
    fun isExpired(nowMillis: Long): Boolean {
        val deadline = expiresAt
        if (status != ApprovalStatus.PENDING || deadline == null) return status == ApprovalStatus.EXPIRED
        return deadline <= Math.floorDiv(nowMillis, 1000L)
    }
}

@Serializable
enum class ApprovalAutoResponse(val wire: String, val label: String) {
    @SerialName("NONE")
    OFF("NONE", "없음"),

    @SerialName("AUTO_APPROVE")
    AUTO_APPROVE("AUTO_APPROVE", "자동 승인"),

    @SerialName("AUTO_REJECT")
    AUTO_REJECT("AUTO_REJECT", "자동 거절");

    val optionLabel: String get() = if (this == OFF) "없음 — 알림으로 받기" else label

    val isAutomatic: Boolean get() = this != OFF

    val bannerTitle: String? get() = if (isAutomatic) "$label 모드가 켜져 있습니다." else null

    val outcomeNotice: String
        get() = when (this) {
            OFF -> "새 요청은 알림으로 받습니다."
            AUTO_APPROVE -> "새 요청은 알림 없이 바로 승인됩니다."
            AUTO_REJECT -> "새 요청은 알림 없이 바로 거절됩니다."
        }
}

@Serializable
data class ApprovalSettings(@SerialName("auto_response") val autoResponse: ApprovalAutoResponse)

enum class ApprovalDecision(val wire: String) {
    APPROVE("approve"),
    REJECT("reject"),
}

sealed interface ApprovalDecisionOutcome {
    val item: ApprovalRequest

    data class Decided(override val item: ApprovalRequest) : ApprovalDecisionOutcome
    data class Conflict(override val item: ApprovalRequest) : ApprovalDecisionOutcome
}

@Serializable
data class ApprovalHistoryPage(
    val items: List<ApprovalRequest> = emptyList(),
    @SerialName("next_cursor") val nextCursor: String? = null,
)

object ApprovalFormatting {
    const val NO_DEADLINE = "기한 없음"
    const val FAILED_DECISION_NOTICE = "결정되지 않았습니다. 연결을 확인한 뒤 다시 시도하세요"

    private val koreanLocale: Locale = Locale.KOREA

    fun newestFirst(items: List<ApprovalRequest>): List<ApprovalRequest> =
        items.sortedWith(compareByDescending<ApprovalRequest> { it.createdAt }.thenByDescending { it.id })

    fun remaining(expiresAt: Long?, nowMillis: Long): String {
        if (expiresAt == null) return NO_DEADLINE
        return "${clock(secondsLeft(expiresAt, nowMillis))} 남음"
    }

    fun countdown(expiresAt: Long, nowMillis: Long): String = clock(secondsLeft(expiresAt, nowMillis))

    fun progress(createdAt: Long, expiresAt: Long, nowMillis: Long): Float {
        val total = (expiresAt - createdAt).toDouble()
        if (total <= 0) return 0f
        val left = expiresAt.toDouble() - nowMillis / 1000.0
        return (left / total).coerceIn(0.0, 1.0).toFloat()
    }

    fun summary(context: String): String =
        context.lines().map { it.trim() }.firstOrNull { it.isNotEmpty() } ?: context.trim()

    fun sessionLabel(item: ApprovalRequest): String? {
        val name = item.sessionName.orEmpty().trim()
        if (name.isNotEmpty()) return name
        val id = item.sessionId.orEmpty().trim()
        if (id.isEmpty()) return null
        return "제목 없음 · ${id.take(8)}"
    }

    fun agentReason(item: ApprovalRequest): String? = item.reason.orEmpty().trim().ifEmpty { null }

    fun pendingHeader(count: Int): String = "대기 · ${count}건"

    fun statusLabel(status: ApprovalStatus): String = when (status) {
        ApprovalStatus.PENDING -> "대기 중"
        ApprovalStatus.APPROVED -> "승인됨"
        ApprovalStatus.REJECTED -> "거절됨"
        ApprovalStatus.EXPIRED -> "만료"
    }

    fun decisionModeLabel(status: ApprovalStatus, mode: String?): String {
        val value = mode.orEmpty()
        if (status == ApprovalStatus.EXPIRED || value == "expired") return "만료 — 결정 없음"
        if (value.startsWith("auto")) return if (status == ApprovalStatus.REJECTED) "자동 거절" else "자동 승인"
        return "수동 결정"
    }

    fun historyMeta(item: ApprovalRequest): String {
        val mode = decisionModeLabel(item.status, item.decisionMode)
        return if (item.keyRevoked) "$mode · 폐기된 키로 생성됨" else mode
    }

    fun historyTime(processedAt: Long, nowMillis: Long, zone: ZoneId = ZoneId.systemDefault()): String =
        dayRelativeTime(processedAt, "HH:mm:ss", nowMillis, zone)

    fun requestTime(createdAt: Long, nowMillis: Long, zone: ZoneId = ZoneId.systemDefault()): String =
        dayRelativeTime(createdAt, "HH:mm", nowMillis, zone)

    fun clockTime(seconds: Long?, zone: ZoneId = ZoneId.systemDefault()): String {
        if (seconds == null) return "—"
        return Instant.ofEpochSecond(seconds).atZone(zone).format(DateTimeFormatter.ofPattern("HH:mm:ss", koreanLocale))
    }

    fun confirmTitle(decision: ApprovalDecision): String =
        if (decision == ApprovalDecision.APPROVE) "승인할까요?" else "거절할까요?"

    fun confirmMessage(decision: ApprovalDecision, context: String): String {
        val verb = if (decision == ApprovalDecision.APPROVE) "허락합니다" else "거절합니다"
        val summary = summary(context)
        return "$summary${objectParticle(summary)} $verb. 확정하면 되돌릴 수 없습니다."
    }

    fun conflictNotice(item: ApprovalRequest): String =
        if (item.status == ApprovalStatus.EXPIRED) {
            "만료된 요청이라 결정되지 않았습니다"
        } else {
            "이미 처리된 요청입니다 — ${statusLabel(item.status)}"
        }

    internal fun objectParticle(text: String): String {
        val last = text.lastOrNull()?.code ?: return "을(를)"
        if (last !in HANGUL_FIRST..HANGUL_LAST) return "을(를)"
        return if ((last - HANGUL_FIRST) % JONG_COUNT == 0) "를" else "을"
    }

    private fun secondsLeft(expiresAt: Long, nowMillis: Long): Long =
        maxOf(0L, expiresAt - Math.floorDiv(nowMillis, 1000L))

    private fun clock(left: Long): String {
        val hours = left / 3600
        val minutes = (left % 3600) / 60
        val seconds = left % 60
        return if (hours > 0) {
            String.format(Locale.ROOT, "%d:%02d:%02d", hours, minutes, seconds)
        } else {
            String.format(Locale.ROOT, "%02d:%02d", minutes, seconds)
        }
    }

    private fun dayRelativeTime(seconds: Long, todayPattern: String, nowMillis: Long, zone: ZoneId): String {
        val moment = Instant.ofEpochSecond(seconds).atZone(zone)
        val today = Instant.ofEpochMilli(nowMillis).atZone(zone).toLocalDate()
        val day = moment.toLocalDate()
        return when (day) {
            today -> moment.format(DateTimeFormatter.ofPattern(todayPattern, koreanLocale))
            today.minusDays(1) -> "어제 " + moment.format(DateTimeFormatter.ofPattern("HH:mm", koreanLocale))
            else -> moment.format(DateTimeFormatter.ofPattern("MM/dd HH:mm", koreanLocale))
        }
    }

    private const val HANGUL_FIRST = 0xAC00
    private const val HANGUL_LAST = 0xD7A3
    private const val JONG_COUNT = 28
}

object ApprovalNotificationAction {
    const val CATEGORY = "APPROVAL_REQUEST"
    const val APPROVE_ACTION = "APPROVAL_APPROVE"
    const val REJECT_ACTION = "APPROVAL_REJECT"

    fun title(decision: ApprovalDecision): String = if (decision == ApprovalDecision.APPROVE) "승인" else "거절"

    fun action(decision: ApprovalDecision): String =
        if (decision == ApprovalDecision.APPROVE) APPROVE_ACTION else REJECT_ACTION

    fun decision(action: String?): ApprovalDecision? = when (action) {
        APPROVE_ACTION -> ApprovalDecision.APPROVE
        REJECT_ACTION -> ApprovalDecision.REJECT
        else -> null
    }

    fun resultNotice(result: Result<ApprovalDecisionOutcome>): String {
        val outcome = result.getOrNull() ?: return ApprovalFormatting.FAILED_DECISION_NOTICE
        return when (outcome) {
            is ApprovalDecisionOutcome.Decided ->
                "${ApprovalFormatting.statusLabel(outcome.item.status)} — ${ApprovalFormatting.summary(outcome.item.context)}"
            is ApprovalDecisionOutcome.Conflict -> ApprovalFormatting.conflictNotice(outcome.item)
        }
    }
}

@Serializable
internal class ApprovalEmptyPayload

@Serializable
internal data class ApprovalListResponse(val items: List<ApprovalRequest> = emptyList())

@Serializable
internal data class ApprovalConflictResponse(val approval: ApprovalRequest)

internal fun percentEncodeAlphanumerics(value: String): String = buildString {
    value.toByteArray(Charsets.UTF_8).forEach { byte ->
        val code = byte.toInt() and 0xFF
        val char = code.toChar()
        if (code < 0x80 && char.isLetterOrDigit()) append(char) else append(String.format(Locale.ROOT, "%%%02X", code))
    }
}

suspend fun ApiClient.listPendingApprovals(): List<ApprovalRequest> =
    ApprovalFormatting.newestFirst(get("/api/approvals", ApprovalListResponse.serializer()).items)

suspend fun ApiClient.approvalHistory(cursor: String? = null): ApprovalHistoryPage {
    val path = if (cursor == null) "/api/approvals/history" else "/api/approvals/history?cursor=${percentEncodeAlphanumerics(cursor)}"
    return get(path, ApprovalHistoryPage.serializer())
}

suspend fun ApiClient.approvalSettings(): ApprovalSettings =
    get("/api/approval-settings", ApprovalSettings.serializer())

suspend fun ApiClient.updateApprovalAutoResponse(mode: ApprovalAutoResponse): ApprovalSettings =
    patch(
        "/api/approval-settings",
        ApprovalSettings(autoResponse = mode),
        ApprovalSettings.serializer(),
        ApprovalSettings.serializer(),
    )

suspend fun ApiClient.approval(id: String): ApprovalRequest =
    get("/api/approvals/$id", ApprovalRequest.serializer())

suspend fun ApiClient.decideApproval(id: String, decision: ApprovalDecision): ApprovalDecisionOutcome =
    try {
        ApprovalDecisionOutcome.Decided(
            post(
                "/api/approvals/$id/${decision.wire}",
                ApprovalEmptyPayload(),
                ApprovalEmptyPayload.serializer(),
                ApprovalRequest.serializer(),
            ),
        )
    } catch (e: ApiException.BadStatus) {
        if (e.status != 409) throw e
        val conflict = try {
            ApiClient.json.decodeFromString(ApprovalConflictResponse.serializer(), e.body)
        } catch (decode: IllegalArgumentException) {
            throw ApiException.Decoding(decode)
        }
        ApprovalDecisionOutcome.Conflict(conflict.approval)
    }
