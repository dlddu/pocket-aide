package com.dlddu.pocketaide.approvals

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.text.selection.SelectionContainer
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetDefaults

@Composable
internal fun ApprovalKeyIssueSheet(state: ApprovalKeysUiState, onIssue: (String) -> Unit, onClose: () -> Unit) {
    var name by rememberSaveable { mutableStateOf("") }
    var copied by remember { mutableStateOf(false) }
    val context = LocalContext.current
    Sheet(area = APPROVAL_AREA, onClose = onClose) {
        Column(
            Modifier
                .fillMaxWidth()
                .padding(
                    start = SheetDefaults.horizontalPadding,
                    end = SheetDefaults.horizontalPadding,
                    top = SheetDefaults.titleTopPadding,
                    bottom = DesignTokens.Spacing.xxl,
                ),
        ) {
            val issued = state.issued
            if (issued == null) {
                SheetTitle("호출자 키 발급")
                FieldLabel("호출자 이름")
                MonoField(value = name, onValueChange = { name = it }, placeholder = "infra-ci", tag = "approval.keys.issue.name")
                state.issueNotice?.let { FailureNotice(it, "approval.keys.issue.notice") }
                Row(
                    Modifier.padding(top = DesignTokens.Spacing.lg),
                    horizontalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical),
                ) {
                    ApprovalButton("취소", primary = false, tag = "approval.keys.issue.cancel", modifier = Modifier.weight(1f)) {
                        onClose()
                    }
                    ApprovalButton(
                        "발급",
                        primary = true,
                        tag = "approval.keys.issue.submit",
                        enabled = ApprovalKeyFormatting.trimmedName(name) != null && !state.isIssuing,
                        modifier = Modifier.weight(1f),
                    ) { onIssue(name) }
                }
            } else {
                SheetTitle(ApprovalKeyFormatting.ISSUED_TITLE, "approval.keys.issued.title")
                FieldLabel("호출자 이름")
                ValueBox(issued.key.name, accent = false, tag = null)
                FieldLabel("키 원문")
                ValueBox(issued.secret, accent = true, tag = "approval.keys.issued.secret")
                Text(
                    ApprovalKeyFormatting.SECRET_WARNING,
                    style = DesignTokens.Typography.style(ApprovalMetrics.noticeSize).copy(lineHeight = ApprovalMetrics.noticeLineHeight),
                    color = DesignTokens.StatusColors.failure(),
                    modifier = Modifier.padding(top = DesignTokens.Spacing.sm),
                )
                Row(
                    Modifier.padding(top = DesignTokens.Spacing.lg),
                    horizontalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical),
                ) {
                    ApprovalButton(
                        if (copied) "복사됨" else "복사",
                        primary = true,
                        tag = "approval.keys.issued.copy",
                        modifier = Modifier.weight(1f),
                    ) {
                        copyToClipboard(context, issued.secret)
                        copied = true
                    }
                    ApprovalButton("완료", primary = false, tag = "approval.keys.issued.done", modifier = Modifier.weight(1f)) {
                        onClose()
                    }
                }
            }
        }
    }
}

@Composable
internal fun ApprovalServiceAccountSheet(
    state: ApprovalKeysUiState,
    onRegister: (name: String, namespace: String, account: String) -> Unit,
    onClose: () -> Unit,
) {
    var name by rememberSaveable { mutableStateOf("") }
    var namespace by rememberSaveable { mutableStateOf("") }
    var account by rememberSaveable { mutableStateOf("") }
    val canSubmit = ApprovalKeyFormatting.trimmedName(name) != null &&
        ApprovalKeyFormatting.trimmedName(namespace) != null &&
        ApprovalKeyFormatting.trimmedName(account) != null &&
        !state.isRegistering
    val accountRefused = state.registerOutcome == ApprovalServiceAccountOutcome.DuplicateServiceAccount
    Sheet(area = APPROVAL_AREA, onClose = onClose) {
        Column(
            Modifier
                .fillMaxWidth()
                .padding(
                    start = SheetDefaults.horizontalPadding,
                    end = SheetDefaults.horizontalPadding,
                    top = SheetDefaults.titleTopPadding,
                    bottom = DesignTokens.Spacing.xxl,
                ),
        ) {
            SheetTitle("k8s 서비스 어카운트 등록")
            FieldLabel("호출자 이름")
            MonoField(value = name, onValueChange = { name = it }, placeholder = "homelab-k3s-mcp", tag = "approval.keys.sa.name")
            Row(horizontalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical), verticalAlignment = Alignment.Top) {
                Column(Modifier.weight(1f)) {
                    FieldLabel("네임스페이스")
                    MonoField(
                        value = namespace,
                        onValueChange = { namespace = it },
                        placeholder = "homelab-k3s-mcp",
                        tag = "approval.keys.sa.namespace",
                    )
                }
                Column(Modifier.weight(1f)) {
                    FieldLabel("서비스 어카운트")
                    MonoField(
                        value = account,
                        onValueChange = { account = it },
                        placeholder = "homelab-k3s-mcp",
                        tag = "approval.keys.sa.account",
                        refused = accountRefused,
                    )
                }
            }
            state.registerNotice?.let { FailureNotice(it, "approval.keys.sa.notice") }
            Text(
                "이 서비스 어카운트로 도는 파드가 audience pocket-aide 토큰으로 요청합니다. 발급할 키 원문은 없습니다.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs).copy(lineHeight = ApprovalMetrics.noticeLineHeight),
                color = DesignTokens.Colors.ink(APPROVAL_AREA).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
                modifier = Modifier.padding(top = DesignTokens.Spacing.sm),
            )
            Row(
                Modifier.padding(top = DesignTokens.Spacing.lg),
                horizontalArrangement = Arrangement.spacedBy(ApprovalMetrics.bannerVertical),
            ) {
                ApprovalButton("취소", primary = false, tag = "approval.keys.sa.cancel", modifier = Modifier.weight(1f)) {
                    onClose()
                }
                ApprovalButton(
                    "등록",
                    primary = true,
                    tag = "approval.keys.sa.submit",
                    enabled = canSubmit,
                    modifier = Modifier.weight(1f),
                ) { onRegister(name, namespace, account) }
            }
        }
    }
}

private fun copyToClipboard(context: Context, text: String) {
    val manager = context.getSystemService(ClipboardManager::class.java) ?: return
    manager.setPrimaryClip(ClipData.newPlainText("pocketaide-approval-key", text))
}

@Composable
private fun SheetTitle(text: String, tag: String? = null) {
    Text(
        text,
        style = DesignTokens.Typography.style(ApprovalMetrics.sheetTitleSize, FontWeight.Bold),
        color = DesignTokens.Colors.ink(APPROVAL_AREA),
        modifier = if (tag != null) Modifier.testTag(tag) else Modifier,
    )
}

@Composable
private fun FieldLabel(text: String) {
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
        color = DesignTokens.Colors.ink(APPROVAL_AREA).copy(alpha = ApprovalMetrics.SECONDARY_OPACITY),
        modifier = Modifier.padding(top = DesignTokens.Spacing.md, bottom = DesignTokens.Spacing.xs),
    )
}

@Composable
private fun FailureNotice(text: String, tag: String) {
    Text(
        text,
        style = DesignTokens.Typography.style(ApprovalMetrics.noticeSize),
        color = DesignTokens.StatusColors.failure(),
        modifier = Modifier
            .padding(top = DesignTokens.Spacing.sm)
            .testTag(tag),
    )
}

@Composable
private fun MonoField(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    tag: String,
    refused: Boolean = false,
) {
    val area = APPROVAL_AREA
    val ink = DesignTokens.Colors.ink(area)
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    PlainTextField(
        value = value,
        onValueChange = onValueChange,
        placeholder = placeholder,
        textStyle = monoStyle(DesignTokens.Typography.body),
        textColor = ink,
        placeholderColor = ink.copy(alpha = ApprovalMetrics.FAINT_OPACITY),
        singleLine = true,
        keyboardOptions = KeyboardOptions(
            capitalization = KeyboardCapitalization.None,
            autoCorrectEnabled = false,
            keyboardType = KeyboardType.Ascii,
        ),
        modifier = Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(area))
            .border(
                ApprovalMetrics.ruleWidth,
                if (refused) DesignTokens.StatusColors.failure() else DesignTokens.Colors.rule(area),
                shape,
            )
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical)
            .testTag(tag),
    )
}

@Composable
private fun ValueBox(text: String, accent: Boolean, tag: String?) {
    val area = APPROVAL_AREA
    val shape = RoundedCornerShape(DesignTokens.Radius.key)
    SelectionContainer {
        Text(
            text,
            style = monoStyle(if (accent) DesignTokens.Typography.captionSm else DesignTokens.Typography.body),
            color = DesignTokens.Colors.ink(area),
            modifier = Modifier
                .fillMaxWidth()
                .clip(shape)
                .background(if (accent) DesignTokens.Colors.soft(area) else DesignTokens.Colors.card(area))
                .border(ApprovalMetrics.ruleWidth, if (accent) DesignTokens.Colors.accent(area) else DesignTokens.Colors.rule(area), shape)
                .padding(horizontal = DesignTokens.Spacing.md, vertical = ApprovalMetrics.bannerVertical)
                .then(if (tag != null) Modifier.testTag(tag) else Modifier),
        )
    }
}
