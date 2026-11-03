package com.dlddu.pocketaide.prmonitor

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Handle

private const val SCRIM_OPACITY = 0.45f
private const val SHEET_HEIGHT_FRACTION = 0.92f
private val NO_ELEVATION = 0.dp

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PrMonitorSheetScaffold(
    title: String,
    onDone: () -> Unit,
    modifier: Modifier = Modifier,
    area: Area = Area.PrMonitor,
    content: @Composable ColumnScope.() -> Unit,
) {
    val state = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    val ink = DesignTokens.Colors.ink(area)
    ModalBottomSheet(
        onDismissRequest = onDone,
        sheetState = state,
        shape = RoundedCornerShape(topStart = DesignTokens.Radius.cardLarge, topEnd = DesignTokens.Radius.cardLarge),
        containerColor = DesignTokens.Colors.surface(area),
        contentColor = ink,
        tonalElevation = NO_ELEVATION,
        scrimColor = ink.copy(alpha = SCRIM_OPACITY),
        dragHandle = { Handle(area) },
    ) {
        Column(modifier.fillMaxWidth().fillMaxHeight(SHEET_HEIGHT_FRACTION)) {
            Box(
                Modifier
                    .fillMaxWidth()
                    .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.sm),
            ) {
                Text(
                    title,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.SemiBold),
                    color = ink,
                    modifier = Modifier.align(Alignment.Center),
                )
                Text(
                    "완료",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
                    color = DesignTokens.Colors.accent(area),
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .clickable(role = Role.Button, onClick = onDone)
                        .padding(vertical = DesignTokens.Spacing.xs)
                        .testTag("prmonitor.sheet.done"),
                )
            }
            content()
        }
    }
}
