package com.dlddu.pocketaide.e2e

import androidx.compose.ui.test.performTouchInput
import androidx.compose.ui.test.swipeLeft
import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class PersonalTodoTest : E2eTest() {
    private val area = "personal"
    private val rowTag = Regex("todos\\.$area\\.row\\.\\d+")

    @Test
    fun addCompleteAndDeletePersistToTheBackend() {
        val token = E2e.uniqueToken()
        val kept = "우유 사기 $token"
        val swiped = "오래된 메모 $token"
        val api = BackendApi.signIn()
        var scenario = launchSignedIn()
        compose.openMore("Personal")

        addTodo(kept)
        addTodo(swiped)
        val keptId = rowId(kept)
        val swipedId = rowId(swiped)
        assertNotNull("The backend should store '$kept'", api.todo(area, kept))
        assertNotNull("The backend should store '$swiped'", api.todo(area, swiped))

        compose.type("todos.$area.search.field", token)
        compose.tap("todos.$area.row.$keptId.toggle")
        assertTrue("Completing should reach the backend", E2e.eventually { api.todo(area, kept)?.isNull("completed_at") == false })
        compose.waitUntil(15_000) { compose.hasAnyText(DONE_ONE) }

        scenario = relaunch(scenario)
        compose.openMore("Personal")
        compose.type("todos.$area.search.field", token)
        compose.waitForTag("todos.$area.row.$keptId")
        assertTrue("The completed item should stay in the done section after a relaunch", compose.hasAnyText(DONE_ONE))
        assertFalse("The other item should stay open", api.todo(area, swiped)?.isNull("completed_at") == false)

        compose.node("todos.$area.row.$swipedId").performTouchInput { swipeLeft() }
        compose.waitForTagGone("todos.$area.row.$swipedId")
        assertTrue("Swipe-deleting should reach the backend", E2e.eventually { api.todo(area, swiped) == null })

        compose.tap("todos.$area.row.$keptId")
        compose.tap("todo.sheet.delete.button")
        compose.waitForTagGone("todos.$area.row.$keptId")
        assertTrue("Sheet-deleting should reach the backend", E2e.eventually { api.todo(area, kept) == null })

        scenario = relaunch(scenario)
        compose.openMore("Personal")
        compose.waitForTag("todos.$area.summary")
        assertFalse("A deleted item stays gone after a relaunch", compose.exists("todos.$area.row.$keptId"))
        assertFalse("A deleted item stays gone after a relaunch", compose.exists("todos.$area.row.$swipedId"))
        scenario.close()
    }

    private fun addTodo(title: String) {
        compose.tap("todos.$area.add.button")
        compose.type("todo.sheet.title.field", title)
        compose.tap("todo.sheet.save.button")
        compose.waitForTagGone("todo.sheet.title.field")
        compose.waitUntil(15_000) { compose.tagWithText(rowTag, title) != null }
    }

    private fun rowId(title: String): String =
        requireNotNull(compose.tagWithText(rowTag, title)) { "'$title' should be listed" }.substringAfterLast('.')

    private companion object {
        const val DONE_ONE = "완료 · 1"
    }
}
