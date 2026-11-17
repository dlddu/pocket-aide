package com.dlddu.pocketaide.notifications

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.EmptyResponse
import com.dlddu.pocketaide.approvals.ApprovalFormatting
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Locale

@Serializable
data class ExternalNotification(
    val id: Long,
    val title: String,
    val body: String = "",
    @SerialName("caller_name") val callerName: String = "",
    @SerialName("caller_kind") val callerKind: String = "",
    val requester: String = "",
    val devices: Int = 0,
    val sent: Int = 0,
    @SerialName("read_at") val readAt: Long? = null,
    @SerialName("created_at") val createdAt: Long,
) {
    val isUnread: Boolean get() = readAt == null

    fun markedRead(at: Long): ExternalNotification = copy(readAt = readAt ?: at)
}

@Serializable
data class ExternalNotificationsPage(
    val items: List<ExternalNotification> = emptyList(),
    @SerialName("unread_count") val unreadCount: Int = 0,
)

@Serializable
data class MarkAllExternalNotificationsReadResult(
    @SerialName("marked_count") val markedCount: Long = 0,
    @SerialName("unread_count") val unreadCount: Int = 0,
)

object ExternalNotificationFormatting {
    const val PAGE_SIZE = 50

    private val fullFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss", Locale.KOREA)

    fun callerLabel(item: ExternalNotification): String =
        "${item.callerName} · ${if (item.callerKind == "k8s_sa") "SA" else "키"}"

    fun requesterLabel(item: ExternalNotification): String =
        if (item.requester.isEmpty()) "요청 주체 미상" else "요청 ${item.requester}"

    fun deliveryLabel(item: ExternalNotification): String =
        if (item.devices == 0) "등록된 기기 없음" else "${item.sent}/${item.devices} 기기"

    fun time(item: ExternalNotification, nowMillis: Long, zone: ZoneId = ZoneId.systemDefault()): String =
        ApprovalFormatting.requestTime(item.createdAt, nowMillis, zone)

    fun fullTime(item: ExternalNotification, zone: ZoneId = ZoneId.systemDefault()): String =
        Instant.ofEpochSecond(item.createdAt).atZone(zone).format(fullFormatter)

    fun requesterValue(item: ExternalNotification): String = item.requester.ifEmpty { "요청 주체 미상" }
}

@Serializable
internal class ExternalNotificationReadPayload

suspend fun ApiClient.externalNotifications(limit: Int? = null, before: Long? = null): ExternalNotificationsPage {
    val query = buildList {
        if (limit != null) add("limit=$limit")
        if (before != null) add("before=$before")
    }
    val path = if (query.isEmpty()) "/api/external-notifications" else "/api/external-notifications?" + query.joinToString("&")
    return get(path, ExternalNotificationsPage.serializer())
}

suspend fun ApiClient.externalNotification(id: Long): ExternalNotification =
    get("/api/external-notifications/$id", ExternalNotification.serializer())

suspend fun ApiClient.markExternalNotificationRead(id: Long) {
    post(
        "/api/external-notifications/$id/read",
        ExternalNotificationReadPayload(),
        ExternalNotificationReadPayload.serializer(),
        EmptyResponse.serializer(),
    )
}

suspend fun ApiClient.markAllExternalNotificationsRead(): MarkAllExternalNotificationsReadResult =
    post(
        "/api/external-notifications/read-all",
        ExternalNotificationReadPayload(),
        ExternalNotificationReadPayload.serializer(),
        MarkAllExternalNotificationsReadResult.serializer(),
    )
