package com.dlddu.pocketaide.todos

import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabelDotShape

private val PERSONAL_CHECK_SIZE = 24.dp
private val WORK_CHECK_SIZE = 20.dp
private val PERSONAL_SECTION_TITLE_SIZE = 14.sp
private val WORK_SECTION_TITLE_TRACKING = 0.3.sp
private val WORK_SECTION_HEADER_BOTTOM = 6.dp

val TodoArea.designArea: Area
    get() = when (this) {
        TodoArea.PERSONAL -> Area.Personal
        TodoArea.WORK -> Area.Work
    }

internal val TodoArea.rowRadius: Dp
    get() = when (this) {
        TodoArea.PERSONAL -> DesignTokens.Radius.card
        TodoArea.WORK -> 0.dp
    }

internal val TodoArea.headerDotShape: AreaLabelDotShape
    get() = when (this) {
        TodoArea.PERSONAL -> AreaLabelDotShape.Circle
        TodoArea.WORK -> AreaLabelDotShape.Square
    }

internal val TodoArea.checkSize: Dp
    get() = when (this) {
        TodoArea.PERSONAL -> PERSONAL_CHECK_SIZE
        TodoArea.WORK -> WORK_CHECK_SIZE
    }

internal val TodoArea.sectionTitleSize: TextUnit
    get() = when (this) {
        TodoArea.PERSONAL -> PERSONAL_SECTION_TITLE_SIZE
        TodoArea.WORK -> DesignTokens.Typography.captionSm
    }

internal val TodoArea.sectionTitleTracking: TextUnit
    get() = when (this) {
        TodoArea.PERSONAL -> 0.sp
        TodoArea.WORK -> WORK_SECTION_TITLE_TRACKING
    }

internal val TodoArea.sectionTitleUppercase: Boolean get() = this == TodoArea.WORK

internal val TodoArea.rowTitleWeight: FontWeight
    get() = when (this) {
        TodoArea.PERSONAL -> FontWeight.Normal
        TodoArea.WORK -> FontWeight.Medium
    }

internal val TodoArea.listTopInset: Dp
    get() = when (this) {
        TodoArea.PERSONAL -> DesignTokens.Spacing.xs
        TodoArea.WORK -> DesignTokens.Spacing.sm
    }

internal val TodoArea.sectionHeaderBottom: Dp
    get() = when (this) {
        TodoArea.PERSONAL -> DesignTokens.Spacing.xs
        TodoArea.WORK -> WORK_SECTION_HEADER_BOTTOM
    }

internal val TodoArea.sectionHeadSpacing: Dp
    get() = when (this) {
        TodoArea.PERSONAL -> 0.dp
        TodoArea.WORK -> DesignTokens.Spacing.sm
    }

internal val TodoArea.sectionBarWidth: Dp
    get() = when (this) {
        TodoArea.PERSONAL -> 0.dp
        TodoArea.WORK -> DesignTokens.Spacing.xs
    }

internal val TodoArea.showsSectionRule: Boolean get() = this == TodoArea.WORK

internal val TodoArea.checkIsRound: Boolean get() = this == TodoArea.PERSONAL

internal val TodoArea.searchRadius: Dp
    get() = if (this == TodoArea.PERSONAL) DesignTokens.Radius.card else DesignTokens.Radius.key
