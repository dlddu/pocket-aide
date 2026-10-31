package com.dlddu.pocketaide.designsystem

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp

enum class Area(val label: String) {
    Personal("Personal"),
    Work("Work"),
    AiChat("AI Chat"),
    Scratchpad("Scratchpad"),
    Routines("Routines"),
    Affirmations("Affirmations"),
    Voice("Voice"),
    System("System"),
    PrMonitor("PR · Monitor"),
}

object DesignTokens {
    object Colors {
        @Composable
        @ReadOnlyComposable
        fun palette(area: Area): AreaPalette {
            val table = if (isSystemInDarkTheme()) GeneratedPalettes.dark else GeneratedPalettes.light
            return table.getValue(area)
        }

        @Composable
        @ReadOnlyComposable
        fun surface(area: Area): Color = palette(area).surface

        @Composable
        @ReadOnlyComposable
        fun ink(area: Area): Color = palette(area).ink

        @Composable
        @ReadOnlyComposable
        fun accent(area: Area): Color = palette(area).accent

        @Composable
        @ReadOnlyComposable
        fun rule(area: Area): Color = palette(area).rule

        @Composable
        @ReadOnlyComposable
        fun soft(area: Area): Color = palette(area).soft

        @Composable
        @ReadOnlyComposable
        fun card(area: Area): Color = palette(area).card

        @Composable
        @ReadOnlyComposable
        fun destructive(area: Area): Color = palette(area).destructive
    }

    object StatusColors {
        @Composable
        @ReadOnlyComposable
        fun success(): Color = adaptive(light = 0xFF4F6E5C, dark = 0xFF7CAB89)

        @Composable
        @ReadOnlyComposable
        fun failure(): Color = adaptive(light = 0xFF9C3F2D, dark = 0xFFD87560)

        @Composable
        @ReadOnlyComposable
        fun inProgress(): Color = adaptive(light = 0xFF8B6F47, dark = 0xFFC49B6F)

        @Composable
        @ReadOnlyComposable
        private fun adaptive(light: Long, dark: Long): Color =
            Color(if (isSystemInDarkTheme()) dark else light)
    }

    object Spacing {
        val xs = 4.dp
        val cardGap = 6.dp
        val sm = 8.dp
        val md = 12.dp
        val lg = 16.dp
        val xl = 20.dp
        val xxl = 24.dp
    }

    object Radius {
        val chip = 999.dp
        val key = 6.dp
        val card = 16.dp
        val cardLarge = 24.dp
        val bubble = 20.dp
    }

    object Typography {
        val caption2xs = 10.sp
        val captionXs = 11.sp
        val captionSm = 12.sp
        val bodySm = 13.sp
        val body = 14.sp
        val bodyLg = 15.sp
        val titleMd = 16.sp
        val h2 = 22.sp
        val h1 = 27.sp

        enum class Family { Sans, Serif }

        fun style(
            size: TextUnit,
            weight: FontWeight = FontWeight.Normal,
            family: Family = Family.Sans,
        ): TextStyle = TextStyle(
            fontSize = size,
            fontWeight = weight,
            fontFamily = if (family == Family.Serif) FontFamily.Serif else FontFamily.Default,
            letterSpacing = (-0.01).em,
        )
    }
}
