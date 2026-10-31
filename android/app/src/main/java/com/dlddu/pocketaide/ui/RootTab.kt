package com.dlddu.pocketaide.ui

import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.Chat
import androidx.compose.material.icons.outlined.Autorenew
import androidx.compose.material.icons.outlined.Description
import androidx.compose.material.icons.outlined.FavoriteBorder
import androidx.compose.material.icons.outlined.MoreHoriz
import androidx.compose.material.icons.outlined.Notifications
import androidx.compose.material.icons.outlined.Person
import androidx.compose.material.icons.outlined.Terminal
import androidx.compose.material.icons.outlined.Verified
import androidx.compose.material.icons.outlined.VerifiedUser
import androidx.compose.material.icons.outlined.WorkOutline
import androidx.compose.ui.graphics.vector.ImageVector
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.push.PushDeepLink

enum class RootTab(val title: String, val icon: ImageVector, val area: Area) {
    Affirmations("다짐", Icons.Outlined.FavoriteBorder, Area.Affirmations),
    PrMonitor("PR 모니터", Icons.Outlined.Verified, Area.PrMonitor),
    Approval("승인", Icons.Outlined.VerifiedUser, Area.PrMonitor),
    Sessions("세션", Icons.Outlined.Terminal, Area.PrMonitor),
    More("더 보기", Icons.Outlined.MoreHoriz, Area.System),
    Chat("채팅", Icons.AutoMirrored.Outlined.Chat, Area.AiChat),
    Scratchpad("임시공간", Icons.Outlined.Description, Area.Scratchpad),
    Personal("개인", Icons.Outlined.Person, Area.Personal),
    Work("회사", Icons.Outlined.WorkOutline, Area.Work),
    Routines("루틴", Icons.Outlined.Autorenew, Area.Routines),
    Notifications("알림 기록", Icons.Outlined.Notifications, Area.PrMonitor);

    val isInMore: Boolean get() = this in MORE_TABS

    companion object {
        val BOTTOM_TABS = listOf(Affirmations, PrMonitor, Approval, Sessions, More)
        val MORE_TABS = listOf(Chat, Scratchpad, Personal, Work, Routines, Notifications)

        fun forDeepLink(link: PushDeepLink): RootTab = when (link) {
            is PushDeepLink.Approval -> Approval
            is PushDeepLink.ExternalNotification -> Notifications
            is PushDeepLink.PrMonitorEvent -> PrMonitor
        }
    }
}
