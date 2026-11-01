package com.dlddu.pocketaide.prmonitor

import kotlinx.coroutines.test.runTest
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Before
import org.junit.Test
import java.time.Instant

class OpenPullRequestsTest {
    private val server = MockWebServer()

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun node(repo: String, number: Int, updated: String, author: String? = "me", rollup: String? = "SUCCESS"): String {
        val authorPart = author?.let { """"author":{"login":"$it"},""" } ?: """"author":null,"""
        val rollupPart = rollup?.let { """{"state":"$it"}""" } ?: "null"
        return """{"number":$number,"title":"PR $number","url":"https://github.com/$repo/pull/$number","updatedAt":"$updated","repository":{"nameWithOwner":"$repo"},$authorPart"commits":{"nodes":[{"commit":{"statusCheckRollup":$rollupPart}}]}}"""
    }

    private fun searchBody(authored: List<String>, requested: List<String> = emptyList(), reviewed: List<String> = emptyList()) =
        """{"data":{"authored":{"nodes":[${authored.joinToString(",")}]},"requested":{"nodes":[${requested.joinToString(",")}]},"reviewed":{"nodes":[${reviewed.joinToString(",")}]}}}"""

    @Test
    fun decodeMergesDedupesAndSorts() {
        val result = OpenPullRequests.decode(
            searchBody(
                authored = listOf(node("a/b", 1, "2026-10-01T00:00:00Z"), "null"),
                requested = listOf(node("a/c", 2, "2026-10-03T00:00:00Z", author = null, rollup = "FAILURE"), "{}"),
                reviewed = listOf(node("a/b", 1, "2026-10-01T00:00:00Z")),
            ),
        )

        assertEquals(listOf("a/c#2", "a/b#1"), result.pullRequests.map { it.id })
        assertEquals(2, result.inaccessibleCount)
        val requested = result.pullRequests.first()
        assertEquals(PullRequestRole.REVIEWER, requested.role)
        assertEquals("ghost", requested.author)
        assertEquals(PullRequestCiStatus.FAILURE, requested.ciStatus)
        assertEquals(PullRequestRole.AUTHOR, result.pullRequests[1].role)
    }

    @Test
    fun decodeRateLimitedGraphQlError() {
        try {
            OpenPullRequests.decode("""{"errors":[{"type":"RATE_LIMITED","message":"slow down"}]}""")
            fail("expected error")
        } catch (e: GitHubError) {
            assertEquals(GitHubError.RateLimited(null), e)
        }
    }

    @Test
    fun decodeGraphQlErrorMessage() {
        try {
            OpenPullRequests.decode("""{"errors":[{"message":"bad query"}]}""")
            fail("expected error")
        } catch (e: GitHubError) {
            assertEquals("GitHub 조회 오류: bad query", e.description)
        }
    }

    @Test
    fun classifyStatuses() {
        val now = Instant.ofEpochSecond(1_000)
        assertNull(GitHubError.classify(200))
        assertEquals(GitHubError.Unauthorized, GitHubError.classify(401))
        assertEquals(GitHubError.Forbidden, GitHubError.classify(403))
        assertEquals(GitHubError.RateLimited(Instant.ofEpochSecond(1_030)), GitHubError.classify(403, retryAfter = "30", now = now))
        assertEquals(GitHubError.RateLimited(Instant.ofEpochSecond(2_000)), GitHubError.classify(403, remaining = "0", reset = "2000"))
        assertEquals(GitHubError.RateLimited(null), GitHubError.classify(429))
        assertEquals(GitHubError.BadStatus(502), GitHubError.classify(502))
    }

    @Test
    fun alertFromErrors() {
        assertEquals(GitHubAlert.TokenRejected, GitHubAlert.make(GitHubError.Unauthorized, 0))
        assertEquals(GitHubAlert.InsufficientAccess(3), GitHubAlert.make(GitHubError.Forbidden, 3))
        assertEquals(GitHubAlert.InsufficientAccess(2), GitHubAlert.make(null, 2))
        assertNull(GitHubAlert.make(GitHubError.BadStatus(500), 0))
        assertEquals("ratelimit", GitHubAlert.RateLimited(null).kind)
        assertEquals("잠시 후 다시 시도하세요.", GitHubAlert.RateLimited(null).message { "x" })
        assertTrue(GitHubAlert.InsufficientAccess(2).message { "" }.startsWith("PR 2개는"))
    }

    @Test
    fun filters() {
        val base = OpenPullRequest("a/b", 1, "t", "me", null, Instant.EPOCH, PullRequestRole.AUTHOR, PullRequestCiStatus.FAILURE)
        val review = base.copy(number = 2, role = PullRequestRole.REVIEWER, ciStatus = PullRequestCiStatus.SUCCESS)
        val all = listOf(base, review)
        assertEquals(listOf(base), OpenPullRequestFilter.MINE.apply(all))
        assertEquals(listOf(review), OpenPullRequestFilter.REVIEW_REQUESTED.apply(all))
        assertEquals(listOf(base), OpenPullRequestFilter.CI_FAILING.apply(all))
        assertEquals(OpenPullRequestFilter.REVIEW_REQUESTED, OpenPullRequestFilter.fromWire("reviewRequested"))
        assertNull(OpenPullRequestFilter.fromWire("bogus"))
    }

    @Test
    fun ciStatusFromRollup() {
        assertEquals(PullRequestCiStatus.IN_PROGRESS, PullRequestCiStatus.fromRollup("EXPECTED"))
        assertEquals(PullRequestCiStatus.FAILURE, PullRequestCiStatus.fromRollup("ERROR"))
        assertEquals(PullRequestCiStatus.NO_CHECKS, PullRequestCiStatus.fromRollup(null))
    }

    private fun viewModel(
        store: GitHubCredentialStore = InMemoryGitHubCredentialStore(),
        filterStore: OpenPullRequestFilterStore = InMemoryOpenPullRequestFilterStore(),
    ) = OpenPullRequestsViewModel(
        store = store,
        client = GitHubClient(server.url("/").toString()),
        filterStore = filterStore,
        clock = { Instant.ofEpochSecond(42) },
    )

    @Test
    fun connectSavesCredentialAndLoads() = runTest {
        val store = InMemoryGitHubCredentialStore()
        val vm = viewModel(store)
        server.enqueue(MockResponse().setBody("""{"login":"octo"}"""))
        server.enqueue(MockResponse().setBody(searchBody(listOf(node("a/b", 1, "2026-10-01T00:00:00Z")))))

        assertTrue(vm.performConnect("  ghp_token  "))

        val user = server.takeRequest()
        assertEquals("/user", user.path)
        assertEquals("Bearer ghp_token", user.getHeader("Authorization"))
        assertEquals("2022-11-28", user.getHeader("X-GitHub-Api-Version"))
        val graphql = server.takeRequest()
        assertEquals("POST", graphql.method)
        assertEquals("/graphql", graphql.path)
        assertTrue(graphql.body.readUtf8().contains("author:octo"))
        assertEquals(GitHubCredential("ghp_token", "octo"), store.load())
        val state = vm.state.value
        assertEquals("octo", state.login)
        assertTrue(state.hasLoaded)
        assertEquals(Instant.ofEpochSecond(42), state.lastRefreshedAt)
        assertEquals(listOf("a/b#1"), state.visiblePullRequests.map { it.id })
    }

    @Test
    fun connectRejectedTokenShowsMessage() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(401))

        assertFalse(vm.performConnect("bad"))

        assertEquals(GitHubError.Unauthorized.description, vm.state.value.connectError)
        assertFalse(vm.state.value.isConnected)
        assertFalse(vm.state.value.isConnecting)
    }

    @Test
    fun refreshRateLimitedProducesAlert() = runTest {
        val vm = viewModel(InMemoryGitHubCredentialStore(GitHubCredential("t", "octo")))
        server.enqueue(MockResponse().setResponseCode(403).setHeader("X-RateLimit-Remaining", "0").setHeader("X-RateLimit-Reset", "2000"))

        vm.performRefresh()

        assertEquals(GitHubAlert.RateLimited(Instant.ofEpochSecond(2_000)), vm.state.value.alert)
        assertFalse(vm.state.value.hasLoaded)
    }

    @Test
    fun disconnectClearsStore() {
        val store = InMemoryGitHubCredentialStore(GitHubCredential("t", "octo"))
        val vm = viewModel(store)
        assertTrue(vm.state.value.isConnected)

        vm.disconnect()

        assertNull(store.load())
        assertFalse(vm.state.value.isConnected)
    }

    @Test
    fun filterIsPersisted() {
        val filters = InMemoryOpenPullRequestFilterStore(OpenPullRequestFilter.CI_FAILING)
        val vm = viewModel(filterStore = filters)
        assertEquals(OpenPullRequestFilter.CI_FAILING, vm.state.value.filter)

        vm.setFilter(OpenPullRequestFilter.MINE)

        assertEquals(OpenPullRequestFilter.MINE, filters.load())
    }
}
