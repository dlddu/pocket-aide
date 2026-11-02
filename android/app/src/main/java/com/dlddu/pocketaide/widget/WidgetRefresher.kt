package com.dlddu.pocketaide.widget

import android.content.Context
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.longPreferencesKey
import androidx.glance.appwidget.GlanceAppWidgetManager
import androidx.glance.appwidget.state.updateAppWidgetState
import androidx.glance.appwidget.updateAll
import com.dlddu.pocketaide.PocketAideApplication
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.launch

object WidgetRefresher {
    val REVISION: Preferences.Key<Long> = longPreferencesKey("widget.revision")

    fun reloadAll(context: Context) {
        val appContext = context.applicationContext
        val container = (appContext as? PocketAideApplication)?.container ?: return
        container.appScope.launch { reloadNow(appContext) }
    }

    /** 백그라운드 작업처럼 끝나기 전에 갱신까지 마쳐야 할 때 쓴다. */
    suspend fun reloadNow(context: Context) {
        try {
            val revision = System.currentTimeMillis()
            GlanceAppWidgetManager(context).getGlanceIds(PocketAideWidget::class.java).forEach { id ->
                updateAppWidgetState(context, id) { it[REVISION] = revision }
            }
            PocketAideWidget().updateAll(context)
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            return
        }
    }
}
