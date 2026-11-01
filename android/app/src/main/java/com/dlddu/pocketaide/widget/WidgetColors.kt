package com.dlddu.pocketaide.widget

import androidx.compose.ui.graphics.Color
import androidx.glance.color.ColorProvider as DayNightColorProvider
import androidx.glance.unit.ColorProvider
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.AreaPalette
import com.dlddu.pocketaide.designsystem.GeneratedPalettes

internal object WidgetColors {
    val surface: ColorProvider = DayNightColorProvider(day = GeneratedPalettes.widgetSurfaceLight, night = GeneratedPalettes.widgetSurfaceDark)
    val rule: ColorProvider = DayNightColorProvider(day = GeneratedPalettes.widgetRuleLight, night = GeneratedPalettes.widgetRuleDark)

    fun ink(area: Area, alpha: Float = 1f): ColorProvider = adaptive(area, alpha) { it.ink }

    fun accent(area: Area): ColorProvider = adaptive(area, 1f) { it.accent }

    private fun adaptive(area: Area, alpha: Float, pick: (AreaPalette) -> Color): ColorProvider =
        DayNightColorProvider(
            day = pick(GeneratedPalettes.light.getValue(area)).copy(alpha = alpha),
            night = pick(GeneratedPalettes.dark.getValue(area)).copy(alpha = alpha),
        )
}
