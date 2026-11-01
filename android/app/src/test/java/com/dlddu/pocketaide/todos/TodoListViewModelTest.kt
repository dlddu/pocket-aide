package com.dlddu.pocketaide.todos

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.InMemoryTokenStore
import com.dlddu.pocketaide.auth.TokenBundle
import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Before
import org.junit.Test

class TodoListViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(area: TodoArea = TodoArea.WORK): TodoListViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return TodoListViewModel(area, ApiClient(server.url("/").toString(), store, clock = { now }))
    }

    private val itemJson =
        """{"id":5,"title":"메일","memo":"","due_date":null,"priority":null,"completed_at":null,"created_at":1,"updated_at":1}"""

    @Test
    fun reloadFetchesOnlyTheBoundArea() = runTest {
        val vm = viewModel(TodoArea.WORK)
        server.enqueue(MockResponse().setBody("""{"items":[$itemJson]}"""))

        vm.reload()

        val request = server.takeRequest()
        assertEquals("GET", request.method)
        assertEquals("/api/todos/work", request.path)
        assertEquals(listOf(5L), vm.state.value.items.map { it.id })
        assertFalse(vm.state.value.isLoading)
        assertNull(vm.state.value.errorMessage)
    }

    @Test
    fun toggleDonePatchesTheFullDraftThenReloads() = runTest {
        val vm = viewModel(TodoArea.PERSONAL)
        server.enqueue(MockResponse().setBody(itemJson))
        server.enqueue(MockResponse().setBody("""{"items":[]}"""))

        vm.performToggleDone(TodoItem(id = 5, title = "메일", memo = "m", priority = TodoPriority.HIGH))

        val patch = server.takeRequest()
        assertEquals("PATCH", patch.method)
        assertEquals("/api/todos/personal/5", patch.path)
        assertEquals("""{"title":"메일","memo":"m","priority":"high","done":true}""", patch.body.readUtf8())
        assertEquals("/api/todos/personal", server.takeRequest().path)
    }

    @Test
    fun createPostsDraftThenReloads() = runTest {
        val vm = viewModel(TodoArea.WORK)
        server.enqueue(MockResponse().setBody(itemJson))
        server.enqueue(MockResponse().setBody("""{"items":[$itemJson]}"""))

        vm.performAdd(TodoDraft(title = "메일", dueDate = "2026-10-14"))

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/todos/work", post.path)
        assertEquals("""{"title":"메일","memo":"","due_date":"2026-10-14","done":false}""", post.body.readUtf8())
        assertEquals(1, vm.state.value.items.size)
    }

    @Test
    fun deleteRemovesItemLocally() = runTest {
        val vm = viewModel(TodoArea.WORK)
        server.enqueue(MockResponse().setBody("""{"items":[$itemJson]}"""))
        vm.reload()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(204))

        vm.performDelete(5)

        val delete = server.takeRequest()
        assertEquals("DELETE", delete.method)
        assertEquals("/api/todos/work/5", delete.path)
        assertEquals(emptyList<TodoItem>(), vm.state.value.items)
    }

    @Test
    fun failureSurfacesErrorMessage() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(500).setBody("boom"))

        vm.reload()

        assertNotNull(vm.state.value.errorMessage)
        assertEquals("HTTP 500: boom", vm.state.value.errorMessage)
    }
}
