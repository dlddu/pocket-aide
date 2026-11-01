package com.dlddu.pocketaide.prmonitor

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.EmptyResponse
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class NotificationHistoryItem(
    val id: Long,
    @SerialName("repo_full_name") val repoFullName: String,
    @SerialName("pr_number") val prNumber: Int? = null,
    @SerialName("pr_title") val prTitle: String? = null,
    @SerialName("pr_url") val prUrl: String? = null,
    @SerialName("commit_url") val commitUrl: String? = null,
    @SerialName("run_url") val runUrl: String? = null,
    @SerialName("workflow_name") val workflowName: String = "",
    @SerialName("head_branch") val headBranch: String = "",
    @SerialName("head_sha") val headSha: String = "",
    val conclusion: String = "",
    @SerialName("acknowledged_at") val acknowledgedAt: Long? = null,
    @SerialName("created_at") val createdAt: Long,
) {
    val isAcknowledged: Boolean get() = acknowledgedAt != null
}

@Serializable
data class ExcludedRepo(
    val id: Long,
    @SerialName("repo_full_name") val repoFullName: String,
    @SerialName("created_at") val createdAt: Long = 0,
)

@Serializable
enum class NotificationOutcomes(val wire: String, val label: String) {
    @SerialName("both")
    BOTH("both", "둘 다"),

    @SerialName("success")
    SUCCESS("success", "성공만"),

    @SerialName("failure")
    FAILURE("failure", "실패만"),
}

@Serializable
data class NotificationSettings(
    val enabled: Boolean = true,
    val outcomes: NotificationOutcomes = NotificationOutcomes.BOTH,
)

@Serializable
data class NotificationSettingsPatch(
    val enabled: Boolean? = null,
    val outcomes: NotificationOutcomes? = null,
)

enum class CiOutcome { SUCCESS, IN_PROGRESS, FAILURE }

object CiConclusion {
    private val inProgress = setOf("queued", "requested", "in_progress", "pending", "waiting")
    private val failures = setOf("failure", "timed_out", "cancelled", "action_required")

    fun isSuccess(conclusion: String): Boolean = conclusion.lowercase() == "success"

    fun isInProgress(conclusion: String): Boolean = conclusion.lowercase() in inProgress

    fun isFailure(conclusion: String): Boolean = conclusion.lowercase() in failures

    fun outcome(conclusion: String): CiOutcome = when {
        isSuccess(conclusion) -> CiOutcome.SUCCESS
        isInProgress(conclusion) -> CiOutcome.IN_PROGRESS
        else -> CiOutcome.FAILURE
    }

    fun verdictLabel(conclusion: String): String = when (conclusion.lowercase()) {
        "success" -> "CI 통과"
        "failure" -> "CI 실패"
        "cancelled" -> "CI 취소"
        "timed_out" -> "CI 타임아웃"
        in inProgress -> "CI 시작"
        else -> "CI $conclusion"
    }
}

data class HistoryGroup(
    val id: String,
    val repoFullName: String,
    val prNumber: Int?,
    val prTitle: String?,
    val prUrl: String?,
    val headBranch: String,
    val headSha: String,
    val items: List<NotificationHistoryItem>,
) {
    val unacknowledgedCount: Int get() = items.count { !it.isAcknowledged }
    val allAcknowledged: Boolean get() = unacknowledgedCount == 0
    val latestCreatedAt: Long get() = items.maxOfOrNull { it.createdAt } ?: 0
    val successCount: Int get() = items.count { CiConclusion.isSuccess(it.conclusion) }
    val failureCount: Int get() = items.count { CiConclusion.isFailure(it.conclusion) }
    val inProgressCount: Int get() = items.count { CiConclusion.isInProgress(it.conclusion) }
    val leadItem: NotificationHistoryItem? get() = items.firstOrNull()

    fun contains(eventId: Long?): Boolean = eventId != null && items.any { it.id == eventId }
}

object HistoryGrouping {
    fun group(items: List<NotificationHistoryItem>): List<HistoryGroup> {
        if (items.isEmpty()) return emptyList()
        val byKey = LinkedHashMap<String, MutableList<NotificationHistoryItem>>()
        items.forEach { byKey.getOrPut(groupKey(it)) { mutableListOf() }.add(it) }
        val groups = byKey.map { (key, bucket) ->
            val sorted = bucket.sortedByDescending { it.createdAt }
            val lead = sorted.first()
            HistoryGroup(
                id = key,
                repoFullName = lead.repoFullName,
                prNumber = lead.prNumber,
                prTitle = lead.prTitle,
                prUrl = lead.prUrl,
                headBranch = lead.headBranch,
                headSha = lead.headSha,
                items = sorted,
            )
        }
        return groups.sortedWith(
            compareBy<HistoryGroup> { it.allAcknowledged }.thenByDescending { it.latestCreatedAt },
        )
    }

    fun groupKey(item: NotificationHistoryItem): String {
        item.prNumber?.let { return "pr:${item.repoFullName}:$it" }
        if (item.headSha.isNotEmpty()) return "sha:${item.repoFullName}:${item.headSha}"
        return "row:${item.id}"
    }
}

object PrMonitorText {
    fun rowTitle(item: NotificationHistoryItem): String {
        val number = item.prNumber
        val title = item.prTitle
        return when {
            number != null && !title.isNullOrEmpty() -> "${item.repoFullName} · #$number $title"
            number != null -> "${item.repoFullName} · #$number"
            item.headBranch.isEmpty() -> item.repoFullName
            else -> "${item.repoFullName} · ${item.headBranch}"
        }
    }

    fun shortSha(sha: String): String = sha.take(SHORT_SHA_LENGTH)

    fun prTitleSuffix(title: String?): String = if (title.isNullOrEmpty()) "" else " $title"

    fun shortShaSuffix(sha: String): String = if (sha.isEmpty()) "" else " @${shortSha(sha)}"

    private const val SHORT_SHA_LENGTH = 7
}

@Serializable
internal data class NotificationHistoryListResponse(val items: List<NotificationHistoryItem> = emptyList())

@Serializable
internal data class ExcludedRepoListResponse(val items: List<ExcludedRepo> = emptyList())

@Serializable
internal data class ExcludedRepoPayload(@SerialName("repo_full_name") val repoFullName: String)

@Serializable
internal class PrMonitorEmptyPayload

suspend fun ApiClient.listNotificationHistory(limit: Int? = null, before: Long? = null): List<NotificationHistoryItem> {
    val query = listOfNotNull(limit?.let { "limit=$it" }, before?.let { "before=$it" })
    val path = if (query.isEmpty()) "/api/notification-history" else "/api/notification-history?" + query.joinToString("&")
    return get(path, NotificationHistoryListResponse.serializer()).items
}

suspend fun ApiClient.acknowledgeNotification(id: Long) {
    post(
        "/api/notification-history/$id/ack",
        PrMonitorEmptyPayload(),
        PrMonitorEmptyPayload.serializer(),
        EmptyResponse.serializer(),
    )
}

suspend fun ApiClient.listExcludedRepos(): List<ExcludedRepo> =
    get("/api/excluded-repos", ExcludedRepoListResponse.serializer()).items

suspend fun ApiClient.addExcludedRepo(repoFullName: String): ExcludedRepo =
    post(
        "/api/excluded-repos",
        ExcludedRepoPayload(repoFullName),
        ExcludedRepoPayload.serializer(),
        ExcludedRepo.serializer(),
    )

suspend fun ApiClient.removeExcludedRepo(id: Long) {
    delete("/api/excluded-repos/$id")
}

suspend fun ApiClient.notificationSettings(): NotificationSettings =
    get("/api/notification-settings", NotificationSettings.serializer())

suspend fun ApiClient.updateNotificationSettings(change: NotificationSettingsPatch): NotificationSettings =
    this.patch(
        "/api/notification-settings",
        change,
        NotificationSettingsPatch.serializer(),
        NotificationSettings.serializer(),
    )
