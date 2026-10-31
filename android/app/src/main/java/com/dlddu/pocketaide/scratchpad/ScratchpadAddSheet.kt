package com.dlddu.pocketaide.scratchpad

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetCancelButton
import com.dlddu.pocketaide.designsystem.components.SheetDefaults
import com.dlddu.pocketaide.designsystem.components.SheetPrimaryButton

private const val TEXT_MIN_LINES = 2
private const val TEXT_MAX_LINES = 6
private val ACTIONS_GAP = DesignTokens.Spacing.xs
private val ACTIONS_BOTTOM = DesignTokens.Spacing.xxl + DesignTokens.Spacing.sm

@Composable
fun ScratchpadAddSheet(onSave: (String) -> Unit, onCancel: () -> Unit) {
    val area = Area.Scratchpad
    val ink = DesignTokens.Colors.ink(area)
    var text by remember { mutableStateOf("") }
    val trimmed = text.trim()

    Sheet(area = area, onClose = onCancel) {
        Column(
            Modifier.fillMaxWidth().padding(horizontal = SheetDefaults.horizontalPadding),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            Text(
                "새 메모",
                style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold),
                color = ink,
                modifier = Modifier.padding(top = SheetDefaults.titleTopPadding).testTag("scratchpad.sheet.title"),
            )
            Card(area = area, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
                PlainTextField(
                    value = text,
                    onValueChange = { text = it },
                    placeholder = "분류는 나중에 — 일단 적어두세요",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = 0.45f),
                    minLines = TEXT_MIN_LINES,
                    maxLines = TEXT_MAX_LINES,
                    modifier = Modifier.fillMaxWidth().testTag("scratchpad.sheet.text.field"),
                )
            }
            Column(
                Modifier.padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM),
                verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP),
            ) {
                SheetPrimaryButton(
                    area = area,
                    label = "저장",
                    enabled = trimmed.isNotEmpty(),
                    tag = "scratchpad.sheet.save.button",
                    onClick = { if (trimmed.isNotEmpty()) onSave(trimmed) },
                )
                SheetCancelButton(area = area, tag = "scratchpad.sheet.cancel.button", onClick = onCancel)
            }
        }
    }
}
