package com.dlddu.pocketaide.prmonitor

import android.content.Intent
import android.provider.Settings
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.NotificationsOff
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Icon
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.FilterPills

private const val NOTICE_FILL_OPACITY = 0.12f
private const val DISABLED_OUTCOMES_OPACITY = 0.45f
private const val OUTCOME_LABEL_OPACITY = 0.7f
private val NOTICE_ICON_SIZE = 16.dp

@Composable
fun PrMonitorNotificationSettingsSheet(
    state: PrMonitorUiState,
    pushDenied: Boolean,
    onEnabledChange: (Boolean) -> Unit,
    onOutcomesChange: (NotificationOutcomes) -> Unit,
    onDismiss: () -> Unit,
    onSignOut: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val accent = DesignTokens.Colors.accent(PR_AREA)
    var confirmingSignOut by remember { mutableStateOf(false) }
    val settings = state.notificationSettings

    PrMonitorSheetScaffold(title = "알림 설정", onDone = onDismiss) {
        Column(
            Modifier
                .fillMaxWidth()
                .weight(1f)
                .padding(horizontal = DesignTokens.Spacing.xl)
                .padding(top = DesignTokens.Spacing.lg),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            if (pushDenied) PermissionNotice()
            Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                Text(
                    "CI 완료 푸시 받기",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
                    color = ink,
                    modifier = Modifier.weight(1f),
                )
                Switch(
                    checked = settings.enabled,
                    onCheckedChange = onEnabledChange,
                    enabled = !state.isLoadingSettings,
                    colors = SwitchDefaults.colors(checkedTrackColor = accent, checkedThumbColor = Color.White),
                    modifier = Modifier.testTag("prmonitor.settings.enabled"),
                )
            }
            val outcomesEnabled = !state.isLoadingSettings && settings.enabled
            Column(
                Modifier.alpha(if (settings.enabled) 1f else DISABLED_OUTCOMES_OPACITY),
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            ) {
                Text(
                    "받을 결과",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                    color = ink.copy(alpha = OUTCOME_LABEL_OPACITY),
                )
                FilterPills(
                    area = PR_AREA,
                    options = NotificationOutcomes.entries,
                    selection = settings.outcomes,
                    onSelect = { if (outcomesEnabled && it != settings.outcomes) onOutcomesChange(it) },
                    optionTag = { it.wire },
                    modifier = Modifier.testTag("prmonitor.settings.outcomes"),
                ) { option, _ ->
                    Text(
                        option.label,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
                    )
                }
            }
            state.settingsError?.let { message ->
                Text(
                    message,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                    color = DesignTokens.StatusColors.failure(),
                    modifier = Modifier.testTag("prmonitor.settings.error"),
                )
            }
            Text(
                "설정과 무관하게 모든 CI 완료 이벤트는 이력에 남습니다.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY),
            )
            Spacer(Modifier.weight(1f))
            Text(
                "로그아웃",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
                color = DesignTokens.StatusColors.failure(),
                modifier = Modifier
                    .align(Alignment.CenterHorizontally)
                    .clickable(role = Role.Button) { confirmingSignOut = true }
                    .padding(bottom = DesignTokens.Spacing.lg, top = DesignTokens.Spacing.sm)
                    .testTag("prmonitor.settings.signOut"),
            )
        }
    }

    if (confirmingSignOut) {
        AlertDialog(
            onDismissRequest = { confirmingSignOut = false },
            containerColor = DesignTokens.Colors.surface(PR_AREA),
            titleContentColor = ink,
            textContentColor = ink.copy(alpha = OUTCOME_LABEL_OPACITY),
            title = { Text("로그아웃할까요?") },
            text = { Text("이 기기에 저장된 GitHub 토큰도 함께 지워집니다. 다른 계정으로 다시 로그인할 수 있습니다.") },
            confirmButton = {
                TextButton(
                    onClick = {
                        confirmingSignOut = false
                        onDismiss()
                        onSignOut()
                    },
                    modifier = Modifier.testTag("prmonitor.settings.signOut.confirm"),
                ) {
                    Text("로그아웃", color = DesignTokens.StatusColors.failure())
                }
            },
            dismissButton = {
                TextButton(onClick = { confirmingSignOut = false }) {
                    Text("취소", color = accent)
                }
            },
        )
    }
}

@Composable
private fun PermissionNotice() {
    val context = LocalContext.current
    val ink = DesignTokens.Colors.ink(PR_AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(PrMonitorStyle.controlRadius))
            .background(DesignTokens.Colors.accent(PR_AREA).copy(alpha = NOTICE_FILL_OPACITY))
            .clickable(role = Role.Button) {
                runCatching {
                    context.startActivity(
                        Intent(Settings.ACTION_APP_NOTIFICATION_SETTINGS)
                            .putExtra(Settings.EXTRA_APP_PACKAGE, context.packageName)
                            .addFlags(Intent.FLAG_ACTIVITY_NEW_TASK),
                    )
                }
            }
            .padding(DesignTokens.Spacing.md)
            .testTag("prmonitor.settings.permissionNotice"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(Icons.Filled.NotificationsOff, contentDescription = null, tint = ink, modifier = Modifier.size(NOTICE_ICON_SIZE))
        Text(
            "시스템 알림 권한이 꺼져 있어 아래 설정과 관계없이 푸시가 도착하지 않습니다. 설정에서 켜기",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = ink,
        )
    }
}
