package com.dlddu.pocketaide.push

sealed interface PushDeepLink {
    data class Approval(val approvalId: String) : PushDeepLink
    data class ExternalNotification(val notificationId: Long) : PushDeepLink
    data class PrMonitorEvent(val eventId: Long) : PushDeepLink

    companion object {
        const val KEY_KIND = "kind"
        const val KEY_APPROVAL_ID = "approval_id"
        const val KEY_NOTIFICATION_ID = "notification_id"
        const val KEY_EVENT_ID = "event_id"
        const val KIND_EXTERNAL_NOTIFICATION = "external_notification"

        val KEYS = listOf(KEY_KIND, KEY_APPROVAL_ID, KEY_NOTIFICATION_ID, KEY_EVENT_ID)

        fun from(data: Map<String, String?>): PushDeepLink? {
            if (data[KEY_KIND] == KIND_EXTERNAL_NOTIFICATION) {
                return data[KEY_NOTIFICATION_ID]?.toLongOrNull()?.let(::ExternalNotification)
            }
            data[KEY_APPROVAL_ID]?.takeIf { it.isNotBlank() }?.let { return Approval(it) }
            return data[KEY_EVENT_ID]?.toLongOrNull()?.let(::PrMonitorEvent)
        }
    }
}
