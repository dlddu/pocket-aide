package com.dlddu.pocketaide.approvals

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
import org.junit.Before
import org.junit.Test

class ApprovalsViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun api(): ApiClient {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ApiClient(server.url("/").toString(), store, clock = { now })
    }

    private fun approvalJson(id: String, status: String = "PENDING", createdAt: Long = 100) =
        """{"id":"$id","status":"$status","external_id":"ext-$id","context":"요청 $id","requester_name":"ci",""" +
            """"key_name":"infra-ci","key_revoked":false,"created_at":$createdAt,"expires_at":null,"processed_at":null}"""

    @Test
    fun reloadListsPendingNewestFirstAndReadsSettings() = runTest {
        val vm = ApprovalsViewModel(api())
        server.enqueue(MockResponse().setBody("""{"items":[${approvalJson("a", createdAt = 1)},${approvalJson("b", createdAt = 2)}]}"""))
        server.enqueue(MockResponse().setBody("""{"auto_response":"AUTO_APPROVE"}"""))

        vm.reload()

        assertEquals("/api/approvals", server.takeRequest().path)
        assertEquals("/api/approval-settings", server.takeRequest().path)
        val state = vm.state.value
        assertEquals(listOf("b", "a"), state.items.map { it.id })
        assertEquals(ApprovalAutoResponse.AUTO_APPROVE, state.autoResponse)
        assertEquals(2, vm.pendingCount.value)
        assertEquals(true, state.hasLoaded)
        assertFalse(state.isLoading)
    }

    @Test
    fun reloadFailureKeepsItemsAndShowsError() = runTest {
        val vm = ApprovalsViewModel(api())
        server.enqueue(MockResponse().setResponseCode(503).setBody("service unavailable"))

        vm.reload()

        assertEquals("HTTP 503: service unavailable", vm.state.value.errorMessage)
        assertFalse(vm.state.value.hasLoaded)
    }

    @Test
    fun turnOffAutoResponsePatchesNone() = runTest {
        val vm = ApprovalsViewModel(api())
        server.enqueue(MockResponse().setBody("""{"auto_response":"NONE"}"""))

        vm.performTurnOffAutoResponse()

        val patch = server.takeRequest()
        assertEquals("PATCH", patch.method)
        assertEquals("/api/approval-settings", patch.path)
        assertEquals("""{"auto_response":"NONE"}""", patch.body.readUtf8())
        assertEquals(ApprovalAutoResponse.OFF, vm.state.value.autoResponse)
    }

    @Test
    fun openFetchesApprovalAndNavigatesToDetail() = runTest {
        val vm = ApprovalsViewModel(api())
        vm.navigate(ApprovalDestination.Settings)
        server.enqueue(MockResponse().setBody(approvalJson("apr_9")))

        vm.performOpen("apr_9")

        assertEquals("/api/approvals/apr_9", server.takeRequest().path)
        val destination = vm.state.value.destination as ApprovalDestination.Detail
        assertEquals("apr_9", destination.item.id)
    }

    @Test
    fun openIgnoresMissingApproval() = runTest {
        val vm = ApprovalsViewModel(api())
        server.enqueue(MockResponse().setResponseCode(404).setBody("""{"error":"not found"}"""))

        vm.performOpen("gone")

        assertEquals(ApprovalDestination.Pending, vm.state.value.destination)
    }

    @Test
    fun decidePostsEmptyBodyAndUpdatesItem() = runTest {
        val vm = ApprovalDetailViewModel(ApprovalsJson.decode(approvalJson("a")), api())
        server.enqueue(MockResponse().setBody(approvalJson("a", status = "APPROVED")))

        vm.performDecide(ApprovalDecision.APPROVE)

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/approvals/a/approve", post.path)
        assertEquals("{}", post.body.readUtf8())
        assertEquals(ApprovalStatus.APPROVED, vm.state.value.item.status)
        assertNull(vm.state.value.notice)
        assertFalse(vm.state.value.isDeciding)
    }

    @Test
    fun decideConflictShowsCurrentState() = runTest {
        val vm = ApprovalDetailViewModel(ApprovalsJson.decode(approvalJson("a")), api())
        server.enqueue(
            MockResponse().setResponseCode(409).setBody("""{"error":"expired","approval":${approvalJson("a", status = "EXPIRED")}}"""),
        )

        vm.performDecide(ApprovalDecision.REJECT)

        assertEquals("/api/approvals/a/reject", server.takeRequest().path)
        assertEquals(ApprovalStatus.EXPIRED, vm.state.value.item.status)
        assertEquals("만료된 요청이라 결정되지 않았습니다", vm.state.value.notice)
    }

    @Test
    fun decideFailureShowsRetryNotice() = runTest {
        val vm = ApprovalDetailViewModel(ApprovalsJson.decode(approvalJson("a")), api())
        server.enqueue(MockResponse().setResponseCode(500).setBody("""{"error":"decide failed"}"""))

        vm.performDecide(ApprovalDecision.APPROVE)

        assertEquals(ApprovalFormatting.FAILED_DECISION_NOTICE, vm.state.value.notice)
        assertEquals(ApprovalStatus.PENDING, vm.state.value.item.status)
    }

    @Test
    fun detailReloadFailureSetsLoadError() = runTest {
        val vm = ApprovalDetailViewModel(ApprovalsJson.decode(approvalJson("a")), api())
        server.enqueue(MockResponse().setResponseCode(500).setBody("boom"))

        vm.performReload()

        assertEquals("/api/approvals/a", server.takeRequest().path)
        assertEquals("HTTP 500: boom", vm.state.value.loadError)
    }

    @Test
    fun historyPagesWithEncodedCursor() = runTest {
        val vm = ApprovalHistoryViewModel(api())
        server.enqueue(
            MockResponse().setBody("""{"items":[${approvalJson("a", status = "APPROVED")}],"next_cursor":"10:a+b"}"""),
        )
        server.enqueue(MockResponse().setBody("""{"items":[${approvalJson("b", status = "REJECTED")}],"next_cursor":null}"""))

        vm.reload()
        vm.performLoadMore()

        assertEquals("/api/approvals/history", server.takeRequest().path)
        assertEquals("/api/approvals/history?cursor=10%3Aa%2Bb", server.takeRequest().path)
        assertEquals(listOf("a", "b"), vm.state.value.items.map { it.id })
        assertFalse(vm.state.value.hasMore)
    }

    private object ApprovalsJson {
        fun decode(raw: String): ApprovalRequest = ApiClient.json.decodeFromString(ApprovalRequest.serializer(), raw)
    }
}
