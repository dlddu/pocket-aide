package com.dlddu.pocketaide.routines

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

@Composable
fun RoutineErrorRow(
    message: String,
    canRetry: Boolean,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Row(modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        Text(
            message,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = DesignTokens.Colors.destructive(Area.Routines),
            modifier = Modifier.alignByBaseline().weight(1f).testTag("routines.error"),
        )
        if (canRetry) {
            Text(
                "다시 시도",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
                color = DesignTokens.Colors.accent(Area.Routines),
                modifier = Modifier
                    .alignByBaseline()
                    .clickable(role = Role.Button, onClick = onRetry)
                    .testTag("routines.error.retry"),
            )
        }
    }
}
