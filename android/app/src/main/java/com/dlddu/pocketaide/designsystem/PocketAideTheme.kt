package com.dlddu.pocketaide.designsystem

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable

@Composable
fun PocketAideTheme(content: @Composable () -> Unit) {
    val system = DesignTokens.Colors.palette(Area.System)
    val base = if (isSystemInDarkTheme()) darkColorScheme() else lightColorScheme()
    val scheme = base.copy(
        primary = system.accent,
        background = system.surface,
        surface = system.surface,
        onBackground = system.ink,
        onSurface = system.ink,
        outline = system.rule,
    )
    MaterialTheme(colorScheme = scheme, content = content)
}
