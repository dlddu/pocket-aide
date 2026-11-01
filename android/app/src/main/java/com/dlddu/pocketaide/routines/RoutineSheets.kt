package com.dlddu.pocketaide.routines

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.Remove
import androidx.compose.material.icons.outlined.RemoveCircleOutline
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.contentDescription
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.semantics.stateDescription
import androidx.compose.ui.semantics.testTag as semanticsTestTag
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.FilterPills
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.Sheet
import com.dlddu.pocketaide.designsystem.components.SheetCancelButton
import com.dlddu.pocketaide.designsystem.components.SheetDefaults
import com.dlddu.pocketaide.designsystem.components.SheetPrimaryButton

private val AREA = Area.Routines
private val TITLE_TRACKING = (-0.5).sp
private val ACTIONS_GAP = DesignTokens.Spacing.xs
private val ACTIONS_BOTTOM = DesignTokens.Spacing.xxl + DesignTokens.Spacing.sm
private val HISTORY_CLOSE_BOTTOM = DesignTokens.Spacing.xxl
private val WEEKDAY_SIZE = 34.dp
private val WEEKDAY_GAP = 6.dp
private val BORDER_WIDTH = 1.dp
private val STEPPER_CELL_WIDTH = 44.dp
private val STEPPER_CELL_HEIGHT = 32.dp
private val STEPPER_RADIUS = 9.dp
private const val STEPPER_DISABLED_OPACITY = 0.4f
private const val MONTH_DAY_MIN = 1
private const val MONTH_DAY_MAX = 31
private val STEP_INDEX_WIDTH = 18.dp
private val STEP_FIELD_VERTICAL = 10.dp
private val STEP_FIELD_RADIUS = 12.dp
private val HEADER_GAP = 2.dp
private val HISTORY_ROW_GAP = 6.dp
private val HEATMAP_GAP = 4.dp
private val HEATMAP_RADIUS = 4.dp
private const val HEATMAP_COLUMNS = 10
private const val HEATMAP_PARTIAL_OPACITY = 0.4f
internal val SMALL_ICON_SIZE = 14.dp

@Composable
fun RoutineAddSheet(startDay: String, onSave: (RoutineDraft) -> Unit, onCancel: () -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    var name by remember { mutableStateOf("") }
    var cadence by remember { mutableStateOf(RoutineCadence.DAILY) }
    var weekdays by remember { mutableIntStateOf(0) }
    var monthDay by remember { mutableIntStateOf(MONTH_DAY_MIN) }
    val steps = remember { mutableStateListOf("") }
    val trimmedName = name.trim()
    val canSave = trimmedName.isNotEmpty() && when (cadence) {
        RoutineCadence.DAILY, RoutineCadence.MONTHLY -> true
        RoutineCadence.WEEKDAYS -> weekdays != 0
        RoutineCadence.WEEKLY -> Integer.bitCount(weekdays) == 1
    }

    fun selectCadence(next: RoutineCadence) {
        if (next == RoutineCadence.WEEKLY && Integer.bitCount(weekdays) != 1) weekdays = 0
        cadence = next
    }

    fun toggleWeekday(index: Int) {
        val bit = RoutineWeekdays.bit(index)
        weekdays = if (cadence == RoutineCadence.WEEKLY) bit else weekdays xor bit
    }

    fun handleSave() {
        if (!canSave) return
        onSave(
            RoutineDraft(
                name = trimmedName,
                cadence = cadence,
                weekdays = if (cadence == RoutineCadence.DAILY || cadence == RoutineCadence.MONTHLY) 0 else weekdays,
                monthDay = if (cadence == RoutineCadence.MONTHLY) monthDay else 0,
                startDay = startDay,
                steps = steps.map { it.trim() }.filter { it.isNotEmpty() },
            ),
        )
    }

    Sheet(area = AREA, onClose = onCancel) {
        Column(
            Modifier.fillMaxWidth().padding(horizontal = SheetDefaults.horizontalPadding),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            SheetTitle("새 루틴", tag = "routines.sheet.title")
            Card(area = AREA, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
                PlainTextField(
                    value = name,
                    onValueChange = { name = it },
                    placeholder = "루틴 이름 — 예: 아침 루틴",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = 0.45f),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth().testTag("routines.sheet.name.field"),
                )
            }
            Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
                SectionLabel("반복 주기")
                FilterPills(
                    area = AREA,
                    options = RoutineCadence.entries,
                    selection = cadence,
                    onSelect = ::selectCadence,
                    optionTag = { it.wire },
                    modifier = Modifier.testTag("routines.sheet.cadence.picker"),
                ) { option, active ->
                    Text(
                        option.displayName,
                        style = DesignTokens.Typography.style(
                            DesignTokens.Typography.bodySm,
                            if (active) FontWeight.Bold else FontWeight.Medium,
                        ),
                        maxLines = 1,
                    )
                }
                when (cadence) {
                    RoutineCadence.DAILY -> Unit
                    RoutineCadence.WEEKDAYS, RoutineCadence.WEEKLY ->
                        WeekdayPicker(weekdays = weekdays, onToggle = ::toggleWeekday)
                    RoutineCadence.MONTHLY ->
                        MonthDayStepper(monthDay = monthDay, onChange = { monthDay = it })
                }
            }
            Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
                SectionLabel("단계")
                steps.forEachIndexed { index, step ->
                    Row(
                        Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(STEP_FIELD_RADIUS))
                            .background(DesignTokens.Colors.card(AREA))
                            .padding(horizontal = DesignTokens.Spacing.md, vertical = STEP_FIELD_VERTICAL),
                        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text(
                            "${index + 1}",
                            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
                            color = ink.copy(alpha = 0.55f),
                            textAlign = TextAlign.Center,
                            modifier = Modifier.width(STEP_INDEX_WIDTH),
                        )
                        PlainTextField(
                            value = step,
                            onValueChange = { steps[index] = it },
                            placeholder = "단계 이름",
                            textStyle = DesignTokens.Typography.style(DesignTokens.Typography.body),
                            textColor = ink,
                            placeholderColor = ink.copy(alpha = 0.45f),
                            singleLine = true,
                            modifier = Modifier.weight(1f).testTag("routines.sheet.step.$index"),
                        )
                    }
                }
                Row(
                    Modifier
                        .clickable(role = Role.Button) { steps.add("") }
                        .padding(vertical = DesignTokens.Spacing.xs)
                        .testTag("routines.sheet.step.add.button"),
                    horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    val accent = DesignTokens.Colors.accent(AREA)
                    Icon(Icons.Filled.Add, contentDescription = null, tint = accent, modifier = Modifier.size(SMALL_ICON_SIZE))
                    Text(
                        "단계 추가",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
                        color = accent,
                    )
                }
            }
            Column(
                Modifier.padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM),
                verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP),
            ) {
                SheetPrimaryButton(
                    area = AREA,
                    label = "저장",
                    enabled = canSave,
                    tag = "routines.sheet.save.button",
                    onClick = ::handleSave,
                )
                SheetCancelButton(area = AREA, tag = "routines.sheet.cancel.button", onClick = onCancel)
            }
        }
    }
}

@Composable
private fun WeekdayPicker(weekdays: Int, onToggle: (Int) -> Unit) {
    Row(horizontalArrangement = Arrangement.spacedBy(WEEKDAY_GAP)) {
        RoutineWeekdays.displayOrder.forEach { index ->
            val selected = RoutineWeekdays.contains(weekdays, index)
            Box(
                Modifier
                    .size(WEEKDAY_SIZE)
                    .clip(CircleShape)
                    .background(if (selected) DesignTokens.Colors.accent(AREA) else DesignTokens.Colors.card(AREA))
                    .border(BORDER_WIDTH, DesignTokens.Colors.rule(AREA), CircleShape)
                    .clickable(role = Role.Checkbox) { onToggle(index) }
                    .testTag("routines.sheet.weekday.$index"),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    RoutineWeekdays.symbols[index],
                    style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
                    color = if (selected) DesignTokens.Colors.surface(AREA) else DesignTokens.Colors.ink(AREA),
                )
            }
        }
    }
}

@Composable
private fun MonthDayStepper(monthDay: Int, onChange: (Int) -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    val label = "매월 ${monthDay}일"
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
        Text(label, style = DesignTokens.Typography.style(DesignTokens.Typography.body), color = ink)
        Spacer(Modifier.weight(1f))
        Row(
            Modifier
                .clip(RoundedCornerShape(STEPPER_RADIUS))
                .background(DesignTokens.Colors.soft(AREA))
                .semantics { stateDescription = label }
                .testTag("routines.sheet.monthday.stepper"),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            StepperCell(Icons.Filled.Remove, enabled = monthDay > MONTH_DAY_MIN) { onChange(monthDay - 1) }
            Box(
                Modifier
                    .width(BORDER_WIDTH)
                    .height(STEPPER_CELL_HEIGHT)
                    .background(DesignTokens.Colors.rule(AREA)),
            )
            StepperCell(Icons.Filled.Add, enabled = monthDay < MONTH_DAY_MAX) { onChange(monthDay + 1) }
        }
    }
}

@Composable
private fun StepperCell(icon: ImageVector, enabled: Boolean, onClick: () -> Unit) {
    Box(
        Modifier
            .size(width = STEPPER_CELL_WIDTH, height = STEPPER_CELL_HEIGHT)
            .alpha(if (enabled) 1f else STEPPER_DISABLED_OPACITY)
            .clickable(enabled = enabled, role = Role.Button, onClick = onClick),
        contentAlignment = Alignment.Center,
    ) {
        Icon(icon, contentDescription = null, tint = DesignTokens.Colors.ink(AREA))
    }
}

@Composable
fun RoutineStepAddSheet(
    routine: Routine,
    onSave: (String) -> Unit,
    onDeleteStep: (RoutineStep) -> Unit,
    onCancel: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(AREA)
    var title by remember(routine.id) { mutableStateOf("") }
    val trimmedTitle = title.trim()

    Sheet(area = AREA, onClose = onCancel) {
        Column(
            Modifier.fillMaxWidth().padding(horizontal = SheetDefaults.horizontalPadding),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            SheetTitle("${routine.name} · 단계", tag = "routines.steps.sheet.title")
            Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
                routine.steps.forEach { step ->
                    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
                        Text(
                            step.title,
                            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                            color = ink,
                            modifier = Modifier.weight(1f),
                        )
                        Icon(
                            Icons.Outlined.RemoveCircleOutline,
                            contentDescription = "삭제",
                            tint = DesignTokens.Colors.destructive(AREA),
                            modifier = Modifier
                                .clip(CircleShape)
                                .clickable(role = Role.Button) { onDeleteStep(step) }
                                .testTag("routines.steps.sheet.delete.${step.id}"),
                        )
                    }
                }
            }
            Card(area = AREA, padding = CardPadding.Medium, modifier = Modifier.fillMaxWidth()) {
                PlainTextField(
                    value = title,
                    onValueChange = { title = it },
                    placeholder = "새 단계",
                    textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg),
                    textColor = ink,
                    placeholderColor = ink.copy(alpha = 0.45f),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth().testTag("routines.steps.sheet.field"),
                )
            }
            Column(
                Modifier.padding(top = DesignTokens.Spacing.sm, bottom = ACTIONS_BOTTOM),
                verticalArrangement = Arrangement.spacedBy(ACTIONS_GAP),
            ) {
                SheetPrimaryButton(
                    area = AREA,
                    label = "단계 추가",
                    enabled = trimmedTitle.isNotEmpty(),
                    tag = "routines.steps.sheet.save.button",
                    onClick = { if (trimmedTitle.isNotEmpty()) onSave(trimmedTitle) },
                )
                SheetTextButton("닫기", tag = "routines.steps.sheet.cancel.button", onClick = onCancel)
            }
        }
    }
}

@Composable
fun RoutineHistorySheet(routine: Routine, days: List<RoutineHistoryDay>, onClose: () -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    val summary = RoutineHistorySummary.of(days)
    val summaryText = "${routine.scheduleSummary} · 예정 ${summary.scheduledDays}일 중 ${summary.completedDays}일 완료"
    val meta = DesignTokens.Typography.style(DesignTokens.Typography.captionXs)

    Sheet(area = AREA, onClose = onClose) {
        Column(
            Modifier.fillMaxWidth().padding(horizontal = SheetDefaults.horizontalPadding),
            verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg),
        ) {
            Column(
                Modifier.padding(top = SheetDefaults.titleTopPadding),
                verticalArrangement = Arrangement.spacedBy(HEADER_GAP),
            ) {
                Text(
                    "${routine.name} · ${days.size}일",
                    style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold),
                    color = ink,
                    modifier = Modifier.testTag("routines.history.title"),
                )
                Text(
                    summaryText,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
                    color = ink.copy(alpha = 0.55f),
                    modifier = Modifier.testTag("routines.history.summary"),
                )
                Text(
                    "지난날의 n/m도 지금의 단계 수를 기준으로 셉니다.",
                    style = meta,
                    color = ink.copy(alpha = 0.55f),
                    modifier = Modifier.testTag("routines.history.note"),
                )
            }
            Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg - DesignTokens.Spacing.sm)) {
                HistoryHeatmap(days)
                RoutineHistoryRange(days)
            }
            Column(verticalArrangement = Arrangement.spacedBy(HISTORY_ROW_GAP)) {
                days.asReversed().forEach { day ->
                    val title = RoutineDayFormat.title(day.day)
                    Row(
                        Modifier
                            .fillMaxWidth()
                            .clearAndSetSemantics {
                                contentDescription = "$title, ${day.statusLabel}"
                                semanticsTestTag = "routines.history.row.${day.day}"
                            },
                    ) {
                        Text(
                            title,
                            style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm, FontWeight.SemiBold),
                            color = ink,
                        )
                        Spacer(Modifier.weight(1f))
                        Text(
                            day.statusLabel,
                            style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
                            color = if (day.completed) DesignTokens.Colors.accent(AREA) else ink.copy(alpha = 0.55f),
                        )
                    }
                }
            }
            Box(Modifier.padding(bottom = HISTORY_CLOSE_BOTTOM)) {
                SheetTextButton("닫기", tag = "routines.history.close.button", onClick = onClose)
            }
        }
    }
}

@Composable
private fun HistoryHeatmap(days: List<RoutineHistoryDay>) {
    val accent = DesignTokens.Colors.accent(AREA)
    val soft = DesignTokens.Colors.soft(AREA)
    val rule = DesignTokens.Colors.rule(AREA)
    val shape = RoundedCornerShape(HEATMAP_RADIUS)
    Column(verticalArrangement = Arrangement.spacedBy(HEATMAP_GAP)) {
        days.chunked(HEATMAP_COLUMNS).forEach { row ->
            Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(HEATMAP_GAP)) {
                row.forEach { day ->
                    val fill = when {
                        !day.scheduled -> Color.Transparent
                        day.completed -> accent
                        day.done > 0 -> accent.copy(alpha = HEATMAP_PARTIAL_OPACITY)
                        else -> soft
                    }
                    Box(
                        Modifier
                            .weight(1f)
                            .aspectRatio(1f)
                            .clip(shape)
                            .background(fill)
                            .then(if (day.scheduled) Modifier else Modifier.border(BORDER_WIDTH, rule, shape))
                            .semantics { stateDescription = day.statusLabel }
                            .testTag("routines.history.cell.${day.day}"),
                    )
                }
                repeat(HEATMAP_COLUMNS - row.size) {
                    Spacer(Modifier.weight(1f))
                }
            }
        }
    }
}

@Composable
private fun SheetTitle(text: String, tag: String) {
    Text(
        text,
        style = DesignTokens.Typography.style(SheetDefaults.titleSize, FontWeight.Bold).copy(letterSpacing = TITLE_TRACKING),
        color = DesignTokens.Colors.ink(AREA),
        modifier = Modifier.padding(top = SheetDefaults.titleTopPadding).testTag(tag),
    )
}

@Composable
private fun SectionLabel(text: String) {
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
        color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.7f),
    )
}

@Composable
internal fun SheetTextButton(label: String, tag: String, onClick: () -> Unit) {
    Box(
        Modifier
            .fillMaxWidth()
            .clickable(role = Role.Button, onClick = onClick)
            .padding(vertical = SheetDefaults.cancelVerticalPadding)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
            color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.55f),
        )
    }
}
