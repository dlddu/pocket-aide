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
        container.appScope.launch { reload(appContext) }
    }

    private suspend fun reload(context: Context) {
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
