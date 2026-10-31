package com.dlddu.pocketaide.push

import com.dlddu.pocketaide.ui.RootTab
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class PushDeepLinkTest {
    @Test
    fun externalNotificationKindWins() {
        val link = PushDeepLink.from(mapOf("kind" to "external_notification", "notification_id" to "42", "event_id" to "9"))

        assertEquals(PushDeepLink.ExternalNotification(42), link)
        assertEquals(RootTab.Notifications, RootTab.forDeepLink(link!!))
    }

    @Test
    fun approvalIdRoutesToApprovalTab() {
        val link = PushDeepLink.from(mapOf("approval_id" to "apr_1", "category" to "APPROVAL_REQUEST"))

        assertEquals(PushDeepLink.Approval("apr_1"), link)
        assertEquals(RootTab.Approval, RootTab.forDeepLink(link!!))
    }

    @Test
    fun eventIdRoutesToPrMonitor() {
        val link = PushDeepLink.from(mapOf("event_id" to "17"))

        assertEquals(PushDeepLink.PrMonitorEvent(17), link)
        assertEquals(RootTab.PrMonitor, RootTab.forDeepLink(link!!))
    }

    @Test
    fun malformedOrEmptyPayloadIsIgnored() {
        assertNull(PushDeepLink.from(emptyMap()))
        assertNull(PushDeepLink.from(mapOf("event_id" to "abc")))
        assertNull(PushDeepLink.from(mapOf("kind" to "external_notification")))
    }

    @Test
    fun approvalCategoryUsesApprovalChannel() {
        assertEquals(NotificationChannels.APPROVAL, NotificationChannels.forCategory("APPROVAL_REQUEST"))
        assertEquals(NotificationChannels.DEFAULT, NotificationChannels.forCategory(null))
    }
}
