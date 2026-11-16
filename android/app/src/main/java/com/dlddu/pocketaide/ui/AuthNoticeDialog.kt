package com.dlddu.pocketaide.ui

import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.testTag
import com.dlddu.pocketaide.auth.AuthNotice
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens

@Composable
fun AuthNoticeDialog(notice: AuthNotice, onDismiss: () -> Unit) {
    val area = Area.AiChat
    val ink = DesignTokens.Colors.ink(area)
    AlertDialog(
        onDismissRequest = onDismiss,
        containerColor = DesignTokens.Colors.surface(area),
        titleContentColor = ink,
        textContentColor = ink.copy(alpha = BODY_OPACITY),
        title = { Text(notice.kind.title) },
        text = {
            SelectionContainer {
                Text(notice.reason.message(), modifier = Modifier.testTag("AuthNotice.message"))
            }
        },
        confirmButton = {
            TextButton(onClick = onDismiss, modifier = Modifier.testTag("AuthNotice.dismiss")) {
                Text("확인", color = DesignTokens.Colors.accent(area))
            }
        },
        modifier = Modifier.testTag("AuthNotice"),
    )
}

private const val BODY_OPACITY = 0.7f
