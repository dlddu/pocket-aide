package com.dlddu.pocketaide.affirmations

import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.outlined.Autorenew
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.clearAndSetSemantics
import androidx.compose.ui.semantics.onLongClick
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.Card
import com.dlddu.pocketaide.designsystem.components.CardPadding
import com.dlddu.pocketaide.designsystem.components.ScreenHeader
import com.dlddu.pocketaide.designsystem.components.SwipeToDelete
import com.dlddu.pocketaide.widget.WidgetRefresher

private val AREA = Area.Affirmations
private val ADD_BUTTON_SIZE = 36.dp
private val ADD_ICON_SIZE = 14.dp
private const val ADD_FILL_OPACITY = 0.4f
private val BORDER_WIDTH = 1.dp
private val HERO_LABEL_TRACKING = 2.2.sp
private val HERO_TEXT_SIZE = 22.sp
private val HERO_TEXT_LINE_HEIGHT = 1.45.em
private val HERO_QUOTE_SIZE = 120.sp
private val HERO_QUOTE_LINE_HEIGHT = 1.em
private val HERO_QUOTE_OFFSET = (-12).dp
private const val HERO_QUOTE_OPACITY = 0.08f
private val ROTATE_ICON_SIZE = 12.dp
private val EMPTY_TITLE_SIZE = 18.sp
private val EMPTY_LINE_HEIGHT = 1.625.em
private val ROW_TEXT_SIZE = 15.5.sp
private val ROW_TEXT_LINE_HEIGHT = 1.375.em
private val SORT_CHIP_GAP = 6.dp
private val SORT_CHIP_VERTICAL = 2.dp
private val LEGEND_ITEM_GAP = 6.dp
private val LEGEND_VERTICAL = 10.dp
private val LEGEND_RADIUS = 12.dp
private const val LEGEND_FILL_OPACITY = 0.6f
private val DOT_SIZE = 6.dp
private val HORIZONTAL_DOT_GAP = 6.dp
private val VERTICAL_DOT_GAP = 2.dp
private val VERTICAL_DOTS_TOP = 4.dp
private const val DOT_IDLE_OPACITY = 0.3f
private const val DOT_COUNT = 3
private const val PULSE_DIMMED_OPACITY = 0.5f
private const val PULSE_DURATION_MILLIS = 1000

@Composable
fun AffirmationsScreen(api: ApiClient) {
    val appContext = LocalContext.current.applicationContext
    val viewModel: AffirmationsViewModel = viewModel(key = "affirmations") {
        AffirmationsViewModel(api, onChanged = { WidgetRefresher.reloadAll(appContext) })
    }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var sheetMode by remember { mutableStateOf<PriorityEditMode?>(null) }

    LaunchedEffect(viewModel) { viewModel.load() }

    Box(Modifier.fillMaxSize().background(DesignTokens.Colors.surface(AREA))) {
        Column(Modifier.fillMaxSize()) {
            ScreenHeader(
                area = AREA,
                title = "자주 읽어줘야 할 것",
                titleFamily = DesignTokens.Typography.Family.Serif,
                trailing = { AddButton(onClick = { sheetMode = PriorityEditMode.Create }) },
            )
            LazyColumn(Modifier.fillMaxSize()) {
                item(key = "hero") {
                    Box(
                        Modifier.padding(
                            start = DesignTokens.Spacing.xl,
                            end = DesignTokens.Spacing.xl,
                            top = DesignTokens.Spacing.sm,
                            bottom = DesignTokens.Spacing.xl,
                        ),
                    ) {
                        HeroSection(
                            hero = state.heroItem,
                            isLoading = state.isLoading,
                            onRotate = viewModel::rotateHero,
                        )
                    }
                }
                if (state.items.isNotEmpty()) {
                    item(key = "list.header") {
                        ListHeader(
                            count = state.items.size,
                            sortOrder = state.sortOrder,
                            onSelect = viewModel::setSortOrder,
                        )
                    }
                    items(state.sortedItems, key = { it.id }) { item ->
                        val shape = RoundedCornerShape(CardPadding.Small.radius)
                        SwipeToDelete(
                            area = AREA,
                            shape = shape,
                            onDelete = { viewModel.delete(item.id) },
                            modifier = Modifier.padding(
                                horizontal = DesignTokens.Spacing.xl,
                                vertical = DesignTokens.Spacing.sm / 2,
                            ),
                        ) {
                            AffirmationRow(item = item, onLongPress = { sheetMode = PriorityEditMode.Edit(item) })
                        }
                    }
                    item(key = "legend") {
                        PriorityLegend(
                            Modifier.padding(
                                start = DesignTokens.Spacing.xl,
                                end = DesignTokens.Spacing.xl,
                                top = DesignTokens.Spacing.xl,
                                bottom = DesignTokens.Spacing.xxl,
                            ),
                        )
                    }
                }
            }
        }
    }

    val mode = sheetMode
    if (mode != null) {
        PriorityEditSheet(
            mode = mode,
            onSave = { text, priority ->
                sheetMode = null
                when (mode) {
                    PriorityEditMode.Create -> viewModel.add(text, priority)
                    is PriorityEditMode.Edit -> viewModel.update(mode.affirmation.id, text, priority)
                }
            },
            onCancel = { sheetMode = null },
            onDelete = {
                if (mode is PriorityEditMode.Edit) {
                    sheetMode = null
                    viewModel.delete(mode.affirmation.id)
                }
            },
        )
    }
}

@Composable
private fun AddButton(onClick: () -> Unit) {
    Box(
        Modifier
            .size(ADD_BUTTON_SIZE)
            .clip(CircleShape)
            .background(DesignTokens.Colors.card(AREA).copy(alpha = ADD_FILL_OPACITY))
            .border(BORDER_WIDTH, DesignTokens.Colors.rule(AREA), CircleShape)
            .clickable(role = Role.Button, onClick = onClick)
            .testTag("affirmations.add.button"),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            Icons.Filled.Add,
            contentDescription = "새 다짐",
            tint = DesignTokens.Colors.ink(AREA),
            modifier = Modifier.size(ADD_ICON_SIZE),
        )
    }
}

@Composable
private fun HeroSection(hero: Affirmation?, isLoading: Boolean, onRotate: () -> Unit) {
    when {
        hero != null -> HeroCard(hero = hero, onRotate = onRotate)
        isLoading -> Card(area = AREA, modifier = Modifier.fillMaxWidth()) {
            Box(
                Modifier.fillMaxWidth().padding(vertical = DesignTokens.Spacing.xl),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = DesignTokens.Colors.accent(AREA))
            }
        }
        else -> EmptyHero()
    }
}

@Composable
private fun HeroCard(hero: Affirmation, onRotate: () -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    val accent = DesignTokens.Colors.accent(AREA)
    Box(Modifier.fillMaxWidth().testTag("affirmations.hero.card")) {
        Card(area = AREA, padding = CardPadding.Large, modifier = Modifier.fillMaxWidth()) {
            Column {
                Row(
                    Modifier.padding(bottom = DesignTokens.Spacing.md),
                    horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    RotationPulseDot()
                    Text(
                        "오늘 회전".uppercase(),
                        style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs, FontWeight.Bold)
                            .copy(letterSpacing = HERO_LABEL_TRACKING),
                        color = accent,
                    )
                }
                Text(
                    hero.text,
                    style = DesignTokens.Typography.style(
                        HERO_TEXT_SIZE,
                        FontWeight.Medium,
                        DesignTokens.Typography.Family.Serif,
                    ).copy(lineHeight = HERO_TEXT_LINE_HEIGHT),
                    color = ink,
                    modifier = Modifier.testTag("affirmations.hero.text"),
                )
                Column(Modifier.padding(top = DesignTokens.Spacing.xl)) {
                    Box(Modifier.fillMaxWidth().height(BORDER_WIDTH).background(DesignTokens.Colors.rule(AREA)))
                    Column(
                        Modifier.padding(top = DesignTokens.Spacing.xl),
                        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
                    ) {
                        Text(
                            buildAnnotatedString {
                                append("우선순위 ")
                                withStyle(SpanStyle(fontWeight = FontWeight.Bold, color = ink)) {
                                    append(hero.priority.displayName)
                                }
                            },
                            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                            color = ink.copy(alpha = 0.55f),
                        )
                        Row(Modifier.fillMaxWidth()) {
                            Spacer(Modifier.weight(1f))
                            Row(
                                Modifier
                                    .clickable(role = Role.Button, onClick = onRotate)
                                    .testTag("affirmations.hero.rotate"),
                                horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                Icon(
                                    Icons.Outlined.Autorenew,
                                    contentDescription = null,
                                    tint = accent,
                                    modifier = Modifier.size(ROTATE_ICON_SIZE),
                                )
                                Text(
                                    "다른 다짐 보기",
                                    style = DesignTokens.Typography.style(
                                        DesignTokens.Typography.captionXs,
                                        FontWeight.SemiBold,
                                    ),
                                    color = accent,
                                )
                            }
                        }
                    }
                }
            }
        }
        Box(
            Modifier
                .matchParentSize()
                .clip(RoundedCornerShape(CardPadding.Large.radius))
                .clearAndSetSemantics {},
        ) {
            Text(
                "\"",
                style = DesignTokens.Typography.style(HERO_QUOTE_SIZE, family = DesignTokens.Typography.Family.Serif)
                    .copy(lineHeight = HERO_QUOTE_LINE_HEIGHT),
                color = accent.copy(alpha = HERO_QUOTE_OPACITY),
                modifier = Modifier.offset(x = HERO_QUOTE_OFFSET, y = HERO_QUOTE_OFFSET),
            )
        }
    }
}

@Composable
private fun EmptyHero() {
    val ink = DesignTokens.Colors.ink(AREA)
    Card(
        area = AREA,
        padding = CardPadding.Large,
        modifier = Modifier.fillMaxWidth().testTag("affirmations.empty.state"),
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
            Text(
                "첫 다짐을 추가해 보세요",
                style = DesignTokens.Typography.style(EMPTY_TITLE_SIZE, FontWeight.Bold, DesignTokens.Typography.Family.Serif),
                color = ink,
            )
            Text(
                "우상단 + 버튼으로 새 다짐을 입력하면 여기에 회전 노출됩니다.",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs)
                    .copy(lineHeight = EMPTY_LINE_HEIGHT),
                color = ink.copy(alpha = 0.55f),
            )
        }
    }
}

@Composable
private fun RotationPulseDot() {
    val transition = rememberInfiniteTransition(label = "affirmations.pulse")
    val opacity by transition.animateFloat(
        initialValue = 1f,
        targetValue = PULSE_DIMMED_OPACITY,
        animationSpec = infiniteRepeatable(tween(PULSE_DURATION_MILLIS), RepeatMode.Reverse),
        label = "affirmations.pulse.opacity",
    )
    Box(
        Modifier
            .size(DOT_SIZE)
            .alpha(opacity)
            .background(DesignTokens.Colors.accent(AREA), CircleShape),
    )
}

@Composable
private fun ListHeader(count: Int, sortOrder: AffirmationSortOrder, onSelect: (AffirmationSortOrder) -> Unit) {
    val ink = DesignTokens.Colors.ink(AREA)
    Row(
        Modifier
            .fillMaxWidth()
            .padding(start = DesignTokens.Spacing.xl, end = DesignTokens.Spacing.xl, bottom = DesignTokens.Spacing.sm)
            .testTag("affirmations.list.header"),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            "전체 ${count}개",
            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.Bold),
            color = ink,
        )
        Spacer(Modifier.weight(1f))
        Row(horizontalArrangement = Arrangement.spacedBy(SORT_CHIP_GAP)) {
            AffirmationSortOrder.entries.forEach { order ->
                val selected = order == sortOrder
                Text(
                    order.label,
                    style = DesignTokens.Typography.style(
                        DesignTokens.Typography.captionXs,
                        if (selected) FontWeight.Bold else FontWeight.Normal,
                    ),
                    color = if (selected) ink else ink.copy(alpha = 0.55f),
                    maxLines = 1,
                    modifier = Modifier
                        .clip(CircleShape)
                        .background(if (selected) DesignTokens.Colors.soft(AREA) else Color.Transparent)
                        .selectable(selected = selected, role = Role.Tab, onClick = { onSelect(order) })
                        .padding(horizontal = DesignTokens.Spacing.sm, vertical = SORT_CHIP_VERTICAL)
                        .testTag("affirmations.sort.${order.wire}"),
                )
            }
        }
    }
}

@Composable
private fun AffirmationRow(item: Affirmation, onLongPress: () -> Unit) {
    Box(
        Modifier
            .fillMaxWidth()
            .background(DesignTokens.Colors.surface(AREA), RoundedCornerShape(CardPadding.Small.radius))
            .pointerInput(item) { detectTapGestures(onLongPress = { onLongPress() }) }
            .semantics {
                onLongClick {
                    onLongPress()
                    true
                }
            }
            .testTag("affirmations.row.${item.id}"),
    ) {
        Card(area = AREA, padding = CardPadding.Small, modifier = Modifier.fillMaxWidth()) {
            Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md), verticalAlignment = Alignment.Top) {
                PriorityDotsVertical(item.priority)
                Text(
                    item.text,
                    style = DesignTokens.Typography.style(ROW_TEXT_SIZE, family = DesignTokens.Typography.Family.Serif)
                        .copy(lineHeight = ROW_TEXT_LINE_HEIGHT),
                    color = DesignTokens.Colors.ink(AREA),
                    modifier = Modifier.weight(1f),
                )
            }
        }
    }
}

@Composable
private fun PriorityLegend(modifier: Modifier = Modifier) {
    val shape = RoundedCornerShape(LEGEND_RADIUS)
    val color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.7f)
    val style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs)
    Row(
        modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.soft(AREA).copy(alpha = LEGEND_FILL_OPACITY))
            .padding(horizontal = DesignTokens.Spacing.md, vertical = LEGEND_VERTICAL)
            .testTag("affirmations.priority.legend"),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        listOf(
            AffirmationPriority.HIGH to "자주 노출",
            AffirmationPriority.NORMAL to "보통",
            AffirmationPriority.LOW to "가끔",
        ).forEach { (priority, label) ->
            Row(
                horizontalArrangement = Arrangement.spacedBy(LEGEND_ITEM_GAP),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                PriorityDotsHorizontal(priority)
                Text(label, style = style, color = color)
            }
        }
    }
}

private fun filledDots(priority: AffirmationPriority): Int = when (priority) {
    AffirmationPriority.HIGH -> 3
    AffirmationPriority.NORMAL -> 2
    AffirmationPriority.LOW -> 1
}

@Composable
private fun PriorityDot(filled: Boolean) {
    Box(
        Modifier
            .size(DOT_SIZE)
            .background(
                DesignTokens.Colors.accent(AREA).copy(alpha = if (filled) 1f else DOT_IDLE_OPACITY),
                CircleShape,
            ),
    )
}

@Composable
private fun PriorityDotsHorizontal(priority: AffirmationPriority) {
    val filled = filledDots(priority)
    Row(horizontalArrangement = Arrangement.spacedBy(HORIZONTAL_DOT_GAP)) {
        repeat(DOT_COUNT) { index -> PriorityDot(filled = index < filled) }
    }
}

@Composable
private fun PriorityDotsVertical(priority: AffirmationPriority) {
    val filled = filledDots(priority)
    Column(
        Modifier.padding(top = VERTICAL_DOTS_TOP),
        verticalArrangement = Arrangement.spacedBy(VERTICAL_DOT_GAP),
    ) {
        repeat(DOT_COUNT) { index -> PriorityDot(filled = index < filled) }
    }
}
