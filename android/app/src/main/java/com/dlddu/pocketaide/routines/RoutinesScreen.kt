package com.dlddu.pocketaide.routines

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.outlined.CalendarToday
import androidx.compose.material.icons.outlined.RadioButtonUnchecked
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.material3.pulltorefresh.PullToRefreshBox
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.semantics.stateDescription
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.ScreenHeader
import com.dlddu.pocketaide.designsystem.components.SwipeToDelete
import kotlinx.coroutines.launch
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.util.Locale

private val AREA = Area.Routines
private val HEADER_TRAILING_GAP = 6.dp
private val ADD_ICON_SIZE = 11.dp
private val ADD_VERTICAL = 7.dp
private val TODAY_ROW_VERTICAL = 6.dp
private val RESTING_ROW_VERTICAL = DesignTokens.Spacing.xs
private val CARD_TITLE_GAP = 2.dp
private val BORDER_WIDTH = 1.dp
private val COMPLETED_BORDER_WIDTH = 2.dp
private val PROGRESS_HEIGHT = 6.dp
private val STEP_ICON_SIZE = 18.dp
private val RESTING_PADDING = 14.dp
private val BADGE_TRACKING = 0.55.sp
private val BADGE_VERTICAL = 2.dp
private const val CHECKED_OPACITY = 0.5f
private const val IDLE_PROGRESS_OPACITY = 0.4f
private const val IDLE_WEEKDAY_OPACITY = 0.4f

private sealed interface RoutineSheetMode {
    data object Add : RoutineSheetMode
    data class AddStep(val routine: Routine) : RoutineSheetMode
    data class History(val routine: Routine, val days: List<RoutineHistoryDay>) : RoutineSheetMode
}

object RoutinesCopy {
    private val weekdayFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("M월 d일 · EEEE", Locale.KOREAN)

    fun todayLabel(date: LocalDate): String = date.format(weekdayFormatter)

    fun progressLabel(routine: RoutineDay): String = when {
        routine.completed -> "완료"
        routine.done == 0 -> "대기"
        else -> "${routine.progressPercent}%"
    }

    fun cardSummary(routine: RoutineDay): String = "${routine.scheduleSummary} · ${routine.done} / ${routine.total} 완료"

    fun restingSummary(routine: Routine): String = "${routine.scheduleSummary} · 단계 ${routine.steps.size}개"
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RoutinesScreen(api: ApiClient) {
    val viewModel: RoutinesViewModel = viewModel(key = "routines") { RoutinesViewModel(api) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var sheetMode by remember { mutableStateOf<RoutineSheetMode?>(null) }
    val scope = rememberCoroutineScope()
    val ink = DesignTokens.Colors.ink(AREA)

    LaunchedEffect(viewModel) { viewModel.load() }

    val openHistory: (Long) -> Unit = { id ->
        state.routine(id)?.let { routine ->
            scope.launch {
                val days = viewModel.history(id)
                sheetMode = RoutineSheetMode.History(routine, days)
            }
        }
    }

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(AREA))
            .testTag("routines.screen"),
    ) {
        Column(Modifier.fillMaxSize()) {
            ScreenHeader(
                area = AREA,
                title = "루틴",
                trailing = {
                    Column(
                        horizontalAlignment = Alignment.End,
                        verticalArrangement = Arrangement.spacedBy(HEADER_TRAILING_GAP),
                    ) {
                        AddButton(onClick = { sheetMode = RoutineSheetMode.Add })
                        Text(
                            RoutinesCopy.todayLabel(viewModel.today()),
                            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
                            color = ink.copy(alpha = 0.55f),
                            modifier = Modifier.testTag("routines.today"),
                        )
                    }
                },
            )
            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
                LazyColumn(Modifier.fillMaxSize()) {
                    state.errorMessage?.let { message ->
                        item(key = "error") {
                            RoutineErrorRow(
                                message = message,
                                canRetry = state.canRetry,
                                onRetry = viewModel::retry,
                                modifier = Modifier.padding(
                                    horizontal = DesignTokens.Spacing.xl,
                                    vertical = DesignTokens.Spacing.sm,
                                ),
                            )
                        }
                    }
                    if (state.routines.isEmpty()) {
                        item(key = "empty") { EmptyState(isLoading = state.isLoading) }
                    } else {
                        item(key = "today.header") { SectionTitle("오늘 · ${state.today.size}") }
                        items(state.today, key = { "today.${it.id}" }) { routine ->
                            val shape = RoundedCornerShape(DesignTokens.Radius.card)
                            SwipeToDelete(
                                area = AREA,
                                shape = shape,
                                onDelete = { viewModel.delete(routine.id) },
                                modifier = Modifier.padding(
                                    horizontal = DesignTokens.Spacing.xl,
                                    vertical = TODAY_ROW_VERTICAL,
                                ),
                            ) {
                                RoutineCard(
                                    routine = routine,
                                    onToggle = { step -> viewModel.toggle(step, routine) },
                                    onAddStep = {
                                        state.routine(routine.id)?.let { sheetMode = RoutineSheetMode.AddStep(it) }
                                    },
                                    onHistory = { openHistory(routine.id) },
                                )
                            }
                        }
                        val resting = state.restingToday
                        if (resting.isNotEmpty()) {
                            item(key = "resting.header") { SectionTitle("오늘 쉬는 루틴") }
                            items(resting, key = { "resting.${it.id}" }) { routine ->
                                val shape = RoundedCornerShape(DesignTokens.Radius.card)
                                SwipeToDelete(
                                    area = AREA,
                                    shape = shape,
                                    onDelete = { viewModel.delete(routine.id) },
                                    modifier = Modifier.padding(
                                        horizontal = DesignTokens.Spacing.xl,
                                        vertical = RESTING_ROW_VERTICAL,
                                    ),
                                ) {
                                    RestingRow(
                                        routine = routine,
                                        daysUntilNext = RoutineSchedule.daysUntilNext(routine, state.dayKey),
                                        onClick = { openHistory(routine.id) },
                                    )
                                }
                            }
                        }
                        item(key = "bottom") { Box(Modifier.height(DesignTokens.Spacing.xxl)) }
                    }
                }
            }
        }
    }

    when (val mode = sheetMode) {
        RoutineSheetMode.Add -> RoutineAddSheet(
            startDay = state.dayKey,
            onSave = { draft ->
                sheetMode = null
                viewModel.create(draft)
            },
            onCancel = { sheetMode = null },
        )
        is RoutineSheetMode.AddStep -> RoutineStepAddSheet(
            routine = mode.routine,
            onSave = { title ->
                sheetMode = null
                viewModel.addStep(mode.routine.id, title)
            },
            onDeleteStep = { step ->
                sheetMode = null
                viewModel.deleteStep(step.id, mode.routine.id)
            },
            onCancel = { sheetMode = null },
        )
        is RoutineSheetMode.History -> RoutineHistorySheet(
            routine = mode.routine,
            days = mode.days,
            onClose = { sheetMode = null },
        )
        null -> Unit
    }
}

@Composable
private fun AddButton(onClick: () -> Unit) {
    val surface = DesignTokens.Colors.surface(AREA)
    Row(
        Modifier
            .clip(CircleShape)
            .background(DesignTokens.Colors.ink(AREA))
            .clickable(role = Role.Button, onClick = onClick)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = ADD_VERTICAL)
            .testTag("routines.add.button"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(Icons.Filled.Add, contentDescription = null, tint = surface, modifier = Modifier.size(ADD_ICON_SIZE))
        Text(
            "새 루틴",
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
            color = surface,
        )
    }
}

@Composable
private fun SectionTitle(text: String) {
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
        color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.7f),
        modifier = Modifier.padding(
            start = DesignTokens.Spacing.xl,
            end = DesignTokens.Spacing.xl,
            top = DesignTokens.Spacing.md,
            bottom = DesignTokens.Spacing.xs,
        ),
    )
}

@Composable
private fun EmptyState(isLoading: Boolean) {
    Box(
        Modifier
            .fillMaxWidth()
            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.xl),
        contentAlignment = if (isLoading) Alignment.Center else Alignment.CenterStart,
    ) {
        if (isLoading) {
            CircularProgressIndicator(color = DesignTokens.Colors.accent(AREA))
        } else {
            Text(
                "아직 루틴이 없습니다. 새 루틴으로 반복되는 하루의 흐름을 적어 보세요.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.body),
                color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.55f),
                modifier = Modifier.testTag("routines.empty.state"),
            )
        }
    }
}

@Composable
private fun RoutineCard(
    routine: RoutineDay,
    onToggle: (RoutineDayStep) -> Unit,
    onAddStep: () -> Unit,
    onHistory: () -> Unit,
) {
    val ink = DesignTokens.Colors.ink(AREA)
    val accent = DesignTokens.Colors.accent(AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    Column(
        Modifier
            .fillMaxWidth()
            .background(DesignTokens.Colors.surface(AREA), shape)
            .clip(shape)
            .background(DesignTokens.Colors.card(AREA))
            .border(
                if (routine.completed) COMPLETED_BORDER_WIDTH else BORDER_WIDTH,
                if (routine.completed) accent else DesignTokens.Colors.rule(AREA),
                shape,
            )
            .padding(DesignTokens.Spacing.lg)
            .testTag("routines.card.${routine.id}"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.Top) {
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(CARD_TITLE_GAP)) {
                Text(
                    routine.name,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.Bold),
                    color = ink,
                )
                Text(
                    RoutinesCopy.cardSummary(routine),
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
                    color = ink.copy(alpha = 0.55f),
                    modifier = Modifier.testTag("routines.card.${routine.id}.summary"),
                )
            }
            Text(
                RoutinesCopy.progressLabel(routine),
                style = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.Bold),
                color = if (!routine.completed && routine.done == 0) ink.copy(alpha = IDLE_PROGRESS_OPACITY) else accent,
                modifier = Modifier.testTag("routines.card.${routine.id}.progress"),
            )
        }
        Box(
            Modifier
                .fillMaxWidth()
                .height(PROGRESS_HEIGHT)
                .clip(CircleShape)
                .background(DesignTokens.Colors.soft(AREA)),
        ) {
            Box(
                Modifier
                    .fillMaxWidth(routine.progress.coerceIn(0f, 1f))
                    .fillMaxHeight()
                    .clip(CircleShape)
                    .background(accent),
            )
        }
        Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
            routine.steps.forEach { step -> StepRow(step = step, onClick = { onToggle(step) }) }
        }
        Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) {
            CardAction(
                icon = Icons.Filled.Add,
                label = "단계",
                tag = "routines.card.${routine.id}.steps.button",
                onClick = onAddStep,
            )
            CardAction(
                icon = Icons.Outlined.CalendarToday,
                label = "이력",
                tag = "routines.card.${routine.id}.history.button",
                onClick = onHistory,
            )
        }
    }
}

@Composable
private fun StepRow(step: RoutineDayStep, onClick: () -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .clickable(role = Role.Checkbox, onClick = onClick)
            .semantics { stateDescription = if (step.checked) "checked" else "unchecked" }
            .testTag("routines.step.${step.id}"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(
            if (step.checked) Icons.Filled.CheckCircle else Icons.Outlined.RadioButtonUnchecked,
            contentDescription = null,
            tint = if (step.checked) DesignTokens.Colors.accent(AREA) else DesignTokens.Colors.rule(AREA),
            modifier = Modifier.size(STEP_ICON_SIZE),
        )
        Text(
            step.title,
            style = DesignTokens.Typography.style(DesignTokens.Typography.body).copy(
                textDecoration = if (step.checked) TextDecoration.LineThrough else null,
            ),
            color = ink.copy(alpha = if (step.checked) CHECKED_OPACITY else 1f),
            modifier = Modifier.weight(1f),
        )
    }
}

@Composable
private fun CardAction(icon: ImageVector, label: String, tag: String, onClick: () -> Unit) {
    val color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.7f)
    Row(
        Modifier
            .clickable(role = Role.Button, onClick = onClick)
            .padding(vertical = DesignTokens.Spacing.xs)
            .testTag(tag),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(icon, contentDescription = null, tint = color, modifier = Modifier.size(SMALL_ICON_SIZE))
        Text(
            label,
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.SemiBold),
            color = color,
        )
    }
}

@Composable
private fun RestingRow(routine: Routine, daysUntilNext: Int?, onClick: () -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    val accent = DesignTokens.Colors.accent(AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    Column(
        Modifier
            .fillMaxWidth()
            .background(DesignTokens.Colors.surface(AREA), shape)
            .clip(shape)
            .background(DesignTokens.Colors.card(AREA))
            .border(BORDER_WIDTH, DesignTokens.Colors.rule(AREA), shape)
            .clickable(role = Role.Button, onClick = onClick)
            .padding(RESTING_PADDING)
            .testTag("routines.resting.${routine.id}"),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
    ) {
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(CARD_TITLE_GAP)) {
                Text(
                    routine.name,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.SemiBold),
                    color = ink,
                )
                Text(
                    RoutinesCopy.restingSummary(routine),
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.SemiBold),
                    color = ink.copy(alpha = 0.55f),
                )
            }
            if (daysUntilNext != null) {
                Text(
                    "D-$daysUntilNext",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs, FontWeight.Bold)
                        .copy(letterSpacing = BADGE_TRACKING),
                    color = accent,
                    modifier = Modifier
                        .clip(CircleShape)
                        .background(DesignTokens.Colors.soft(AREA))
                        .padding(horizontal = DesignTokens.Spacing.sm, vertical = BADGE_VERTICAL),
                )
            }
        }
        if (routine.cadence == RoutineCadence.WEEKDAYS || routine.cadence == RoutineCadence.WEEKLY) {
            WeekdayBar(routine.weekdays)
        }
    }
}

@Composable
private fun WeekdayBar(weekdays: Int) {
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
        RoutineWeekdays.displayOrder.forEach { index ->
            val scheduled = RoutineWeekdays.contains(weekdays, index)
            Text(
                RoutineWeekdays.symbols[index],
                style = DesignTokens.Typography.style(
                    DesignTokens.Typography.caption2xs,
                    if (scheduled) FontWeight.Bold else FontWeight.Normal,
                ),
                color = if (scheduled) {
                    DesignTokens.Colors.accent(AREA)
                } else {
                    DesignTokens.Colors.ink(AREA).copy(alpha = IDLE_WEEKDAY_OPACITY)
                },
                textAlign = TextAlign.Center,
                modifier = Modifier.weight(1f),
            )
        }
    }
}
