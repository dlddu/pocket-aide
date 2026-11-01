package com.dlddu.pocketaide.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

@Composable
fun LoginScreen(state: AuthUiState, canSignIn: Boolean, onSignIn: () -> Unit) {
    val area = Area.AiChat
    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(area))
            .safeDrawingPadding()
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.xxl),
    ) {
        Column(
            Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xxl),
        ) {
            Spacer(Modifier.weight(1f))
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            ) {
                Text(
                    "PocketAide",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.h1, FontWeight.Bold),
                    color = DesignTokens.Colors.ink(area),
                    modifier = Modifier.testTag("BrandTitle"),
                )
                Text(
                    "당신의 작은 비서",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    color = DesignTokens.Colors.accent(area),
                    modifier = Modifier.testTag("BrandTagline"),
                )
            }
            Spacer(Modifier.weight(1f))
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
            ) {
                Button(
                    onClick = onSignIn,
                    enabled = canSignIn && !state.isSigningIn,
                    colors = ButtonDefaults.buttonColors(containerColor = DesignTokens.Colors.accent(area)),
                    modifier = Modifier.fillMaxWidth().testTag("SignInButton"),
                ) {
                    Text(
                        if (state.isSigningIn) "Signing in…" else "Sign in with OIDC",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold),
                        modifier = Modifier.padding(vertical = DesignTokens.Spacing.xs),
                    )
                }
                state.signInError?.let {
                    Text(
                        it,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                        color = DesignTokens.Colors.accent(Area.Personal),
                        textAlign = TextAlign.Center,
                        modifier = Modifier.testTag("SignInError"),
                    )
                }
            }
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            ) {
                Box(
                    Modifier
                        .size(DesignTokens.Spacing.sm)
                        .background(
                            if (state.healthOk) DesignTokens.Colors.accent(Area.Routines) else DesignTokens.Colors.accent(Area.Personal),
                            CircleShape,
                        ),
                )
                Text(
                    if (state.healthOk) "Health: OK" else "Health: ${state.healthError ?: "checking…"}",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = DesignTokens.Colors.ink(area).copy(alpha = 0.7f),
                    modifier = Modifier.testTag("HealthStatus"),
                )
            }
        }
    }
}
