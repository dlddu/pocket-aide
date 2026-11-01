package com.dlddu.pocketaide.prmonitor

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

class PrMonitorViewModelTest {
    private val server = MockWebServer()
    private val now = 1_700_000_000_000L

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(): PrMonitorViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return PrMonitorViewModel(ApiClient(server.url("/").toString(), store, clock = { now }), clock = { now })
    }

    private fun itemJson(id: Long, pr: Int? = 42, acked: Long? = null): String {
        val prPart = pr?.let { ""","pr_number":$it,"pr_title":"제목","pr_url":"https://github.com/a/b/pull/$it"""" } ?: ""
        val ackPart = acked?.let { ""","acknowledged_at":$it""" } ?: ""
        return """{"id":$id,"repo_full_name":"a/b"$prPart,"workflow_name":"ci","head_branch":"feat","head_sha":"abc$id","conclusion":"success","created_at":${100 + id}$ackPart}"""
    }

    private suspend fun loaded(vararg items: String): PrMonitorViewModel {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${items.joinToString(",")}]}"""))
        vm.reload()
        server.takeRequest()
        return vm
    }

    @Test
    fun reloadFetchesHistoryAndGroups() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${itemJson(1)},${itemJson(2)},${itemJson(3, pr = null, acked = 5)}]}"""))

        vm.reload()

        val request = server.takeRequest()
        assertEquals("GET", request.method)
        assertEquals("/api/notification-history", request.path)
        assertEquals("Bearer access", request.getHeader("Authorization"))
        val state = vm.state.value
        assertEquals(2, state.totalUnacknowledgedCount)
        assertEquals(listOf("pr:a/b:42"), state.unacknowledgedGroups.map { it.id })
        assertEquals(listOf("sha:a/b:abc3"), state.acknowledgedGroups.map { it.id })
        assertNull(state.errorMessage)
    }

    @Test
    fun reloadFailureKeepsItemsAndSetsError() = runTest {
        val vm = loaded(itemJson(1))
        server.enqueue(MockResponse().setResponseCode(500).setBody("boom"))

        vm.reload()

        assertEquals(1, vm.state.value.items.size)
        assertNotNull(vm.state.value.errorMessage)
    }

    @Test
    fun acknowledgeStampsAndPostsAck() = runTest {
        val vm = loaded(itemJson(1))
        server.enqueue(MockResponse().setResponseCode(204))

        vm.performAcknowledge(1)

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/notification-history/1/ack", post.path)
        assertEquals("{}", post.body.readUtf8())
        assertEquals(now / 1000, vm.state.value.items.single().acknowledgedAt)
        assertEquals(0, vm.state.value.totalUnacknowledgedCount)
    }

    @Test
    fun acknowledgeFailureReverts() = runTest {
        val vm = loaded(itemJson(1))
        server.enqueue(MockResponse().setResponseCode(500))

        vm.performAcknowledge(1)

        assertNull(vm.state.value.items.single().acknowledgedAt)
        assertNotNull(vm.state.value.errorMessage)
    }

    @Test
    fun acknowledgeSkipsAlreadyAcknowledged() = runTest {
        val vm = loaded(itemJson(1, acked = 9))

        vm.performAcknowledge(1)

        assertEquals(1, server.requestCount)
    }

    @Test
    fun acknowledgeGroupAcksEveryUnacknowledgedItem() = runTest {
        val vm = loaded(itemJson(1), itemJson(2), itemJson(3, acked = 9))
        server.enqueue(MockResponse().setResponseCode(204))
        server.enqueue(MockResponse().setResponseCode(204))

        vm.performAcknowledgeGroup(vm.state.value.groups.single())

        val paths = listOf(server.takeRequest().path, server.takeRequest().path).toSet()
        assertEquals(setOf("/api/notification-history/1/ack", "/api/notification-history/2/ack"), paths)
        assertTrue(vm.state.value.groups.single().allAcknowledged)
    }

    @Test
    fun settingsPatchSendsOnlyChangedField() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"enabled":false,"outcomes":"both"}"""))

        vm.performSetNotificationsEnabled(false)

        val patch = server.takeRequest()
        assertEquals("PATCH", patch.method)
        assertEquals("/api/notification-settings", patch.path)
        assertEquals("""{"enabled":false}""", patch.body.readUtf8())
        assertEquals(false, vm.state.value.notificationSettings.enabled)
    }

    @Test
    fun outcomesPatchFailureReverts() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(400).setBody("invalid"))

        vm.performSetNotificationOutcomes(NotificationOutcomes.FAILURE)

        assertEquals("""{"outcomes":"failure"}""", server.takeRequest().body.readUtf8())
        assertEquals(NotificationOutcomes.BOTH, vm.state.value.notificationSettings.outcomes)
        assertNotNull(vm.state.value.settingsError)
    }

    @Test
    fun loadSettingsDecodesOutcomes() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"enabled":true,"outcomes":"success"}"""))

        vm.reloadNotificationSettings()

        assertEquals("/api/notification-settings", server.takeRequest().path)
        assertEquals(NotificationOutcomes.SUCCESS, vm.state.value.notificationSettings.outcomes)
    }

    @Test
    fun excludeRepoPrependsCreated() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[{"id":1,"repo_full_name":"a/old","created_at":1}]}"""))
        vm.reloadExcludedRepos()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(201).setBody("""{"id":2,"repo_full_name":"a/new","created_at":2}"""))

        vm.performExcludeRepo("a/new")

        val post = server.takeRequest()
        assertEquals("/api/excluded-repos", post.path)
        assertEquals("""{"repo_full_name":"a/new"}""", post.body.readUtf8())
        assertEquals(listOf(2L, 1L), vm.state.value.excludedRepos.map { it.id })
    }

    @Test
    fun removeExcludedRepoRevertsOnFailure() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[{"id":1,"repo_full_name":"a/old","created_at":1}]}"""))
        vm.reloadExcludedRepos()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(500))

        vm.performRemoveExcludedRepo(1)

        val delete = server.takeRequest()
        assertEquals("DELETE", delete.method)
        assertEquals("/api/excluded-repos/1", delete.path)
        assertEquals(listOf(1L), vm.state.value.excludedRepos.map { it.id })
        assertNotNull(vm.state.value.excludedRepoError)
    }
}
