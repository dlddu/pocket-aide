package com.dlddu.pocketaide.commitstatus

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.approvals.ApprovalDecision
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

class CommitStatusViewModelTest {
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

    private fun requestJson(id: String, status: String = "PENDING", createdAt: Long = 100, lastError: String = "") =
        """{"id":"$id","status":"$status","owner":"dlddu","repo":"pocket-aide",""" +
            """"sha":"0123456789abcdef0123456789abcdef01234567","state":"success","context":"review/manual-approval",""" +
            """"description":"수동 승인","target_url":"","requester":"Claude","reason":"","session_id":"","session_name":"",""" +
            """"key_name":"homelab-k3s-mcp","key_revoked":false,"github_status_id":null,"last_error":"$lastError",""" +
            """"created_at":$createdAt,"processed_at":null}"""

    @Test
    fun reloadListsPendingNewestFirstWithHistory() = runTest {
        val vm = CommitStatusViewModel(api())
        server.enqueue(
            MockResponse().setBody(
                """{"items":[${requestJson("a", createdAt = 1)},${requestJson("b", createdAt = 2)}],""" +
                    """"history":[${requestJson("h", status = "APPROVED")}]}""",
            ),
        )

        vm.reload()

        assertEquals("/api/commit-status-requests", server.takeRequest().path)
        val state = vm.state.value
        assertEquals(listOf("b", "a"), state.items.map { it.id })
        assertEquals(listOf("h"), state.history.map { it.id })
        assertEquals(2, vm.pendingCount.value)
        assertFalse(state.isLoading)
    }

    @Test
    fun approveWritesThroughTheBackend() = runTest {
        val vm = CommitStatusDetailViewModel(CommitStatusRequest(id = "a", owner = "dlddu", repo = "r", sha = "s", state = "success", context = "review/x", createdAt = 1), api())
        server.enqueue(MockResponse().setBody(requestJson("a", status = "APPROVED")))

        vm.performDecide(ApprovalDecision.APPROVE)

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/commit-status-requests/a/approve", post.path)
        assertEquals(CommitStatusRequestStatus.APPROVED, vm.state.value.item.status)
        assertNull(vm.state.value.notice)
    }

    @Test
    fun githubFailureKeepsRequestPendingWithItsMessage() = runTest {
        val vm = CommitStatusDetailViewModel(CommitStatusRequest(id = "a", owner = "dlddu", repo = "r", sha = "s", state = "success", context = "review/x", createdAt = 1), api())
        server.enqueue(
            MockResponse().setResponseCode(502)
                .setBody("""{"error":"GitHub answered 422: No commit","request":${requestJson("a", lastError = "GitHub answered 422: No commit")}}"""),
        )

        vm.performDecide(ApprovalDecision.APPROVE)

        assertEquals(CommitStatusRequestStatus.PENDING, vm.state.value.item.status)
        assertEquals("GitHub 에 기록하지 못했습니다 — GitHub answered 422: No commit", vm.state.value.notice)
        assertEquals(true, vm.state.value.canDecide)
    }

    @Test
    fun conflictAndUnavailableAreExplained() = runTest {
        val vm = CommitStatusDetailViewModel(CommitStatusRequest(id = "a", owner = "dlddu", repo = "r", sha = "s", state = "success", context = "review/x", createdAt = 1), api())
        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"already processed","request":${requestJson("a", status = "REJECTED")}}"""))
        vm.performDecide(ApprovalDecision.APPROVE)
        assertEquals("이미 처리된 요청입니다 — 거절됨", vm.state.value.notice)

        val other = CommitStatusDetailViewModel(CommitStatusRequest(id = "b", owner = "dlddu", repo = "r", sha = "s", state = "success", context = "review/x", createdAt = 1), api())
        server.enqueue(MockResponse().setResponseCode(503).setBody("""{"error":"commit status writes are not configured","request":${requestJson("b")}}"""))
        other.performDecide(ApprovalDecision.APPROVE)
        assertEquals(CommitStatusFormatting.UNAVAILABLE_NOTICE, other.state.value.notice)
        assertEquals(CommitStatusRequestStatus.PENDING, other.state.value.item.status)
    }

    @Test
    fun notificationActionsMapToDecisions() {
        assertEquals(ApprovalDecision.APPROVE, CommitStatusNotificationAction.decision(CommitStatusNotificationAction.APPROVE_ACTION))
        assertEquals(ApprovalDecision.REJECT, CommitStatusNotificationAction.decision(CommitStatusNotificationAction.REJECT_ACTION))
        assertNull(CommitStatusNotificationAction.decision("APPROVAL_APPROVE"))
        assertEquals(CommitStatusFormatting.FAILED_DECISION_NOTICE, CommitStatusNotificationAction.resultNotice(Result.failure(RuntimeException())))
    }
}
