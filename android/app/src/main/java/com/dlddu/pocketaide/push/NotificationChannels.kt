package com.dlddu.pocketaide.push

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import com.dlddu.pocketaide.R

object NotificationChannels {
    const val DEFAULT = "default"
    const val APPROVAL = "approval"

    fun register(context: Context) {
        val manager = context.getSystemService(NotificationManager::class.java) ?: return
        manager.createNotificationChannels(
            listOf(
                NotificationChannel(
                    DEFAULT,
                    context.getString(R.string.notification_channel_default),
                    NotificationManager.IMPORTANCE_DEFAULT,
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
