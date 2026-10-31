package com.dlddu.pocketaide.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

@Composable
fun PlaceholderScreen(area: Area, title: String) {
    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(area))
            .testTag("placeholder.${area.name}"),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
        ) {
            Text(
                area.label.uppercase(),
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold),
                color = DesignTokens.Colors.accent(area),
            )
            Text(
                title,
                style = DesignTokens.Typography.style(DesignTokens.Typography.h2, FontWeight.Bold),
                color = DesignTokens.Colors.ink(area),
            )
            Text(
                "준비 중",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Medium),
                color = DesignTokens.Colors.ink(area).copy(alpha = 0.55f),
            )
        }
    }
}
