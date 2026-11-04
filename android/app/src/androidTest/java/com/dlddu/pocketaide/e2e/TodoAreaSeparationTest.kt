// 검증 시나리오: test-todo.md#시나리오 1
package com.dlddu.pocketaide.e2e

import androidx.test.ext.junit.runners.AndroidJUnit4
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class TodoAreaSeparationTest : E2eTest() {
    private val rowTag = { area: String -> Regex("todos\\.$area\\.row\\.\\d+") }

    @Test
    fun itemAddedToOneAreaNeverAppearsInTheOther() {
        val token = E2e.uniqueToken()
        val personalTitle = "치과 예약 잡기 $token"
        val workTitle = "분기 리뷰 자료 $token"
        val api = BackendApi.signIn()
        var scenario = launchSignedIn()

        compose.openMore("Personal")
        addTodo("personal", personalTitle)
        compose.openMore("Work")
        addTodo("work", workTitle)

        assertNotNull("The personal list should store its item", api.todo("personal", personalTitle))
        assertNotNull("The work list should store its item", api.todo("work", workTitle))
        assertNull("The work list must not hold the personal item", api.todo("work", personalTitle))
        assertNull("The personal list must not hold the work item", api.todo("personal", workTitle))

        assertSeparated(api, token, personalTitle, workTitle)

        scenario = relaunch(scenario)
        assertSeparated(api, token, personalTitle, workTitle)
        scenario.close()
    }

    private fun assertSeparated(api: BackendApi, token: String, personalTitle: String, workTitle: String) {
        compose.openMore("Personal")
        showOnly("personal", token, mine = personalTitle, theirs = workTitle)
        val personalOpen = api.todos("personal").count { it.isNull("completed_at") }
        val personalDone = api.todos("personal").size - personalOpen
        compose.waitForText("todos.personal.summary", "${personalOpen}개 남음 · ${personalDone}개 완료")

        compose.openMore("Work")
        showOnly("work", token, mine = workTitle, theirs = personalTitle)
        val workOpen = api.todos("work").count { it.isNull("completed_at") }
        val workDone = api.todos("work").size - workOpen
        compose.waitUntil(15_000) { compose.textOf("todos.work.summary").startsWith("$workOpen OPEN · $workDone DONE") }
    }

    private fun showOnly(area: String, token: String, mine: String, theirs: String) {
        compose.type("todos.$area.search.field", token)
        compose.waitUntil(15_000) { compose.tagWithText(rowTag(area), mine) != null }
        assertTrue("'$mine' should be listed in $area", compose.tagWithText(rowTag(area), mine) != null)
        assertFalse("'$theirs' must not be listed in $area", compose.hasAnyText(theirs))
    }

    private fun addTodo(area: String, title: String) {
        compose.tap("todos.$area.add.button")
        compose.type("todo.sheet.title.field", title)
        compose.tap("todo.sheet.save.button")
        compose.waitForTagGone("todo.sheet.title.field")
        compose.waitUntil(15_000) { compose.tagWithText(rowTag(area), title) != null }
    }
}
