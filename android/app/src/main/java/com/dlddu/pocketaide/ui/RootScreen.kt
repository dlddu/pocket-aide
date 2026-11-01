package com.dlddu.pocketaide.ui

import android.content.Intent
import android.provider.Settings
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.consumeWindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.NotificationsOff
import androidx.compose.material3.Badge
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.PocketAideApplication
import com.dlddu.pocketaide.auth.AuthBrowser
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.scratchpad.ScratchpadScreen
import com.dlddu.pocketaide.scratchpad.ScratchpadViewModel
import com.dlddu.pocketaide.todos.TodoArea
import com.dlddu.pocketaide.todos.TodoListScreen

@Composable
fun RootScreen(authBrowser: AuthBrowser, onRequestNotificationPermission: () -> Unit) {
    val auth: AuthViewModel = viewModel()
    val state by auth.state.collectAsStateWithLifecycle()
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { auth.refreshPushPermission() }

    if (!state.signedIn) {
        LoginScreen(state = state, canSignIn = auth.canSignIn, onSignIn = { auth.signIn(authBrowser) })
        return
    }
    LaunchedEffect(Unit) { onRequestNotificationPermission() }
    SignedInRoot(auth = auth, pushDenied = state.pushDenied)
}

@Composable
private fun SignedInRoot(auth: AuthViewModel, pushDenied: Boolean) {
    val context = LocalContext.current
    val container = (context.applicationContext as PocketAideApplication).container
    val scratchpad: ScratchpadViewModel = viewModel { ScratchpadViewModel(auth.api) }
    val unclassified by scratchpad.unclassifiedCount.collectAsStateWithLifecycle()
    var selected by rememberSaveable { mutableStateOf(RootTab.Affirmations) }

    val pendingLink by container.deepLinks.pending.collectAsStateWithLifecycle()
    LaunchedEffect(pendingLink) {
        pendingLink?.let {
            selected = RootTab.forDeepLink(it)
            container.deepLinks.consume(it)
        }
    }
    LaunchedEffect(Unit) { scratchpad.load() }
    BackHandler(enabled = selected.isInMore) { selected = RootTab.More }

    Scaffold(
        topBar = {
            if (pushDenied) {
                PushDeniedBanner(
                    onClick = {
                        context.startActivity(
                            Intent(Settings.ACTION_APP_NOTIFICATION_SETTINGS)
                                .putExtra(Settings.EXTRA_APP_PACKAGE, context.packageName),
                        )
                    },
                )
            }
        },
        bottomBar = {
            RootNavigationBar(selected = selected, onSelect = { selected = it })
        },
    ) { padding ->
        Box(Modifier.fillMaxSize().padding(padding).consumeWindowInsets(padding)) {
            when (selected) {
                RootTab.More -> MoreList(scratchpadBadge = unclassified, onSelect = { selected = it })
                RootTab.Scratchpad -> ScratchpadScreen(viewModel = scratchpad)
                RootTab.Personal -> TodoListScreen(area = TodoArea.PERSONAL, api = auth.api)
                RootTab.Work -> TodoListScreen(area = TodoArea.WORK, api = auth.api)
                RootTab.Chat -> PlaceholderScreen(Area.AiChat, "AI 채팅")
                else -> PlaceholderScreen(selected.area, selected.title)
            }
        }
    }
}

@Composable
private fun RootNavigationBar(selected: RootTab, onSelect: (RootTab) -> Unit) {
    val activeTab = if (selected.isInMore) RootTab.More else selected
    val activeTint = when (selected) {
        RootTab.More -> DesignTokens.Colors.ink(Area.System)
        else -> DesignTokens.Colors.accent(selected.area)
    }
    val idle = DesignTokens.Colors.ink(Area.System).copy(alpha = 0.5f)
    NavigationBar(containerColor = DesignTokens.Colors.surface(Area.System)) {
        RootTab.BOTTOM_TABS.forEach { tab ->
            NavigationBarItem(
                selected = tab == activeTab,
                onClick = { onSelect(tab) },
                icon = { Icon(tab.icon, contentDescription = null) },
                label = {
                    Text(
                        tab.title,
                        style = DesignTokens.Typography.style(
                            DesignTokens.Typography.caption2xs,
                            if (tab == activeTab) FontWeight.Bold else FontWeight.Normal,
                        ),
                    )
                },
                colors = NavigationBarItemDefaults.colors(
                    selectedIconColor = activeTint,
                    selectedTextColor = activeTint,
                    indicatorColor = Color.Transparent,
                    unselectedIconColor = idle,
                    unselectedTextColor = idle,
                ),
                modifier = Modifier.testTag("tab.${tab.name}"),
            )
        }
    }
}

@Composable
private fun MoreList(scratchpadBadge: Int, onSelect: (RootTab) -> Unit) {
    val area = Area.System
    Column(Modifier.fillMaxSize().background(DesignTokens.Colors.surface(area))) {
        Text(
            RootTab.More.title,
            style = DesignTokens.Typography.style(DesignTokens.Typography.h1, FontWeight.Bold),
            color = DesignTokens.Colors.ink(area),
            modifier = Modifier
                .statusBarsPadding()
                .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.md),
        )
        LazyColumn {
            items(RootTab.MORE_TABS) { tab ->
                Row(
                    Modifier
                        .fillMaxWidth()
                        .clickable { onSelect(tab) }
                        .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.md)
                        .testTag("more.${tab.name}"),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
                ) {
                    Icon(tab.icon, contentDescription = null, tint = DesignTokens.Colors.accent(tab.area))
                    Text(
                        tab.title,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                        color = DesignTokens.Colors.ink(area),
                        modifier = Modifier.weight(1f),
                    )
                    if (tab == RootTab.Scratchpad && scratchpadBadge > 0) {
                        Badge(containerColor = DesignTokens.Colors.accent(Area.PrMonitor), contentColor = Color.White) {
                            Text("$scratchpadBadge")
                        }
                    }
                    Icon(
                        Icons.AutoMirrored.Outlined.KeyboardArrowRight,
                        contentDescription = null,
                        tint = DesignTokens.Colors.ink(area).copy(alpha = 0.45f),
                    )
                }
                HorizontalDivider(color = DesignTokens.Colors.rule(area), modifier = Modifier.padding(start = DesignTokens.Spacing.xl))
            }
        }
    }
}

@Composable
private fun PushDeniedBanner(onClick: () -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .background(DesignTokens.Colors.accent(Area.Work).copy(alpha = 0.18f))
            .statusBarsPadding()
            .clickable(onClick = onClick)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm)
            .testTag("PushDeniedBanner"),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Icon(Icons.Outlined.NotificationsOff, contentDescription = null, tint = DesignTokens.Colors.ink(Area.Work))
        Text(
            "알림 권한이 꺼져 있어 PR 푸시가 도착하지 않습니다. 설정에서 켜기",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = DesignTokens.Colors.ink(Area.Work),
        )
    }
}
