package com.dlddu.pocketaide.routines

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
import java.time.Clock
import java.time.ZoneId
import java.time.ZonedDateTime

class RoutinesViewModelTest {
    private val server = MockWebServer()
    private val now = 1_000_000L
    private val zone: ZoneId = ZoneId.of("Asia/Seoul")
    private val clock: Clock = Clock.fixed(ZonedDateTime.of(2026, 10, 12, 23, 30, 0, 0, zone).toInstant(), zone)

    @Before
    fun setUp() = server.start()

    @After
    fun tearDown() = server.shutdown()

    private fun viewModel(): RoutinesViewModel {
        val store = InMemoryTokenStore(TokenBundle("access", "refresh", now + 3_600_000))
        return RoutinesViewModel(ApiClient(server.url("/").toString(), store, clock = { now }), clock)
    }

    private fun routineJson(id: Long, name: String = "루틴 $id", cadence: String = "daily", weekdays: Int = 0) =
        """{"id":$id,"name":"$name","cadence":"$cadence","weekdays":$weekdays,"month_day":0,"start_day":"2026-10-01","steps":[{"id":${id * 10},"title":"단계","position":0}],"created_at":1}"""

    private fun dayJson(id: Long, checked: Boolean = false) =
        """{"id":$id,"name":"루틴 $id","cadence":"daily","weekdays":0,"month_day":0,"day":"2026-10-12","steps":[{"id":${id * 10},"title":"단계","position":0,"checked":$checked}],"done":${if (checked) 1 else 0},"total":1,"completed":$checked}"""

    private fun enqueueLoad(routines: List<String>, today: List<String>) {
        server.enqueue(MockResponse().setBody("""{"items":[${routines.joinToString(",")}]}"""))
        server.enqueue(MockResponse().setBody("""{"day":"2026-10-12","items":[${today.joinToString(",")}]}"""))
    }

    @Test
    fun reloadFetchesAllRoutinesAndTodayByLocalDay() = runTest {
        val vm = viewModel()
        enqueueLoad(
            listOf(routineJson(1), routineJson(2, cadence = "weekly", weekdays = 1)),
            listOf(dayJson(1)),
        )

        vm.reload()

        assertEquals("/api/routines", server.takeRequest().path)
        assertEquals("/api/routines/days/2026-10-12", server.takeRequest().path)
        val state = vm.state.value
        assertEquals("2026-10-12", state.dayKey)
        assertEquals(listOf(1L), state.today.map { it.id })
        assertEquals(listOf(2L), state.restingToday.map { it.id })
        assertNull(state.errorMessage)
        assertFalse(state.canRetry)
    }

    @Test
    fun toggleSendsInverseCheckForTheRoutineDay() = runTest {
        val vm = viewModel()
        enqueueLoad(listOf(routineJson(1)), listOf(dayJson(1)))
        vm.reload()
        server.takeRequest()
        server.takeRequest()
        server.enqueue(MockResponse().setBody(dayJson(1, checked = true)))

        val day = vm.state.value.today.single()
        vm.performToggle(day.steps.single(), day)

        val patch = server.takeRequest()
        assertEquals("PATCH", patch.method)
        assertEquals("/api/routines/1/days/2026-10-12/steps/10", patch.path)
        assertEquals("""{"checked":true}""", patch.body.readUtf8())
        assertTrue(vm.state.value.today.single().completed)
    }

    @Test
    fun failedCheckShowsCopyAndRetriesTheSameRequest() = runTest {
        val vm = viewModel()
        enqueueLoad(listOf(routineJson(1)), listOf(dayJson(1)))
        vm.reload()
        server.takeRequest()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(503).setBody("service unavailable"))

        val day = vm.state.value.today.single()
        vm.performToggle(day.steps.single(), day)
        server.takeRequest()

        val failed = vm.state.value
        assertEquals("단계 체크를 저장하지 못했습니다. 서버가 응답하지 않습니다. 잠시 후 다시 시도해 주세요.", failed.errorMessage)
        assertTrue(failed.canRetry)
    }

    @Test
    fun createPostsDraftThenReloads() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(201).setBody(routineJson(3, name = "아침")))
        enqueueLoad(listOf(routineJson(3, name = "아침")), listOf(dayJson(3)))

        vm.performCreate(
            RoutineDraft(
                name = "아침",
                cadence = RoutineCadence.WEEKDAYS,
                weekdays = RoutineWeekdays.bit(1) or RoutineWeekdays.bit(3),
                monthDay = 0,
                startDay = "2026-10-12",
                steps = listOf("물 한 컵"),
            ),
        )

        val post = server.takeRequest()
        assertEquals("POST", post.method)
        assertEquals("/api/routines", post.path)
        assertEquals(
            """{"name":"아침","cadence":"weekdays","weekdays":10,"month_day":0,"start_day":"2026-10-12","steps":["물 한 컵"]}""",
            post.body.readUtf8(),
        )
        assertEquals("/api/routines", server.takeRequest().path)
        assertEquals(listOf(3L), vm.state.value.routines.map { it.id })
    }

    @Test
    fun deleteRemovesFromBothLists() = runTest {
        val vm = viewModel()
        enqueueLoad(listOf(routineJson(1), routineJson(2)), listOf(dayJson(1)))
        vm.reload()
        server.takeRequest()
        server.takeRequest()
        server.enqueue(MockResponse().setResponseCode(204))

        vm.performDelete(1)

        val delete = server.takeRequest()
        assertEquals("DELETE", delete.method)
        assertEquals("/api/routines/1", delete.path)
        assertEquals(listOf(2L), vm.state.value.routines.map { it.id })
        assertTrue(vm.state.value.today.isEmpty())
    }

    @Test
    fun stepEndpointsMatchBackend() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(201).setBody(routineJson(1)))
        enqueueLoad(listOf(routineJson(1)), emptyList())
        vm.performAddStep(1, "스트레칭")
        val add = server.takeRequest()
        assertEquals("/api/routines/1/steps", add.path)
        assertEquals("""{"title":"스트레칭"}""", add.body.readUtf8())
        server.takeRequest()
        server.takeRequest()

        server.enqueue(MockResponse().setResponseCode(204))
        enqueueLoad(listOf(routineJson(1)), emptyList())
        vm.performDeleteStep(10, 1)
        val delete = server.takeRequest()
        assertEquals("DELETE", delete.method)
        assertEquals("/api/routines/1/steps/10", delete.path)
    }

    @Test
    fun historyEndsOnTheLoadedDay() = runTest {
        val vm = viewModel()
        server.enqueue(
            MockResponse().setBody("""{"items":[{"day":"2026-10-12","scheduled":true,"done":1,"total":2,"completed":false}]}"""),
        )

        val days = vm.history(7)

        assertEquals("/api/routines/7/history/2026-10-12", server.takeRequest().path)
        assertEquals(50, days.single().progressPercent)
    }

    @Test
    fun failedHistoryCannotBeRetried() = runTest {
        val vm = viewModel()
        server.enqueue(MockResponse().setResponseCode(404).setBody("not found"))

        assertTrue(vm.history(7).isEmpty())

        assertEquals("이력을 불러오지 못했습니다. 잠시 후 다시 시도해 주세요.", vm.state.value.errorMessage)
        assertFalse(vm.state.value.canRetry)
    }
}
