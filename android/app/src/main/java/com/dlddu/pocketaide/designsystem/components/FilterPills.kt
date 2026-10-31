package com.dlddu.pocketaide.designsystem.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.LocalContentColor
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

private val PILL_BORDER_WIDTH = 1.dp

@Composable
fun <T> FilterPills(
    area: Area,
    options: List<T>,
    selection: T,
    onSelect: (T) -> Unit,
    optionTag: (T) -> String,
    modifier: Modifier = Modifier,
    label: @Composable (option: T, active: Boolean) -> Unit,
) {
    val ink = DesignTokens.Colors.ink(area)
    val soft = DesignTokens.Colors.soft(area)
    val accent = DesignTokens.Colors.accent(area)
    val rule = DesignTokens.Colors.rule(area)
    Row(modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
        options.forEach { option ->
            val active = option == selection
            Box(
                Modifier
                    .weight(1f)
                    .clip(CircleShape)
                    .background(if (active) soft else Color.Transparent)
                    .border(PILL_BORDER_WIDTH, if (active) accent.copy(alpha = 0.4f) else rule, CircleShape)
                    .selectable(selected = active, role = Role.RadioButton, onClick = { onSelect(option) })
                    .padding(vertical = DesignTokens.Spacing.md)
                    .testTag("filter.pill.${optionTag(option)}"),
                contentAlignment = Alignment.Center,
            ) {
                CompositionLocalProvider(LocalContentColor provides if (active) ink else ink.copy(alpha = 0.55f)) {
                    label(option, active)
                }
            }
        }
    }
}
