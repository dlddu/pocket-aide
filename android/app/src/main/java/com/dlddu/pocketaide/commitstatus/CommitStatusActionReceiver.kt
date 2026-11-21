package com.dlddu.pocketaide.commitstatus

import android.Manifest
import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import androidx.core.content.ContextCompat
import com.dlddu.pocketaide.MainActivity
import com.dlddu.pocketaide.PocketAideApplication
import com.dlddu.pocketaide.R
import com.dlddu.pocketaide.approvals.ApprovalDecision
import com.dlddu.pocketaide.push.NotificationChannels
import com.dlddu.pocketaide.push.PushDeepLink
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeout

class CommitStatusActionReceiver : BroadcastReceiver() {
    override fun onReceive(context: Context, intent: Intent) {
        val decision = CommitStatusNotificationAction.decision(intent.action) ?: return
        val requestId = intent.getStringExtra(EXTRA_REQUEST_ID)?.takeIf { it.isNotBlank() } ?: return
        val notificationId = intent.getIntExtra(EXTRA_NOTIFICATION_ID, 0)
        val title = intent.getStringExtra(EXTRA_TITLE)
        val app = context.applicationContext as PocketAideApplication
        val container = app.container
        val pending = goAsync()
        container.appScope.launch {
            try {
                val result = try {
                    Result.success(withTimeout(DECISION_TIMEOUT_MILLIS) { container.api.decideCommitStatusRequest(requestId, decision) })
                } catch (e: Exception) {
                    Result.failure(e)
                }
                if (result.getOrNull() is CommitStatusDecisionOutcome.Decided) {
                    NotificationManagerCompat.from(app).cancel(notificationId)
                } else {
                    showResult(app, notificationId, requestId, title, CommitStatusNotificationAction.resultNotice(result))
                }
            } finally {
                pending.finish()
            }
        }
    }

    private fun showResult(context: Context, notificationId: Int, requestId: String, title: String?, body: String) {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU &&
            ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) !=
            PackageManager.PERMISSION_GRANTED
        ) {
            return
        }
        val notification = NotificationCompat.Builder(context, NotificationChannels.APPROVAL)
            .setSmallIcon(R.drawable.ic_notification)
            .setColor(ContextCompat.getColor(context, R.color.icon_background))
            .setContentTitle(title)
            .setContentText(body)
            .setStyle(NotificationCompat.BigTextStyle().bigText(body))
            .setAutoCancel(true)
            .setOnlyAlertOnce(true)
            .setContentIntent(openIntent(context, notificationId, requestId))
            .build()
        NotificationManagerCompat.from(context).notify(notificationId, notification)
    }

    companion object {
        const val EXTRA_REQUEST_ID = "com.dlddu.pocketaide.commitstatus.REQUEST_ID"
        const val EXTRA_NOTIFICATION_ID = "com.dlddu.pocketaide.commitstatus.NOTIFICATION_ID"
        const val EXTRA_TITLE = "com.dlddu.pocketaide.commitstatus.TITLE"
        private const val DECISION_TIMEOUT_MILLIS = 9_000L
        private const val CATEGORY_KEY = "category"

        fun addActions(
            context: Context,
            builder: NotificationCompat.Builder,
            data: Map<String, String>,
            notificationId: Int,
            title: String?,
        ) {
            if (data[CATEGORY_KEY] != CommitStatusNotificationAction.CATEGORY) return
            val requestId = data[PushDeepLink.KEY_COMMIT_STATUS_REQUEST_ID]?.takeIf { it.isNotBlank() } ?: return
            ApprovalDecision.entries.forEach { decision ->
                val action = NotificationCompat.Action.Builder(
                    0,
                    CommitStatusNotificationAction.title(decision),
                    decisionIntent(context, decision, requestId, notificationId, title),
                )
                    .setAuthenticationRequired(decision == ApprovalDecision.APPROVE)
                    .build()
                builder.addAction(action)
            }
        }

        private fun decisionIntent(
            context: Context,
            decision: ApprovalDecision,
            requestId: String,
            notificationId: Int,
            title: String?,
        ): PendingIntent {
            val intent = Intent(context, CommitStatusActionReceiver::class.java)
                .setAction(CommitStatusNotificationAction.action(decision))
                .putExtra(EXTRA_REQUEST_ID, requestId)
                .putExtra(EXTRA_NOTIFICATION_ID, notificationId)
                .putExtra(EXTRA_TITLE, title)
            return PendingIntent.getBroadcast(
                context,
                notificationId * 2 + decision.ordinal,
                intent,
                PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
            )
        }

        private fun openIntent(context: Context, notificationId: Int, requestId: String): PendingIntent {
            val intent = Intent(context, MainActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
                putExtra(PushDeepLink.KEY_KIND, PushDeepLink.KIND_COMMIT_STATUS_REQUEST)
                putExtra(PushDeepLink.KEY_COMMIT_STATUS_REQUEST_ID, requestId)
            }
            return PendingIntent.getActivity(
                context,
                notificationId,
                intent,
                PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
            )
        }
    }
}
