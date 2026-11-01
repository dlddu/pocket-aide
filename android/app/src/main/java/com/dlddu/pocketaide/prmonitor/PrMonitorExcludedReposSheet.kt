package com.dlddu.pocketaide.prmonitor

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Delete
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.PlainTextField

private val DELETE_ICON_SIZE = 18.dp
private val ROW_RULE_HEIGHT = 1.dp

@Composable
fun PrMonitorExcludedReposSheet(
    state: PrMonitorUiState,
    onAdd: (String) -> Unit,
    onRemove: (Long) -> Unit,
    onDismiss: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(PR_AREA)
    val rule = DesignTokens.Colors.rule(PR_AREA)
    var newRepo by remember { mutableStateOf("") }
    val submit = {
        val value = newRepo.trim()
        if (value.isNotEmpty()) {
            newRepo = ""
            onAdd(value)
        }
    }

    PrMonitorSheetScaffold(title = "제외한 레포", onDone = onDismiss) {
        Row(
            Modifier
                .fillMaxWidth()
                .padding(horizontal = DesignTokens.Spacing.xl)
                .padding(top = DesignTokens.Spacing.lg, bottom = DesignTokens.Spacing.md),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            val shape = RoundedCornerShape(PrMonitorStyle.controlRadius)
            PlainTextField(
                value = newRepo,
                onValueChange = { newRepo = it },
                placeholder = "owner/repo",
                textStyle = DesignTokens.Typography.style(DesignTokens.Typography.body),
                textColor = ink,
                placeholderColor = ink.copy(alpha = PrMonitorStyle.TERTIARY_OPACITY),
                singleLine = true,
                keyboardOptions = KeyboardOptions(
                    capitalization = KeyboardCapitalization.None,
                    autoCorrectEnabled = false,
                    keyboardType = KeyboardType.Uri,
                    imeAction = ImeAction.Done,
                ),
                modifier = Modifier
                    .weight(1f)
                    .clip(shape)
                    .background(DesignTokens.Colors.card(PR_AREA))
                    .border(PrMonitorStyle.borderWidth, rule, shape)
                    .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm)
                    .testTag("prmonitor.excluded.input"),
            )
            Text(
                "추가",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                color = Color.White,
                modifier = Modifier
                    .clip(shape)
                    .background(DesignTokens.Colors.accent(PR_AREA))
                    .clickable(role = Role.Button, onClick = submit)
                    .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm)
                    .testTag("prmonitor.excluded.add.button"),
            )
        }
        state.excludedRepoError?.let { message ->
            Text(
                message,
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = DesignTokens.StatusColors.failure(),
                modifier = Modifier.padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.xs),
            )
        }
        LazyColumn(Modifier.fillMaxWidth().weight(1f)) {
            if (state.excludedRepos.isEmpty() && !state.isLoadingExcluded) {
                item(key = "empty") {
                    Text(
                        "제외한 레포가 없습니다",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                        color = ink.copy(alpha = PrMonitorStyle.MUTED_OPACITY),
                        modifier = Modifier.padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.md),
                    )
                }
            }
            items(state.excludedRepos, key = { it.id }) { repo ->
                Row(
                    Modifier
                        .fillMaxWidth()
                        .padding(start = DesignTokens.Spacing.xl, end = DesignTokens.Spacing.md)
                        .padding(vertical = DesignTokens.Spacing.sm)
                        .testTag("prmonitor.excluded.row.${repo.id}"),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        repo.repoFullName,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                        color = ink,
                        modifier = Modifier.weight(1f),
                    )
                    Box(
                        Modifier
                            .clip(RoundedCornerShape(PrMonitorStyle.controlRadius))
                            .clickable(role = Role.Button) { onRemove(repo.id) }
                            .padding(DesignTokens.Spacing.sm)
                            .testTag("prmonitor.excluded.row.${repo.id}.delete"),
                    ) {
                        Icon(
                            Icons.Outlined.Delete,
                            contentDescription = "삭제",
                            tint = DesignTokens.StatusColors.failure(),
                            modifier = Modifier.size(DELETE_ICON_SIZE),
                        )
                    }
                }
                Box(
                    Modifier
                        .fillMaxWidth()
                        .padding(start = DesignTokens.Spacing.xl)
                        .height(ROW_RULE_HEIGHT)
                        .background(rule),
                )
            }
        }
    }
}
