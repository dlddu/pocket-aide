package com.dlddu.pocketaide.notifications

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.time.ZoneId
import java.time.ZonedDateTime

class ExternalNotificationsViewModelTest {
    private val server = MockWebServer()
    private val now = 1_800_000_000_000L
    private var markedRead = 0

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(): ExternalNotificationsViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ExternalNotificationsViewModel(
            ApiClient(server.url("/").toString(), store, clock = { now }),
            onMarkedRead = { markedRead += 1 },
            clock = { now },
        )
    }

    private fun itemJson(id: Long, read: Boolean = false): String {
        val readField = if (read) ",\"read_at\":5" else ""
        return """{"id":$id,"title":"알림 $id","body":"","caller_name":"infra-ci","caller_kind":"key","requester":"",""" +
            """"devices":1,"sent":1$readField,"created_at":${1_700_000_000 + id}}"""
    }

    private fun page(ids: List<Long>, unread: Int) =
        """{"items":[${ids.joinToString(",") { itemJson(it) }}],"unread_count":$unread}"""

    @Test
    fun reloadRequestsFirstPage() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(page(listOf(3, 2), unread = 2)))

        vm.reload()

        assertEquals("/api/external-notifications?limit=50", server.takeRequest().path)
        assertEquals(listOf(3L, 2L), vm.state.value.items.map { it.id })
        assertEquals(2, vm.state.value.unreadCount)
        assertFalse(vm.state.value.hasMore)
        assertTrue(vm.state.value.hasLoaded)
    }

    @Test
    fun loadMoreAsksBeforeOldestAndSkipsDuplicates() = runTest {
        val vm = viewModel()
        val first = (100L downTo 51L).toList()
        server.enqueue(MockResponse().setBody(page(first, unread = 0)))
        vm.reload()
        server.takeRequest()
        assertTrue(vm.state.value.hasMore)
        server.enqueue(MockResponse().setBody(page(listOf(51, 50, 49), unread = 0)))

        vm.performLoadMore()

        assertEquals("/api/external-notifications?limit=50&before=51", server.takeRequest().path)
        assertEquals(52, vm.state.value.items.size)
        assertFalse(vm.state.value.hasMore)
    }

    @Test
    fun markReadPostsAndDecrementsUnread() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(page(listOf(3), unread = 1)))
        vm.reload()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(204))

        vm.markRead(vm.state.value.items.single())

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/external-notifications/3/read", post.path)
        assertEquals(0, vm.state.value.unreadCount)
        assertEquals(now / 1000, vm.state.value.items.single().readAt)
        assertEquals(1, markedRead)
    }

    @Test
    fun markReadSkipsAlreadyRead() = runTest {
        val vm = viewModel()

        vm.markRead(ExternalNotification(id = 1, title = "t", readAt = 3, createdAt = 1))

        assertEquals(0, server.requestCount)
        assertEquals(0, markedRead)
    }

    @Test
    fun notificationFetchesWhenNotLoaded() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody(itemJson(42, read = true)))

        val item = vm.notification(42)

        assertEquals("/api/external-notifications/42", server.takeRequest().path)
        assertEquals(42L, item?.id)
        assertFalse(item!!.isUnread)
    }

    @Test
    fun notificationFailureReturnsNull() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(404).setBody("""{"error":"not found"}"""))

        assertNull(vm.notification(9))
        assertEquals("""HTTP 404: {"error":"not found"}""", vm.state.value.errorMessage)
    }

    @Test
    fun formattingMatchesIos() {
        val zone = ZoneId.of("Asia/Seoul")
        val created = ZonedDateTime.of(2026, 10, 27, 9, 41, 12, 0, zone).toEpochSecond()
        val item = ExternalNotification(
            id = 1,
            title = "t",
            callerName = "homelab-k3s-mcp",
            callerKind = "k8s_sa",
            requester = "",
            devices = 0,
            createdAt = created,
        )
        assertEquals("homelab-k3s-mcp · SA", ExternalNotificationFormatting.callerLabel(item))
        assertEquals("infra-ci · 키", ExternalNotificationFormatting.callerLabel(item.copy(callerName = "infra-ci", callerKind = "key")))
        assertEquals("요청 주체 미상", ExternalNotificationFormatting.requesterLabel(item))
        assertEquals("요청 nightly", ExternalNotificationFormatting.requesterLabel(item.copy(requester = "nightly")))
        assertEquals("등록된 기기 없음", ExternalNotificationFormatting.deliveryLabel(item))
        assertEquals("1/2 기기", ExternalNotificationFormatting.deliveryLabel(item.copy(devices = 2, sent = 1)))
        assertEquals("2026-10-27 09:41:12", ExternalNotificationFormatting.fullTime(item, zone))
        assertEquals("09:41", ExternalNotificationFormatting.time(item, created * 1000 + 60_000, zone))
    }
}
