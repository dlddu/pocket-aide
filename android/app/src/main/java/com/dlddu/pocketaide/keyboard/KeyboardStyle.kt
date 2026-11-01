package com.dlddu.pocketaide.keyboard

import android.content.res.Configuration
import android.content.res.Resources
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.unit.Dp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.GeneratedPalettes

internal data class KeyboardStyle(
    val surface: Int,
    val key: Int,
    val ink: Int,
    val accent: Int,
    val onAccent: Int,
) {
    companion object {
        fun of(resources: Resources): KeyboardStyle {
            val dark = (resources.configuration.uiMode and Configuration.UI_MODE_NIGHT_MASK) ==
                Configuration.UI_MODE_NIGHT_YES
            val system = (if (dark) GeneratedPalettes.dark else GeneratedPalettes.light).getValue(Area.System)
            val aiChat = (if (dark) GeneratedPalettes.dark else GeneratedPalettes.light).getValue(Area.AiChat)
            return KeyboardStyle(
                surface = (if (dark) GeneratedPalettes.keyboardSurfaceDark else GeneratedPalettes.keyboardSurfaceLight).toArgb(),
                key = (if (dark) GeneratedPalettes.keyboardKeyDark else GeneratedPalettes.keyboardKeyLight).toArgb(),
                ink = system.ink.toArgb(),
                accent = aiChat.accent.toArgb(),
                onAccent = system.card.toArgb(),
            )
        }

        fun px(resources: Resources, dp: Dp): Int = (dp.value * resources.displayMetrics.density).toInt()

        val textSizeSp: Float get() = DesignTokens.Typography.body.value
    }
}

internal object KeyboardCopy {
    const val INSERT_TITLE = "Insert hello from pocket-aide"
    const val INSERTED_TEXT = "Hello from pocket-aide"
    const val NEXT_KEYBOARD = "next ⇄"
}
