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
data class ApprovalCallerKey(
    val id: Long,
    val name: String,
    @SerialName("key_prefix") val keyPrefix: String = "",
    @SerialName("service_account") val serviceAccount: String? = null,
    @SerialName("created_at") val createdAt: Long,
    @SerialName("last_used_at") val lastUsedAt: Long? = null,
    @SerialName("revoked_at") val revokedAt: Long? = null,
) {
    val isRevoked: Boolean get() = revokedAt != null

    val isServiceAccount: Boolean get() = serviceAccount != null
}

@Serializable
internal data class IssuedApprovalKeyResponse(
    val id: Long,
    val name: String,
    @SerialName("key_prefix") val keyPrefix: String = "",
    @SerialName("service_account") val serviceAccount: String? = null,
    @SerialName("created_at") val createdAt: Long,
    @SerialName("last_used_at") val lastUsedAt: Long? = null,
    @SerialName("revoked_at") val revokedAt: Long? = null,
    val key: String,
) {
    fun toIssued(): IssuedApprovalKey = IssuedApprovalKey(
        key = ApprovalCallerKey(
            id = id,
            name = name,
            keyPrefix = keyPrefix,
            serviceAccount = serviceAccount,
            createdAt = createdAt,
            lastUsedAt = lastUsedAt,
            revokedAt = revokedAt,
        ),
        secret = key,
    )
}

data class IssuedApprovalKey(val key: ApprovalCallerKey, val secret: String)

sealed interface ApprovalKeyIssueOutcome {
    data class Issued(val issued: IssuedApprovalKey) : ApprovalKeyIssueOutcome
    data object DuplicateName : ApprovalKeyIssueOutcome
}

sealed interface ApprovalServiceAccountOutcome {
    data class Registered(val key: ApprovalCallerKey) : ApprovalServiceAccountOutcome
    data object DuplicateName : ApprovalServiceAccountOutcome
    data object DuplicateServiceAccount : ApprovalServiceAccountOutcome
    data object InvalidServiceAccount : ApprovalServiceAccountOutcome
}

object ApprovalKeyFormatting {
    const val DUPLICATE_NAME_NOTICE = "이미 쓰이고 있는 호출자 이름입니다. 다른 이름으로 발급하세요"
    const val ISSUED_TITLE = "호출자 키를 발급했습니다"
    const val SECRET_WARNING = "이 화면을 닫으면 원문을 다시 볼 수 없습니다. 지금 복사해 호출자의 시크릿에 넣으세요."
    const val SERVICE_ACCOUNT_DUPLICATE_NAME_NOTICE = "이미 쓰이고 있는 호출자 이름입니다. 다른 이름으로 등록하세요"
    const val DUPLICATE_SERVICE_ACCOUNT_NOTICE = "이미 등록된 서비스 어카운트입니다"
    const val INVALID_SERVICE_ACCOUNT_NOTICE = "네임스페이스와 서비스 어카운트는 소문자·숫자·하이픈으로 63자 이내입니다"
    const val REVOKE_MESSAGE = "폐기하면 이 키로 요청을 만들거나 조회할 수 없고 되돌릴 수 없습니다. 이미 만든 대기 요청은 목록에 남습니다."
    const val SERVICE_ACCOUNT_REVOKE_MESSAGE =
        "폐기하면 이 서비스 어카운트의 토큰으로 요청을 만들거나 조회할 수 없고 되돌릴 수 없습니다. 이미 만든 대기 요청은 목록에 남고, 같은 서비스 어카운트는 다시 등록할 수 있습니다."

    private val dnsLabel = Regex("^[a-z0-9]([-a-z0-9]{0,61}[a-z0-9])?$")
    private val dayFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("MM/dd", Locale.KOREA)
    private val timeFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("HH:mm", Locale.KOREA)

    fun ordered(keys: List<ApprovalCallerKey>): List<ApprovalCallerKey> =
        keys.sortedWith(
            compareBy<ApprovalCallerKey> { it.isRevoked }
                .thenByDescending { it.createdAt }
                .thenByDescending { it.id },
        )

    fun meta(key: ApprovalCallerKey, nowMillis: Long, zone: ZoneId = ZoneId.systemDefault()): String {
        val lastUsed = key.lastUsedAt?.let { "마지막 사용 ${moment(it, nowMillis, zone)}" } ?: "마지막 사용 없음"
        val account = key.serviceAccount
        if (account != null) return "SA $account · 등록 ${day(key.createdAt, zone)} · $lastUsed"
        return "${key.keyPrefix}… · 발급 ${day(key.createdAt, zone)} · $lastUsed"
    }

    fun revokedLabel(key: ApprovalCallerKey, zone: ZoneId = ZoneId.systemDefault()): String? =
        key.revokedAt?.let { "폐기됨 ${day(it, zone)}" }

    fun revokeTitle(key: ApprovalCallerKey): String = "${key.name} 키를 폐기할까요?"

    fun revokeMessage(key: ApprovalCallerKey): String =
        if (key.isServiceAccount) SERVICE_ACCOUNT_REVOKE_MESSAGE else REVOKE_MESSAGE

    fun trimmedName(raw: String): String? = raw.trim().ifEmpty { null }

    fun serviceAccount(namespace: String, name: String): String? {
        val ns = namespace.trim()
        val account = name.trim()
        if (!isDnsLabel(ns) || !isDnsLabel(account)) return null
        return "$ns/$account"
    }

    internal fun isDnsLabel(value: String): Boolean = dnsLabel.matches(value)

    internal fun day(seconds: Long, zone: ZoneId): String = Instant.ofEpochSecond(seconds).atZone(zone).format(dayFormatter)

    private fun moment(seconds: Long, nowMillis: Long, zone: ZoneId): String {
        val date = Instant.ofEpochSecond(seconds).atZone(zone)
        val today = Instant.ofEpochMilli(nowMillis).atZone(zone).toLocalDate()
        return date.format(if (date.toLocalDate() == today) timeFormatter else dayFormatter)
    }
}

@Serializable
internal data class ApprovalKeyListResponse(val keys: List<ApprovalCallerKey> = emptyList())

@Serializable
internal data class ApprovalKeyIssuePayload(val name: String)

@Serializable
internal data class ApprovalServiceAccountPayload(
    val name: String,
    val kind: String = "k8s_sa",
    @SerialName("service_account") val serviceAccount: String,
)

suspend fun ApiClient.listApprovalKeys(): List<ApprovalCallerKey> =
    ApprovalKeyFormatting.ordered(get("/api/approval-keys", ApprovalKeyListResponse.serializer()).keys)

suspend fun ApiClient.issueApprovalKey(name: String): ApprovalKeyIssueOutcome =
    try {
        val issued = post(
            "/api/approval-keys",
            ApprovalKeyIssuePayload(name = name),
            ApprovalKeyIssuePayload.serializer(),
            IssuedApprovalKeyResponse.serializer(),
        )
        ApprovalKeyIssueOutcome.Issued(issued.toIssued())
    } catch (e: ApiException.BadStatus) {
        if (e.status != 409) throw e
        ApprovalKeyIssueOutcome.DuplicateName
    }

suspend fun ApiClient.registerApprovalServiceAccount(name: String, serviceAccount: String): ApprovalServiceAccountOutcome =
    try {
        val registered = post(
            "/api/approval-keys",
            ApprovalServiceAccountPayload(name = name, serviceAccount = serviceAccount),
            ApprovalServiceAccountPayload.serializer(),
            ApprovalCallerKey.serializer(),
        )
        ApprovalServiceAccountOutcome.Registered(registered)
    } catch (e: ApiException.BadStatus) {
        when {
            e.status == 409 && e.body.contains("service account already registered") ->
                ApprovalServiceAccountOutcome.DuplicateServiceAccount
            e.status == 409 -> ApprovalServiceAccountOutcome.DuplicateName
            e.status == 400 && e.body.contains("service_account") -> ApprovalServiceAccountOutcome.InvalidServiceAccount
            else -> throw e
        }
    }

suspend fun ApiClient.revokeApprovalKey(id: Long): ApprovalCallerKey =
    post(
        "/api/approval-keys/$id/revoke",
        ApprovalEmptyPayload(),
        ApprovalEmptyPayload.serializer(),
        ApprovalCallerKey.serializer(),
    )
