package com.dlddu.pocketaide.ui

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetCancelButton
import com.dlddu.pocketaide.designsystem.components.SheetDefaults
import com.dlddu.pocketaide.designsystem.components.SheetPrimaryButton
import com.dlddu.pocketaide.push.DeviceName

private val ACTIONS_GAP = DesignTokens.Spacing.xs
private val ACTIONS_BOTTOM = DesignTokens.Spacing.xxl + DesignTokens.Spacing.sm
private val DESCRIPTION_GAP = DesignTokens.Spacing.xs

@Composable
fun DeviceNameSheet(initial: String, onSave: (String) -> Unit, onSkip: () -> Unit) {
    val area = Area.PrMonitor
    val ink = DesignTokens.Colors.ink(area)
    var text by remember { mutableStateOf(initial) }
    val name = DeviceName.normalize(text)

    Sheet(area = area, onClose = onSkip) {
        Column(
            Modifier.fillMaxWidth().padding(horizontal = SheetDefaults.horizontalPadding),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            Column(
                Modifier.padding(top = SheetDefaults.titleTopPadding),
                verticalArrangement = Arrangement.spacedBy(DESCRIPTION_GAP),
            ) {
                Text(
                    "이 기기 이름",
                    style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold),
                    color = ink,
                    modifier = Modifier.testTag("device.name.sheet.title"),
                )
                Text(
                    "푸시를 받는 기기를 구분하는 이름이에요.",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
                    color = ink.copy(alpha = 0.6f),
                )
            }
            Card(area = area, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
                PlainTextField(
                    value = text,
                    onValueChange = { text = it.take(DeviceName.MAX_LENGTH) },
                    placeholder = "예: 출근용 폰",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = 0.45f),
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(imeAction = ImeAction.Done),
                    modifier = Modifier.fillMaxWidth().testTag("device.name.sheet.field"),
                )
            }
            Column(
                Modifier.padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM),
                verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP),
            ) {
                SheetPrimaryButton(
                    area = area,
                    label = "저장",
                    enabled = name.isNotEmpty(),
                    tag = "device.name.sheet.save.button",
                    onClick = { if (name.isNotEmpty()) onSave(name) },
                )
                SheetCancelButton(
                    area = area,
                    tag = "device.name.sheet.skip.button",
                    onClick = onSkip,
                    label = "건너뛰기",
                )
            }
        }
    }
}
