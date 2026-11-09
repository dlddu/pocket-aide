package com.dlddu.pocketaide.push

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import com.dlddu.pocketaide.R

object NotificationChannels {
    // A channel's importance is fixed once created, so raising it to heads-up needs a new ID.
    const val DEFAULT = "general"
    const val APPROVAL = "approval"

    private const val LEGACY_DEFAULT = "default"

    fun register(context: Context) {
        val manager = context.getSystemService(NotificationManager::class.java) ?: return
        manager.deleteNotificationChannel(LEGACY_DEFAULT)
        manager.createNotificationChannels(
            listOf(
                NotificationChannel(
                    DEFAULT,
                    context.getString(R.string.notification_channel_default),
                    NotificationManager.IMPORTANCE_HIGH,
                ),
                NotificationChannel(
                    APPROVAL,
                    context.getString(R.string.notification_channel_approval),
                    NotificationManager.IMPORTANCE_HIGH,
                ),
            ),
        )
    }

    fun forCategory(category: String?): String = if (category == "APPROVAL_REQUEST") APPROVAL else DEFAULT
}
