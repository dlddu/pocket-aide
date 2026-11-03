package com.dlddu.pocketaide.todos

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.DatePicker
import androidx.compose.material3.DatePickerDialog
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.rememberDatePickerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.FilterPills
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetCancelButton
import com.dlddu.pocketaide.designsystem.components.SheetDefaults
import com.dlddu.pocketaide.designsystem.components.SheetDestructiveButton
import com.dlddu.pocketaide.designsystem.components.SheetPrimaryButton
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneOffset

private const val MEMO_MAX_LINES = 4
private val ACTIONS_GAP = DesignTokens.Spacing.xs
private val ACTIONS_BOTTOM = DesignTokens.Spacing.xxl + DesignTokens.Spacing.sm

sealed interface TodoSheetMode {
    data object Create : TodoSheetMode
    data class Edit(val item: TodoItem) : TodoSheetMode
}

enum class TodoPriorityChoice(val wire: String, val priority: TodoPriority?) {
    UNSET("unset", null),
    HIGH("high", TodoPriority.HIGH),
    NORMAL("normal", TodoPriority.NORMAL),
    LOW("low", TodoPriority.LOW),
    ;

    val displayName: String get() = priority?.displayName ?: "없음"

    companion object {
        fun of(priority: TodoPriority?): TodoPriorityChoice = entries.firstOrNull { it.priority == priority } ?: UNSET
    }
}

@Composable
fun TodoEditSheet(
    area: TodoArea,
    mode: TodoSheetMode,
    onSave: (TodoDraft) -> Unit,
    onCancel: () -> Unit,
    onDelete: (() -> Unit)? = null,
) {
    val tone = area.designArea
    val ink = DesignTokens.Colors.ink(tone)
    val existing = (mode as? TodoSheetMode.Edit)?.item
    val isEditing = existing != null
    val parsedDue = existing?.dueDate?.let { TodoDueDate.date(it) }
    var title by remember(mode) { mutableStateOf(existing?.title ?: "") }
    var memo by remember(mode) { mutableStateOf(existing?.memo ?: "") }
    var hasDueDate by remember(mode) { mutableStateOf(parsedDue != null) }
    var dueDate by remember(mode) { mutableStateOf(parsedDue ?: LocalDate.now()) }
    var priority by remember(mode) { mutableStateOf(TodoPriorityChoice.of(existing?.priority)) }
    var showsPicker by remember(mode) { mutableStateOf(false) }
    val done = existing?.isDone ?: false
    val trimmedTitle = title.trim()

    fun handleSave() {
        if (trimmedTitle.isEmpty()) return
        onSave(
            TodoDraft(
                title = trimmedTitle,
                memo = memo.trim(),
                dueDate = if (hasDueDate) TodoDueDate.string(dueDate) else null,
                priority = priority.priority,
                done = done,
            ),
        )
    }

    Sheet(area = tone, onClose = onCancel) {
        Column(
            Modifier.fillMaxWidth().padding(horizontal = SheetDefaults.horizontalPadding),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            Text(
                if (isEditing) "할 일 편집" else "새 할 일",
                style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold),
                color = ink,
                modifier = Modifier.padding(top = SheetDefaults.titleTopPadding).testTag("todo.sheet.title"),
            )
            Card(area = tone, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
                Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) {
                    PlainTextField(
                        value = title,
                        onValueChange = { title = it },
                        placeholder = "제목",
                        textStyle = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.SemiBold),
                        textColor = ink,
                        placeholderColor = ink.copy(alpha = 0.45f),
                        singleLine = true,
                        modifier = Modifier.fillMaxWidth().testTag("todo.sheet.title.field"),
                    )
                    PlainTextField(
                        value = memo,
                        onValueChange = { memo = it },
                        placeholder = "메모",
                        textStyle = DesignTokens.Typography.style(DesignTokens.Typography.body),
                        textColor = ink,
                        placeholderColor = ink.copy(alpha = 0.45f),
                        maxLines = MEMO_MAX_LINES,
                        modifier = Modifier.fillMaxWidth().testTag("todo.sheet.memo.field"),
                    )
                    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            "마감일",
                            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Medium),
                            color = ink,
                        )
                        Spacer(Modifier.weight(1f))
                        Switch(
                            checked = hasDueDate,
                            onCheckedChange = { hasDueDate = it },
                            colors = SwitchDefaults.colors(
                                checkedTrackColor = DesignTokens.Colors.accent(tone),
                                checkedThumbColor = Color.White,
                            ),
                            modifier = Modifier.testTag("todo.sheet.due.toggle"),
                        )
                    }
                    if (hasDueDate) {
                        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                            Text(
                                "날짜",
                                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                                color = ink,
                            )
                            Spacer(Modifier.weight(1f))
                            Text(
                                TodoDueDate.string(dueDate),
                                style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Medium),
                                color = DesignTokens.Colors.accent(tone),
                                modifier = Modifier
                                    .clickable(role = Role.Button) { showsPicker = true }
                                    .padding(vertical = DesignTokens.Spacing.xs)
                                    .testTag("todo.sheet.due.picker"),
                            )
                        }
                    }
                }
            }
            Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
                AreaLabel(area = tone, text = "우선순위 (선택)")
                FilterPills(
                    area = tone,
                    options = TodoPriorityChoice.entries,
                    selection = priority,
                    onSelect = { priority = it },
                    optionTag = { it.wire },
                ) { option, active ->
                    Text(
                        option.displayName,
                        style = DesignTokens.Typography.style(
                            DesignTokens.Typography.body,
                            if (active) FontWeight.Bold else FontWeight.Medium,
                        ),
                    )
                }
            }
            Column(
                Modifier.padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM),
                verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP),
            ) {
                SheetPrimaryButton(
                    area = tone,
                    label = "저장",
                    enabled = trimmedTitle.isNotEmpty(),
                    tag = "todo.sheet.save.button",
                    onClick = { handleSave() },
                )
                if (isEditing && onDelete != null) {
                    SheetDestructiveButton(
                        area = tone,
                        label = "삭제",
                        tag = "todo.sheet.delete.button",
                        onClick = onDelete,
                    )
                }
                SheetCancelButton(area = tone, tag = "todo.sheet.cancel.button", onClick = onCancel)
            }
        }
    }

    if (showsPicker) {
        DueDatePickerDialog(
            initial = dueDate,
            onPick = {
                dueDate = it
                showsPicker = false
            },
            onDismiss = { showsPicker = false },
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun DueDatePickerDialog(initial: LocalDate, onPick: (LocalDate) -> Unit, onDismiss: () -> Unit) {
    val state = rememberDatePickerState(
        initialSelectedDateMillis = initial.atStartOfDay(ZoneOffset.UTC).toInstant().toEpochMilli(),
    )
    DatePickerDialog(
        onDismissRequest = onDismiss,
        confirmButton = {
            TextButton(
                onClick = {
                    val millis = state.selectedDateMillis
                    if (millis == null) {
                        onDismiss()
                    } else {
                        onPick(Instant.ofEpochMilli(millis).atZone(ZoneOffset.UTC).toLocalDate())
                    }
                },
            ) {
                Text("확인")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("취소")
            }
        },
    ) {
        DatePicker(state = state)
    }
}
