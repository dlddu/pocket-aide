package com.dlddu.pocketaide.affirmations

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
import org.junit.Before
import org.junit.Test

class AffirmationsViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(seed: ULong = 2009uL): AffirmationsViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return AffirmationsViewModel(ApiClient(server.url("/").toString(), store, clock = { now }), SeededRng(seed))
    }

    private fun json(id: Long, priority: String = "normal", createdAt: Long = id) =
        """{"id":$id,"text":"다짐 $id","priority":"$priority","created_at":$createdAt,"updated_at":$createdAt}"""

    @Test
    fun reloadListsAndPicksAHero() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${json(1, "high")},${json(2, "low")}]}"""))

        vm.reload()

        val request = server.takeRequest()
        assertEquals("GET", request.method)
        assertEquals("/api/affirmations", request.path)
        val state = vm.state.value
        assertEquals(listOf(1L, 2L), state.items.map { it.id })
        assertNotNull(state.heroId)
        assertEquals(state.heroId, state.heroItem?.id)
        assertNull(state.errorMessage)
    }

    @Test
    fun seededRotationIsReproducible() = runTest {
        val body = """{"items":[${json(1, "high")},${json(2, "normal")},${json(3, "low")}]}"""
        val first = viewModel(seed = 77uL)
        server.enqueue(MockResponse().setBody(body))
        first.reload()
        val second = viewModel(seed = 77uL)
        server.enqueue(MockResponse().setBody(body))
        second.reload()

        assertEquals(first.state.value.heroId, second.state.value.heroId)
    }

    @Test
    fun addPostsPayloadPrependsAndShowsItAsHero() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${json(1)}]}"""))
        vm.reload()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(201).setBody(json(2, "high")))

        vm.performAdd("다짐 2", AffirmationPriority.HIGH)

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/affirmations", post.path)
        assertEquals("""{"text":"다짐 2","priority":"high"}""", post.body.readUtf8())
        assertEquals(listOf(2L, 1L), vm.state.value.items.map { it.id })
        assertEquals(2L, vm.state.value.heroId)
    }

    @Test
    fun updateReplacesTheItemInPlace() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${json(1)},${json(2)}]}"""))
        vm.reload()
        server.takeRequest()
        server.enqueue(MockResponse().setBody("""{"id":2,"text":"고친 다짐","priority":"low","created_at":2,"updated_at":9}"""))

        vm.performUpdate(2, "고친 다짐", AffirmationPriority.LOW)

        val patch = server.takeRequest()
        assertEquals("PATCH", patch.method)
        assertEquals("/api/affirmations/2", patch.path)
        assertEquals("""{"text":"고친 다짐","priority":"low"}""", patch.body.readUtf8())
        assertEquals(listOf("다짐 1", "고친 다짐"), vm.state.value.items.map { it.text })
        assertEquals(2L, vm.state.value.heroId)
    }

    @Test
    fun deletingTheHeroRotatesToARemainingItem() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setBody("""{"items":[${json(1)}]}"""))
        vm.reload()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(201).setBody(json(2)))
        vm.performAdd("다짐 2", AffirmationPriority.NORMAL)
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(204))

        vm.performDelete(2)

        val delete = server.takeRequest()
        assertEquals("DELETE", delete.method)
        assertEquals("/api/affirmations/2", delete.path)
        assertEquals(listOf(1L), vm.state.value.items.map { it.id })
        assertEquals(1L, vm.state.value.heroId)
    }

    @Test
    fun failedLoadKeepsItemsAndRecordsError() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(500).setBody("boom"))

        vm.reload()

        assertEquals(emptyList<Affirmation>(), vm.state.value.items)
        assertNotNull(vm.state.value.errorMessage)
        assertNull(vm.state.value.heroItem)
    }

    @Test
    fun heroFallsBackToFirstItemWhenUnset() {
        val items = listOf(
            Affirmation(id = 1, text = "a", priority = AffirmationPriority.LOW, createdAt = 1, updatedAt = 1),
            Affirmation(id = 2, text = "b", priority = AffirmationPriority.HIGH, createdAt = 2, updatedAt = 2),
        )
        assertEquals(1L, AffirmationsUiState(items = items).heroItem?.id)
        assertEquals(1L, AffirmationsUiState(items = items, heroId = 99).heroItem?.id)
        assertEquals(listOf(2L, 1L), AffirmationsUiState(items = items).sortedItems.map { it.id })
    }
}
