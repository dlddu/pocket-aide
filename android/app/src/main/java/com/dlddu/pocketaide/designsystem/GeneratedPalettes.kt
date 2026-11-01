package com.dlddu.pocketaide.designsystem

import androidx.compose.ui.graphics.Color

data class AreaPalette(
    val surface: Color,
    val ink: Color,
    val accent: Color,
    val rule: Color,
    val soft: Color,
    val card: Color,
    val destructive: Color,
)

internal object GeneratedPalettes {
    val light: Map<Area, AreaPalette> = mapOf(
        Area.Personal to AreaPalette(surface = Color(0xFFFBF1EA), ink = Color(0xFF3D2A22), accent = Color(0xFFB65A3C), rule = Color(0xFFEBD9CB), soft = Color(0xFFF4E2D4), card = Color(0xFFFFFFFF), destructive = Color(0xFF9E2B3C)),
        Area.Work to AreaPalette(surface = Color(0xFFEEF2F8), ink = Color(0xFF1E2A3A), accent = Color(0xFF355577), rule = Color(0xFFD6DEE9), soft = Color(0xFFDDE5F0), card = Color(0xFFFFFFFF), destructive = Color(0xFF9C3F2D)),
        Area.AiChat to AreaPalette(surface = Color(0xFFFAFAF7), ink = Color(0xFF1C2624), accent = Color(0xFF5E8B73), rule = Color(0xFFECEAE3), soft = Color(0xFFECEAE3), card = Color(0xFFFFFFFF), destructive = Color(0x00000000)),
        Area.Scratchpad to AreaPalette(surface = Color(0xFFF5EFE0), ink = Color(0xFF2A2723), accent = Color(0xFFB6855E), rule = Color(0xFFE0D8C2), soft = Color(0xFFFBF7EC), card = Color(0xFFFBF7EC), destructive = Color(0xFF9C3F2D)),
        Area.Routines to AreaPalette(surface = Color(0xFFF0F2EC), ink = Color(0xFF243329), accent = Color(0xFF4F6E5C), rule = Color(0xFFD6DDD2), soft = Color(0xFFE0E7DA), card = Color(0xFFFFFFFF), destructive = Color(0xFF9C3F2D)),
        Area.Affirmations to AreaPalette(surface = Color(0xFFF4EBDD), ink = Color(0xFF2E251A), accent = Color(0xFF8B6F47), rule = Color(0xFFE5D7C0), soft = Color(0xFFEADCC2), card = Color(0xFFFFFFFF), destructive = Color(0xFF9C3F2D)),
        Area.Voice to AreaPalette(surface = Color(0xFF0F1614), ink = Color(0xFFFFFFFF), accent = Color(0xFF5E8B73), rule = Color(0xFF1C2624), soft = Color(0xFF1A2421), card = Color(0xFF1A2421), destructive = Color(0x00000000)),
        Area.System to AreaPalette(surface = Color(0xFFFFFFFF), ink = Color(0xFF1C2624), accent = Color(0xFF5E8B73), rule = Color(0xFFE5E5E5), soft = Color(0xFFF5F5F5), card = Color(0xFFFFFFFF), destructive = Color(0x00000000)),
        Area.PrMonitor to AreaPalette(surface = Color(0xFFEEEDF5), ink = Color(0xFF221F33), accent = Color(0xFF5B4DB8), rule = Color(0xFFD8D5E4), soft = Color(0xFFDDDAEB), card = Color(0xFFFFFFFF), destructive = Color(0x00000000)),
    )
    val dark: Map<Area, AreaPalette> = mapOf(
        Area.Personal to AreaPalette(surface = Color(0xFF1F1410), ink = Color(0xFFF4E2D4), accent = Color(0xFFD67852), rule = Color(0xFF4A2E22), soft = Color(0xFF3D2218), card = Color(0xFF3D2218), destructive = Color(0xFFE07583)),
        Area.Work to AreaPalette(surface = Color(0xFF0E1A2B), ink = Color(0xFFDDE5F0), accent = Color(0xFF7A9BC2), rule = Color(0xFF1F2D3F), soft = Color(0xFF1A2638), card = Color(0xFF1A2638), destructive = Color(0xFFD87560)),
        Area.AiChat to AreaPalette(surface = Color(0xFF0F1614), ink = Color(0xFFECEAE3), accent = Color(0xFF7CB293), rule = Color(0xFF2A3530), soft = Color(0xFF1A2320), card = Color(0xFF1A2320), destructive = Color(0x00000000)),
        Area.Scratchpad to AreaPalette(surface = Color(0xFF1F1B12), ink = Color(0xFFE0D8C2), accent = Color(0xFFD6A57E), rule = Color(0xFF3D3528), soft = Color(0xFF272219), card = Color(0xFF272219), destructive = Color(0xFFD87560)),
        Area.Routines to AreaPalette(surface = Color(0xFF1A2218), ink = Color(0xFFD6DDD2), accent = Color(0xFF7CAB89), rule = Color(0xFF2D3829), soft = Color(0xFF243029), card = Color(0xFF243029), destructive = Color(0xFFD87560)),
        Area.Affirmations to AreaPalette(surface = Color(0xFF2E251A), ink = Color(0xFFEADCC2), accent = Color(0xFFC49B6F), rule = Color(0xFF4A3D28), soft = Color(0xFF3A2E1E), card = Color(0xFF3A2E1E), destructive = Color(0xFFD87560)),
        Area.Voice to AreaPalette(surface = Color(0xFF0F1614), ink = Color(0xFFFFFFFF), accent = Color(0xFF5E8B73), rule = Color(0xFF1C2624), soft = Color(0xFF1A2421), card = Color(0xFF1A2421), destructive = Color(0x00000000)),
        Area.System to AreaPalette(surface = Color(0xFF000000), ink = Color(0xFFFFFFFF), accent = Color(0xFF7CB293), rule = Color(0xFF2C2C2E), soft = Color(0xFF1C1C1E), card = Color(0xFF1C1C1E), destructive = Color(0x00000000)),
        Area.PrMonitor to AreaPalette(surface = Color(0xFF221C3F), ink = Color(0xFFE0DCEB), accent = Color(0xFF8478D8), rule = Color(0xFF3F365B), soft = Color(0xFF2D274E), card = Color(0xFF2D274E), destructive = Color(0x00000000)),
    )
    val widgetSurfaceLight = Color(0xFFF5EFE0)
    val widgetSurfaceDark = Color(0xFF1F1B12)
}
