package com.dlddu.pocketaide.notifications

import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import com.dlddu.pocketaide.PocketAideApplication
import com.dlddu.pocketaide.push.PushDeepLink
import com.dlddu.pocketaide.widget.WidgetRefresher
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeout

class ExternalNotificationActionReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action != ACTION_MARK_READ) return
        val externalId = intent.getLongExtra(EXTRA_EXTERNAL_ID, -1L).takeIf { it > 0 } ?: return
        val notificationId = intent.getIntExtra(EXTRA_NOTIFICATION_ID, 0)
        val app = context.applicationContext as PocketAideApplication
        val container = app.container
        val pending = goAsync()
        container.appScope.launch {
            try {
                val marked = runCatching {
                    withTimeout(MARK_TIMEOUT_MILLIS) { container.api.markExternalNotificationRead(externalId) }
                }.isSuccess
                if (marked) {
                    NotificationManagerCompat.from(app).cancel(notificationId)
                    WidgetRefresher.reloadAll(app)
                }
            } finally {
                pending.finish()
            }
        }
    }

    companion object {
        const val ACTION_MARK_READ = "EXTERNAL_NOTIFICATION_MARK_READ"
        const val MARK_READ_TITLE = "읽음"
        const val EXTRA_EXTERNAL_ID = "com.dlddu.pocketaide.notifications.EXTERNAL_ID"
        const val EXTRA_NOTIFICATION_ID = "com.dlddu.pocketaide.notifications.NOTIFICATION_ID"
        private const val MARK_TIMEOUT_MILLIS = 9_000L

        fun addAction(context: Context, builder: NotificationCompat.Builder, data: Map<String, String>, notificationId: Int) {
            val target = PushDeepLink.from(data) as? PushDeepLink.ExternalNotification ?: return
            val intent = Intent(context, ExternalNotificationActionReceiver::class.java)
                .setAction(ACTION_MARK_READ)
                .putExtra(EXTRA_EXTERNAL_ID, target.notificationId)
                .putExtra(EXTRA_NOTIFICATION_ID, notificationId)
            val pending = PendingIntent.getBroadcast(
                context,
                notificationId,
                intent,
                PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
            )
            builder.addAction(NotificationCompat.Action.Builder(0, MARK_READ_TITLE, pending).build())
        }
    }
}
