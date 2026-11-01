package com.dlddu.pocketaide.sessions

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowDropDown
import androidx.compose.material.icons.filled.RadioButtonChecked
import androidx.compose.material.icons.outlined.RadioButtonUnchecked
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.unit.dp
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetDefaults

private val FIELD_LABEL_GAP = 6.dp
private val TYPE_TEXT_GAP = 2.dp
private val RADIO_SIZE = 20.dp
private val RULE_HEIGHT = 1.dp
private val ACTIONS_TOP = 8.dp
private val ACTIONS_BOTTOM = 32.dp
private val ACTIONS_GAP = 4.dp
private val PRIMARY_VERTICAL = 14.dp
private val CANCEL_VERTICAL = 8.dp
private val INLINE_SPINNER_SIZE = 16.dp
private val INLINE_SPINNER_STROKE = 2.dp

@Composable
fun SessionCreateSheet(
    config: AgentSessionConfig?,
    configFailed: Boolean,
    onRetryConfig: () -> Unit,
    isCreating: Boolean,
    error: AgentSessionError?,
    onCreate: (String, AgentSessionWorkloadType, String?) -> Unit,
    onCancel: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    var name by remember { mutableStateOf("") }
    var pickedType by remember { mutableStateOf<String?>(null) }
    var pickedModel by remember { mutableStateOf<String?>(null) }
    var typedModel by remember { mutableStateOf("") }

    val workload = config?.let { c -> c.workloadTypes.firstOrNull { it.id == pickedType } ?: c.initialType }
    val trimmedName = name.trim()
    val canCreate = trimmedName.isNotEmpty() && workload != null && !isCreating
    val chosenModel: String? = workload?.model?.let { catalog ->
        catalog.initialPick?.let { pickedModel ?: it } ?: typedModel
    }

    Sheet(area = SESSIONS_AREA, onClose = onCancel) {
        Column(
            Modifier.fillMaxWidth().padding(horizontal = SheetDefaults.horizontalPadding),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            Text(
                "새 세션",
                style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold),
                color = ink,
                modifier = Modifier.padding(top = SheetDefaults.titleTopPadding).testTag("sessions.create.title"),
            )
            Field(label = "이름") {
                PlainTextField(
                    value = name,
                    onValueChange = { name = it },
                    placeholder = "예: 테스트 실패 원인 찾기",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = SessionStyle.FAINT_OPACITY),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth().testTag("sessions.create.name.field"),
                )
            }
            WorkloadPicker(
                config = config,
                configFailed = configFailed,
                selected = workload,
                onRetryConfig = onRetryConfig,
                onSelect = { kind ->
                    pickedType = kind.id
                    pickedModel = null
                    typedModel = ""
                },
            )
            workload?.model?.let { catalog ->
                Field(label = "모델") {
                    ModelInput(
                        catalog = catalog,
                        pickedModel = pickedModel,
                        typedModel = typedModel,
                        onPick = { pickedModel = it },
                        onType = { typedModel = it },
                    )
                }
            }
            Text(
                if (workload?.takesModel == true) {
                    "타입과 모델은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요."
                } else {
                    "타입은 만든 뒤에 바꿀 수 없어요. 세션마다 전용 환경이 하나씩 뜨고, 처음 뜨는 데 몇 초 걸릴 수 있어요."
                },
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = SessionStyle.HINT_OPACITY),
            )
            error?.let {
                Text(
                    createFailureMessage(it, workload),
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm),
                    color = DesignTokens.StatusColors.failure(),
                    modifier = Modifier.testTag("sessions.create.error"),
                )
            }
            Column(
                Modifier.padding(top = ACTIONS_TOP, bottom = ACTIONS_BOTTOM),
                verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP),
            ) {
                CreateButton(
                    isCreating = isCreating,
                    enabled = canCreate,
                    onClick = {
                        if (canCreate && workload != null) onCreate(trimmedName, workload, chosenModel)
                    },
                )
                Text(
                    "취소",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
                    color = ink.copy(alpha = SessionStyle.MUTED_OPACITY),
                    modifier = Modifier
                        .align(Alignment.CenterHorizontally)
                        .clickable(role = Role.Button, onClick = onCancel)
                        .padding(vertical = CANCEL_VERTICAL)
                        .testTag("sessions.create.cancel.button"),
                )
            }
        }
    }
}

@Composable
private fun Field(label: String, content: @Composable () -> Unit) {
    Column(verticalArrangement = Arrangement.spacedBy(FIELD_LABEL_GAP)) {
        FieldLabel(label)
        Card(area = SESSIONS_AREA, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
            Box(Modifier.fillMaxWidth()) { content() }
        }
    }
}

@Composable
private fun FieldLabel(label: String) {
    Text(
        label,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
        color = DesignTokens.Colors.ink(SESSIONS_AREA).copy(alpha = SessionStyle.LABEL_OPACITY),
    )
}

@Composable
private fun WorkloadPicker(
    config: AgentSessionConfig?,
    configFailed: Boolean,
    selected: AgentSessionWorkloadType?,
    onRetryConfig: () -> Unit,
    onSelect: (AgentSessionWorkloadType) -> Unit,
) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val accent = DesignTokens.Colors.accent(SESSIONS_AREA)
    val types = config?.workloadTypes.orEmpty()
    Column(verticalArrangement = Arrangement.spacedBy(FIELD_LABEL_GAP)) {
        FieldLabel("타입")
        Card(area = SESSIONS_AREA, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
            if (types.isNotEmpty()) {
                Column(Modifier.fillMaxWidth()) {
                    types.forEachIndexed { index, kind ->
                        if (index > 0) {
                            Box(Modifier.fillMaxWidth().height(RULE_HEIGHT).background(DesignTokens.Colors.rule(SESSIONS_AREA)))
                        }
                        val isSelected = kind.id == selected?.id
                        Row(
                            Modifier
                                .fillMaxWidth()
                                .selectable(selected = isSelected, role = Role.RadioButton) { onSelect(kind) }
                                .padding(vertical = DesignTokens.Spacing.sm)
                                .testTag("sessions.create.workload.${kind.id}"),
                            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
                            verticalAlignment = Alignment.Top,
                        ) {
                            Icon(
                                if (isSelected) Icons.Filled.RadioButtonChecked else Icons.Outlined.RadioButtonUnchecked,
                                contentDescription = null,
                                tint = if (isSelected) accent else ink.copy(alpha = NO_SELECTION_OPACITY),
                                modifier = Modifier.size(RADIO_SIZE),
                            )
                            Column(verticalArrangement = Arrangement.spacedBy(TYPE_TEXT_GAP)) {
                                Text(
                                    kind.displayName,
                                    style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold),
                                    color = ink,
                                )
                                kind.description?.takeIf { it.isNotEmpty() }?.let { summary ->
                                    Text(
                                        summary,
                                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                                        color = ink.copy(alpha = SessionStyle.HINT_OPACITY),
                                    )
                                }
                            }
                        }
                    }
                }
            } else {
                Row(
                    Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    if (config == null && !configFailed) {
                        CircularProgressIndicator(
                            color = accent,
                            strokeWidth = INLINE_SPINNER_STROKE,
                            modifier = Modifier.size(INLINE_SPINNER_SIZE),
                        )
                    }
                    Text(
                        when {
                            configFailed -> "세션 플랫폼에 연결하지 못해 타입 목록을 불러오지 못했어요."
                            config == null -> "타입 목록을 불러오는 중…"
                            else -> "세션 플랫폼이 제공하는 타입이 없어요."
                        },
                        style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                        color = ink.copy(alpha = SessionStyle.HINT_OPACITY),
                        modifier = Modifier.weight(1f),
                    )
                    if (configFailed) {
                        Text(
                            "다시 시도",
                            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
                            color = accent,
                            modifier = Modifier
                                .clickable(role = Role.Button, onClick = onRetryConfig)
                                .testTag("sessions.create.workload.retry"),
                        )
                    }
                }
            }
        }
    }
}

private const val NO_SELECTION_OPACITY = 0.4f

@Composable
private fun ModelInput(
    catalog: AgentSessionModelCatalog,
    pickedModel: String?,
    typedModel: String,
    onPick: (String) -> Unit,
    onType: (String) -> Unit,
) {
    val ink = DesignTokens.Colors.ink(SESSIONS_AREA)
    val initial = catalog.initialPick
    if (initial == null) {
        PlainTextField(
            value = typedModel,
            onValueChange = onType,
            placeholder = catalog.defaultModel.ifEmpty { "platform-default" },
            textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
            textColor = ink,
            placeholderColor = ink.copy(alpha = SessionStyle.FAINT_OPACITY),
            singleLine = true,
            keyboardOptions = KeyboardOptions(capitalization = KeyboardCapitalization.None, autoCorrectEnabled = false),
            modifier = Modifier.fillMaxWidth().testTag("sessions.create.model.field"),
        )
        return
    }
    var expanded by remember { mutableStateOf(false) }
    Box {
        Row(
            Modifier
                .fillMaxWidth()
                .clickable(role = Role.DropdownList) { expanded = true }
                .testTag("sessions.create.model.picker"),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                pickedModel ?: initial,
                style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                color = ink,
                modifier = Modifier.weight(1f),
            )
            Icon(Icons.Filled.ArrowDropDown, contentDescription = null, tint = ink)
        }
        DropdownMenu(
            expanded = expanded,
            onDismissRequest = { expanded = false },
            containerColor = DesignTokens.Colors.card(SESSIONS_AREA),
        ) {
            catalog.models.forEach { model ->
                DropdownMenuItem(
                    text = {
                        Text(
                            model,
                            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                            color = ink,
                        )
                    },
                    onClick = {
                        onPick(model)
                        expanded = false
                    },
                    modifier = Modifier.testTag("sessions.create.model.option.$model"),
                )
            }
        }
    }
}

@Composable
private fun CreateButton(isCreating: Boolean, enabled: Boolean, onClick: () -> Unit) {
    val surface = DesignTokens.Colors.surface(SESSIONS_AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .alpha(if (enabled) 1f else SessionStyle.DISABLED_OPACITY)
            .clip(CircleShape)
            .background(DesignTokens.Colors.ink(SESSIONS_AREA))
            .clickable(enabled = enabled, role = Role.Button, onClick = onClick)
            .padding(vertical = PRIMARY_VERTICAL)
            .testTag("sessions.create.save.button"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm, Alignment.CenterHorizontally),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (isCreating) {
            CircularProgressIndicator(
                color = surface,
                strokeWidth = INLINE_SPINNER_STROKE,
                modifier = Modifier.size(INLINE_SPINNER_SIZE),
            )
        }
        Text(
            if (isCreating) "세션 띄우는 중…" else "만들기",
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold),
            color = surface,
        )
    }
}

internal fun createFailureMessage(error: AgentSessionError, workload: AgentSessionWorkloadType?): String =
    if (error == AgentSessionError.UNKNOWN) {
        val name = workload?.displayName ?: "선택한"
        "세션을 띄우지 못했어요. 세션 플랫폼에 「$name」 타입 설정(이미지·자격증명)이 있는지 확인해 주세요."
    } else {
        error.message
    }
