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
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

class ApprovalKeysViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(): ApprovalKeysViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ApprovalKeysViewModel(ApiClient(server.url("/").toString(), store, clock = { now }))
    }

    private fun keyJson(id: Long, name: String, extra: String = "") =
        """{"id":$id,"name":"$name","kind":"key","key_prefix":"pa_$id","service_account":null,"created_at":$id,""" +
            """"last_used_at":null,"revoked_at":null$extra}"""

    private fun enqueueReload(vararg keys: String) {
        server.enqueue(MockResponse().setBody("""{"keys":[${keys.joinToString(",")}]}"""))
        server.enqueue(MockResponse().setBody("""{"auto_response":"NONE"}"""))
    }

    @Test
    fun reloadReadsKeysAndMode() = runTest {
        val vm = viewModel()
        enqueueReload(keyJson(1, "old"), keyJson(2, "new"))

        vm.reload()

        assertEquals("/api/approval-keys", server.takeRequest().path)
        assertEquals("/api/approval-settings", server.takeRequest().path)
        assertEquals(listOf("new", "old"), vm.state.value.keys.map { it.name })
        assertEquals(ApprovalAutoResponse.OFF, vm.state.value.autoResponse)
    }

    @Test
    fun issueShowsSecretOnceAndReloads() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(201).setBody(keyJson(3, "infra-ci", ""","key":"pa_secret"""")))
        enqueueReload(keyJson(3, "infra-ci"))

        vm.performIssue("  infra-ci ")

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/approval-keys", post.path)
        assertEquals("""{"name":"infra-ci"}""", post.body.readUtf8())
        assertEquals("pa_secret", vm.state.value.issued?.secret)
        assertEquals("infra-ci", vm.state.value.issued?.key?.name)
        assertEquals(1, vm.state.value.keys.size)

        vm.finishIssuing()
        assertNull(vm.state.value.issued)
    }

    @Test
    fun issueDuplicateNameShowsNotice() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"caller name already in use"}"""))

        vm.performIssue("infra-ci")

        assertEquals(ApprovalKeyFormatting.DUPLICATE_NAME_NOTICE, vm.state.value.issueNotice)
        assertNull(vm.state.value.issued)
    }

    @Test
    fun registerRejectsInvalidServiceAccountLocally() = runTest {
        val vm = viewModel()

        assertFalse(vm.performRegisterServiceAccount("mcp", "Homelab", "mcp"))

        assertEquals(0, server.requestCount)
        assertEquals(ApprovalServiceAccountOutcome.InvalidServiceAccount, vm.state.value.registerOutcome)
        assertEquals(ApprovalKeyFormatting.INVALID_SERVICE_ACCOUNT_NOTICE, vm.state.value.registerNotice)
    }

    @Test
    fun registerPostsServiceAccountKind() = runTest {
        val vm = viewModel()
        server.enqueue(
            MockResponse().setResponseCode(201).setBody(
                """{"id":4,"name":"mcp","kind":"k8s_sa","key_prefix":"","service_account":"homelab/mcp","created_at":4}""",
            ),
        )
        enqueueReload()

        assertTrue(vm.performRegisterServiceAccount("mcp", "homelab", "mcp"))

        val post = server.takeRequest()
        assertEquals("""{"name":"mcp","kind":"k8s_sa","service_account":"homelab/mcp"}""", post.body.readUtf8())
        assertNull(vm.state.value.registerNotice)
        assertFalse(vm.state.value.isRegistering)
    }

    @Test
    fun registerDuplicateServiceAccountIsDistinguished() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"service account already registered"}"""))

        assertFalse(vm.performRegisterServiceAccount("mcp", "homelab", "mcp"))

        assertEquals(ApprovalServiceAccountOutcome.DuplicateServiceAccount, vm.state.value.registerOutcome)
        assertEquals(ApprovalKeyFormatting.DUPLICATE_SERVICE_ACCOUNT_NOTICE, vm.state.value.registerNotice)
    }

    @Test
    fun registerDuplicateNameIsDistinguished() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(409).setBody("""{"error":"caller name already in use"}"""))

        vm.performRegisterServiceAccount("mcp", "homelab", "mcp")

        assertEquals(ApprovalKeyFormatting.SERVICE_ACCOUNT_DUPLICATE_NAME_NOTICE, vm.state.value.registerNotice)
    }

    @Test
    fun setAutoResponseSkipsCurrentMode() = runTest {
        val vm = viewModel()
        enqueueReload()
        vm.reload()
        server.takeRequest()
        server.takeRequest()

        vm.performSetAutoResponse(ApprovalAutoResponse.OFF)
        assertEquals(2, server.requestCount)

        server.enqueue(MockResponse().setBody("""{"auto_response":"AUTO_REJECT"}"""))
        vm.performSetAutoResponse(ApprovalAutoResponse.AUTO_REJECT)

        val patch = server.takeRequest()
        assertEquals("""{"auto_response":"AUTO_REJECT"}""", patch.body.readUtf8())
        assertEquals(ApprovalAutoResponse.AUTO_REJECT, vm.state.value.autoResponse)
    }

    @Test
    fun revokePostsAndReloads() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"id":7,"name":"infra-ci","key_prefix":"pa_7","created_at":7,"revoked_at":9}"""))
        enqueueReload(keyJson(7, "infra-ci"))

        vm.performRevoke(ApprovalCallerKey(id = 7, name = "infra-ci", createdAt = 7))

        val post = server.takeRequest()
        assertEquals("/api/approval-keys/7/revoke", post.path)
        assertEquals("{}", post.body.readUtf8())
        assertEquals("/api/approval-keys", server.takeRequest().path)
    }
}
