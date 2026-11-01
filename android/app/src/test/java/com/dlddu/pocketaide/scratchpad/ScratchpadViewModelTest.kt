package com.dlddu.pocketaide.scratchpad

import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.affirmations.AffirmationPriority
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

class ScratchpadViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(): ScratchpadViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return ScratchpadViewModel(ApiClient(server.url("/").toString(), store, clock = { now }))
    }

    private fun itemJson(id: Long) = """{"id":$id,"text":"메모 $id","source":"text","captured_at":${1_700_000_000 + id},"created_at":1}"""

    @Test
    fun reloadUpdatesUnclassifiedCount() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${itemJson(1)},${itemJson(2)}]}"""))

        vm.reload()

        assertEquals("/api/scratchpad", server.takeRequest().path)
        assertEquals(2, vm.unclassifiedCount.value)
    }

    @Test
    fun addPostsTextSourceAndPrepends() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${itemJson(1)}]}"""))
        vm.reload()
        server.takeRequest()
        server.enqueue(MockResponse().setBody(itemJson(2)))

        vm.performAdd("메모 2")

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/scratchpad", post.path)
        assertEquals("""{"text":"메모 2","source":"text"}""", post.body.readUtf8())
        assertEquals(listOf(2L, 1L), vm.state.value.items.map { it.id })
        assertEquals(2, vm.unclassifiedCount.value)
    }

    @Test
    fun moveToAffirmationReturnsTheCreatedAffirmation() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${itemJson(1)}]}"""))
        vm.reload()
        server.takeRequest()
        server.enqueue(
            MockResponse().setBody(
                """{"target":"affirmation","affirmation":{"id":9,"text":"메모 1","priority":"normal","created_at":1,"updated_at":1}}""",
            ),
        )

        val affirmation = vm.performMove(vm.state.value.items.single(), ScratchpadMoveTarget.AFFIRMATION)

        val post = server.takeRequest()
        assertEquals("/api/scratchpad/1/move", post.path)
        assertEquals("""{"target":"affirmation"}""", post.body.readUtf8())
        assertEquals(9L, affirmation?.id)
        assertEquals(0, vm.unclassifiedCount.value)
    }

    @Test
    fun moveToTodoReturnsNoAffirmation() = runTest {
        val vm = viewModel()
        server.enqueue(
            MockResponse().setBody(
                """{"target":"personal","todo":{"id":3,"title":"메모 1","memo":"","created_at":1,"updated_at":1}}""",
            ),
        )

        assertNull(vm.performMove(ScratchpadItem(id = 1, text = "메모 1", capturedAt = 1), ScratchpadMoveTarget.PERSONAL))
        assertEquals("/api/scratchpad/1/move", server.takeRequest().path)
    }

    @Test
    fun setPriorityPatchesAffirmation() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"id":9,"text":"매일 1%","priority":"high","created_at":1,"updated_at":2}"""))

        vm.performSetPriority(
            Affirmation(id = 9, text = "매일", priority = AffirmationPriority.NORMAL, createdAt = 1, updatedAt = 1),
            "매일 1%",
            AffirmationPriority.HIGH,
        )

        val patch = server.takeRequest()
        assertEquals("PATCH", patch.method)
        assertEquals("/api/affirmations/9", patch.path)
        assertEquals("""{"text":"매일 1%","priority":"high"}""", patch.body.readUtf8())
        assertNull(vm.state.value.errorMessage)
    }

    @Test
    fun deleteRemovesItem() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${itemJson(1)}]}"""))
        vm.reload()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(204))

        vm.performDelete(1)

        val delete = server.takeRequest()
        assertEquals("DELETE", delete.method)
        assertEquals("/api/scratchpad/1", delete.path)
        assertEquals(0, vm.unclassifiedCount.value)
    }
}
