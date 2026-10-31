package com.dlddu.pocketaide.designsystem.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Text
import androidx.compose.material3.rememberModalBottomSheetState
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

private const val BACKDROP_OPACITY = 0.45f
private val HANDLE_WIDTH = 36.dp
private val HANDLE_HEIGHT = 4.dp
private val HANDLE_TOP = 10.dp
private val HANDLE_BOTTOM = 6.dp
private val ACTION_VERTICAL_PADDING = 14.dp
private val DESTRUCTIVE_STROKE = 1.4.dp
private const val DISABLED_OPACITY = 0.5f

object SheetDefaults {
    val horizontalPadding: Dp = DesignTokens.Spacing.xxl
    val titleSize = 18.sp
    val titleTopPadding: Dp = DesignTokens.Spacing.sm
    val cancelVerticalPadding: Dp = DesignTokens.Spacing.sm
}

@Composable
fun Handle(area: Area) {
    Box(
        Modifier
            .padding(top = HANDLE_TOP, bottom = HANDLE_BOTTOM)
            .size(width = HANDLE_WIDTH, height = HANDLE_HEIGHT)
            .clip(CircleShape)
            .background(DesignTokens.Colors.rule(area))
            .testTag("sheet.handle"),
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun Sheet(
    area: Area,
    onClose: () -> Unit,
    content: @Composable ColumnScope.() -> Unit,
) {
    val state = rememberModalBottomSheetState(skipPartiallyExpanded = true)
    ModalBottomSheet(
        onDismissRequest = onClose,
        sheetState = state,
        shape = RoundedCornerShape(topStart = DesignTokens.Radius.cardLarge, topEnd = DesignTokens.Radius.cardLarge),
        containerColor = DesignTokens.Colors.surface(area),
        contentColor = DesignTokens.Colors.ink(area),
        tonalElevation = 0.dp,
        scrimColor = DesignTokens.Colors.ink(area).copy(alpha = BACKDROP_OPACITY),
        dragHandle = { Handle(area) },
    ) {
        Column(Modifier.fillMaxWidth().verticalScroll(rememberScrollState()), content = content)
    }
}

@Composable
fun SheetPrimaryButton(
    area: Area,
    label: String,
    enabled: Boolean,
    tag: String,
    onClick: () -> Unit,
) {
    Box(
        Modifier
            .fillMaxWidth()
            .alpha(if (enabled) 1f else DISABLED_OPACITY)
            .clip(CircleShape)
            .background(DesignTokens.Colors.ink(area))
            .clickable(enabled = enabled, role = Role.Button, onClick = onClick)
            .padding(vertical = ACTION_VERTICAL_PADDING)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold),
            color = DesignTokens.Colors.surface(area),
        )
    }
}

@Composable
fun SheetDestructiveButton(
    area: Area,
    label: String,
    tag: String,
    onClick: () -> Unit,
) {
    val destructive = DesignTokens.Colors.destructive(area)
    Box(
        Modifier
            .fillMaxWidth()
            .clip(CircleShape)
            .border(DESTRUCTIVE_STROKE, destructive, CircleShape)
            .clickable(role = Role.Button, onClick = onClick)
            .padding(vertical = ACTION_VERTICAL_PADDING)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold),
            color = destructive,
        )
    }
}

@Composable
fun SheetCancelButton(
    area: Area,
    tag: String,
    onClick: () -> Unit,
    verticalPadding: Dp = SheetDefaults.cancelVerticalPadding,
) {
    Box(
        Modifier
            .fillMaxWidth()
            .clickable(role = Role.Button, onClick = onClick)
            .padding(vertical = verticalPadding)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            "취소",
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
            color = DesignTokens.Colors.ink(area).copy(alpha = 0.55f),
        )
    }
}
