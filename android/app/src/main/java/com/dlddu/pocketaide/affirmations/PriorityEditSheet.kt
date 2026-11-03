package com.dlddu.pocketaide.affirmations

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.FilterPills
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetCancelButton
import com.dlddu.pocketaide.designsystem.components.SheetDefaults
import com.dlddu.pocketaide.designsystem.components.SheetDestructiveButton
import com.dlddu.pocketaide.designsystem.components.SheetPrimaryButton

private val TITLE_TRACKING = (-0.5).sp
private val TEXT_SIZE = 16.5.sp
private val TEXT_LINE_HEIGHT = 1.5.em
private const val TEXT_MIN_LINES = 2
private const val TEXT_MAX_LINES = 6
private val QUOTE_SIZE = 64.sp
private val QUOTE_LINE_HEIGHT = 1.em
private val QUOTE_OFFSET_X = (-4).dp
private val QUOTE_OFFSET_Y = (-8).dp
private const val QUOTE_OPACITY = 0.1f
private val CREATE_FIELD_VERTICAL = 14.dp
private val CREATE_FIELD_RADIUS = 24.dp
private val FIELD_BORDER = 1.dp
private val HINT_SIZE = 11.5.sp
private val HINT_LINE_HEIGHT = 1.625.em
private const val RULE_OPACITY = 0.6f
private val RULE_HEIGHT = 1.dp
private val ACTIONS_GAP = DesignTokens.Spacing.xs
private val ACTIONS_BOTTOM = 28.dp
private val CANCEL_VERTICAL = 10.dp
private val DOT_SIZE = 6.dp
private val DOT_GAP = 2.dp
private val DOTS_LABEL_GAP = 4.dp
private const val DOT_IDLE_OPACITY = 0.3f
private const val DOT_COUNT = 3

sealed interface PriorityEditMode {
    val title: String

    data object Create : PriorityEditMode {
        override val title: String = "새 다짐"
    }

    data class Edit(val affirmation: Affirmation) : PriorityEditMode {
        override val title: String = "우선순위 설정"
    }
}

@Composable
fun PriorityEditSheet(
    mode: PriorityEditMode,
    onSave: (String, AffirmationPriority) -> Unit,
    onCancel: () -> Unit,
    onDelete: (() -> Unit)? = null,
) {
    val area = Area.Affirmations
    val ink = DesignTokens.Colors.ink(area)
    val existing = (mode as? PriorityEditMode.Edit)?.affirmation
    val isEditing = existing != null
    var text by remember(mode) { mutableStateOf(existing?.text ?: "") }
    var priority by remember(mode) { mutableStateOf(existing?.priority ?: AffirmationPriority.NORMAL) }
    val trimmed = text.trim()

    Sheet(area = area, onClose = onCancel) {
        Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) {
            Text(
                mode.title,
                style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold)
                    .copy(letterSpacing = TITLE_TRACKING),
                color = ink,
                modifier = Modifier
                    .padding(horizontal = SheetDefaults.horizontalPadding)
                    .padding(top = SheetDefaults.titleTopPadding)
                    .testTag("sheet.title"),
            )
            Column(
                Modifier.padding(horizontal = SheetDefaults.horizontalPadding),
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xl),
            ) {
                EditorCard(isEditing = isEditing, text = text, onTextChange = { text = it })
                PrioritySection(priority = priority, onSelect = { priority = it })
            }
            Column {
                Box(
                    Modifier
                        .fillMaxWidth()
                        .height(RULE_HEIGHT)
                        .background(DesignTokens.Colors.rule(area).copy(alpha = RULE_OPACITY)),
                )
                Column(
                    Modifier
                        .padding(horizontal = SheetDefaults.horizontalPadding)
                        .padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM),
                    verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP),
                ) {
                    SheetPrimaryButton(
                        area = area,
                        label = "저장",
                        enabled = trimmed.isNotEmpty(),
                        tag = "sheet.save.button",
                        onClick = { if (trimmed.isNotEmpty()) onSave(trimmed, priority) },
                    )
                    if (isEditing && onDelete != null) {
                        SheetDestructiveButton(area = area, label = "삭제", tag = "sheet.delete.button", onClick = onDelete)
                    }
                    SheetCancelButton(
                        area = area,
                        tag = "sheet.cancel.button",
                        onClick = onCancel,
                        verticalPadding = CANCEL_VERTICAL,
                    )
                }
            }
        }
    }
}

@Composable
private fun EditorCard(isEditing: Boolean, text: String, onTextChange: (String) -> Unit) {
    val area = Area.Affirmations
    if (isEditing) {
        Box(Modifier.fillMaxWidth().clip(RoundedCornerShape(DesignTokens.Radius.card))) {
            Card(area = area, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
                AffirmationTextField(text = text, onTextChange = onTextChange)
            }
            Text(
                "\"",
                style = DesignTokens.Typography.style(QUOTE_SIZE, family = DesignTokens.Typography.Family.Serif)
                    .copy(lineHeight = QUOTE_LINE_HEIGHT),
                color = DesignTokens.Colors.accent(area).copy(alpha = QUOTE_OPACITY),
                modifier = Modifier
                    .align(Alignment.TopStart)
                    .offset(x = QUOTE_OFFSET_X, y = QUOTE_OFFSET_Y)
                    .clearAndSetSemantics {},
            )
        }
    } else {
        val shape = RoundedCornerShape(CREATE_FIELD_RADIUS)
        Box(
            Modifier
                .fillMaxWidth()
                .clip(shape)
                .background(DesignTokens.Colors.card(area))
                .border(FIELD_BORDER, DesignTokens.Colors.accent(area), shape)
                .padding(horizontal = DesignTokens.Spacing.lg, vertical = CREATE_FIELD_VERTICAL),
        ) {
            AffirmationTextField(text = text, onTextChange = onTextChange)
        }
    }
}

@Composable
private fun AffirmationTextField(text: String, onTextChange: (String) -> Unit) {
    val ink = DesignTokens.Colors.ink(Area.Affirmations)
    PlainTextField(
        value = text,
        onValueChange = onTextChange,
        placeholder = "다짐 문장을 입력하세요",
        textStyle = DesignTokens.Typography.style(TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)
            .copy(lineHeight = TEXT_LINE_HEIGHT),
        textColor = ink,
        placeholderColor = ink.copy(alpha = 0.45f),
        minLines = TEXT_MIN_LINES,
        maxLines = TEXT_MAX_LINES,
        modifier = Modifier.fillMaxWidth().testTag("sheet.text.field"),
    )
}

@Composable
private fun PrioritySection(priority: AffirmationPriority, onSelect: (AffirmationPriority) -> Unit) {
    val area = Area.Affirmations
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) {
        Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
            AreaLabel(area = area, text = "노출 빈도")
            FilterPills(
                area = area,
                options = AffirmationPriority.entries,
                selection = priority,
                onSelect = onSelect,
                optionTag = { it.wire },
            ) { option, active ->
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(DOTS_LABEL_GAP),
                ) {
                    PriorityDots(option)
                    Text(
                        option.displayName,
                        style = DesignTokens.Typography.style(
                            DesignTokens.Typography.body,
                            if (active) FontWeight.Bold else FontWeight.Medium,
                        ),
                    )
                }
            }
        }
        Text(
            "위젯과 다짐 회전 노출에 얼마나 자주 등장할지 정합니다. 나중에 카드를 길게 눌러 변경할 수 있습니다.",
            style = DesignTokens.Typography.style(HINT_SIZE).copy(lineHeight = HINT_LINE_HEIGHT),
            color = DesignTokens.Colors.ink(area).copy(alpha = 0.55f),
        )
    }
}

@Composable
private fun PriorityDots(option: AffirmationPriority) {
    val accent = DesignTokens.Colors.accent(Area.Affirmations)
    val filled = when (option) {
        AffirmationPriority.HIGH -> 3
        AffirmationPriority.NORMAL -> 2
        AffirmationPriority.LOW -> 1
    }
    Row(horizontalArrangement = Arrangement.spacedBy(DOT_GAP)) {
        repeat(DOT_COUNT) { index ->
            Box(
                Modifier
                    .size(DOT_SIZE)
                    .background(accent.copy(alpha = if (index < filled) 1f else DOT_IDLE_OPACITY), CircleShape),
            )
        }
    }
}
