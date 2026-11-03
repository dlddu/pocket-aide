package com.dlddu.pocketaide.todos

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.outlined.Search
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
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.em
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.PlainTextField
import com.dlddu.pocketaide.designsystem.components.ScreenHeader
import com.dlddu.pocketaide.designsystem.components.SwipeToDelete

private val AREA_STRIP_HEIGHT = 3.dp
private val TITLE_SIZE = 28.sp
private val LABEL_DOT_SIZE = 8.dp
private val ADD_BUTTON_SIZE = 40.dp
private val ADD_ICON_SIZE = 18.dp
private val SEARCH_ICON_SIZE = 14.dp
private val BORDER_WIDTH = 1.dp
private val ROW_PADDING = 14.dp
private val ROW_CHECK_TOP = 2.dp
private val ROW_TEXT_SPACING = 6.dp
private val ROW_TITLE_LINE_HEIGHT = 1.375.em
private val EMPTY_LINE_HEIGHT = 1.625.em
private val SECTION_RULE_HEIGHT = 1.dp
private const val DONE_TITLE_OPACITY = 0.45f

@Composable
fun TodoListScreen(area: TodoArea, api: ApiClient) {
    val viewModel: TodoListViewModel = viewModel(key = "todos.${area.wire}") { TodoListViewModel(area, api) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    var sheetMode by remember { mutableStateOf<TodoSheetMode?>(null) }
    var isDoneCollapsed by rememberSaveable { mutableStateOf(false) }
    val tone = area.designArea
    val idPrefix = "todos.${area.wire}"

    LaunchedEffect(viewModel) { viewModel.load() }

    Box(Modifier.fillMaxSize().background(DesignTokens.Colors.surface(tone))) {
        Column(Modifier.fillMaxSize()) {
            Box(Modifier.fillMaxWidth().height(AREA_STRIP_HEIGHT).background(DesignTokens.Colors.accent(tone)))
            ScreenHeader(
                area = tone,
                title = area.screenTitle,
                titleSize = TITLE_SIZE,
                labelShowsDot = true,
                labelDotShape = area.headerDotShape,
                labelDotSize = LABEL_DOT_SIZE,
                labelSpacing = DesignTokens.Spacing.xs,
                subtitle = area.summary(state.openCount, state.doneCount, state.deadlineCount),
                subtitleTag = "$idPrefix.summary",
                subtitleMonospaced = area == TodoArea.WORK,
                bottomPadding = DesignTokens.Spacing.md,
                trailing = {
                    AddButton(area = area, tag = "$idPrefix.add.button", onClick = { sheetMode = TodoSheetMode.Create })
                },
            )
            SearchField(area = area, query = state.query, onQueryChange = viewModel::setQuery, tag = "$idPrefix.search.field")
            if (area == TodoArea.PERSONAL && state.items.isNotEmpty()) {
                TodoDueFilterChips(
                    area = area,
                    total = state.items.size,
                    selection = state.dueFilter,
                    onSelect = viewModel::setDueFilter,
                )
            }
            TodoList(
                area = area,
                state = state,
                isDoneCollapsed = isDoneCollapsed,
                onToggleCollapse = { isDoneCollapsed = !isDoneCollapsed },
                onRefresh = viewModel::refresh,
                onToggle = viewModel::toggleDone,
                onEdit = { sheetMode = TodoSheetMode.Edit(it) },
                onDelete = { viewModel.delete(it.id) },
            )
        }
    }

    val mode = sheetMode
    if (mode != null) {
        TodoEditSheet(
            area = area,
            mode = mode,
            onSave = { draft ->
                sheetMode = null
                when (mode) {
                    TodoSheetMode.Create -> viewModel.add(draft)
                    is TodoSheetMode.Edit -> viewModel.update(mode.item.id, draft)
                }
            },
            onCancel = { sheetMode = null },
            onDelete = {
                if (mode is TodoSheetMode.Edit) {
                    sheetMode = null
                    viewModel.delete(mode.item.id)
                }
            },
        )
    }
}

@Composable
private fun AddButton(area: TodoArea, tag: String, onClick: () -> Unit) {
    val shape: Shape = if (area.checkIsRound) CircleShape else RoundedCornerShape(area.rowRadius)
    Box(
        Modifier
            .size(ADD_BUTTON_SIZE)
            .clip(shape)
            .background(DesignTokens.Colors.accent(area.designArea))
            .clickable(role = Role.Button, onClick = onClick)
            .testTag(tag),
        contentAlignment = Alignment.Center,
    ) {
        Icon(Icons.Filled.Add, contentDescription = "추가", tint = Color.White, modifier = Modifier.size(ADD_ICON_SIZE))
    }
}

@Composable
private fun SearchField(area: TodoArea, query: String, onQueryChange: (String) -> Unit, tag: String) {
    val tone = area.designArea
    val ink = DesignTokens.Colors.ink(tone)
    val shape = RoundedCornerShape(area.searchRadius)
    Row(
        Modifier
            .padding(start = DesignTokens.Spacing.xl, end = DesignTokens.Spacing.xl, bottom = DesignTokens.Spacing.md)
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(tone))
            .border(BORDER_WIDTH, DesignTokens.Colors.rule(tone), shape)
            .padding(horizontal = DesignTokens.Spacing.md, vertical = DesignTokens.Spacing.sm),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Icon(
            Icons.Outlined.Search,
            contentDescription = null,
            tint = ink.copy(alpha = 0.45f),
            modifier = Modifier.size(SEARCH_ICON_SIZE),
        )
        PlainTextField(
            value = query,
            onValueChange = onQueryChange,
            placeholder = area.searchPrompt,
            textStyle = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
            textColor = ink,
            placeholderColor = ink.copy(alpha = 0.45f),
            singleLine = true,
            keyboardOptions = KeyboardOptions(autoCorrectEnabled = false),
            modifier = Modifier.weight(1f).testTag(tag),
        )
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun TodoList(
    area: TodoArea,
    state: TodoListUiState,
    isDoneCollapsed: Boolean,
    onToggleCollapse: () -> Unit,
    onRefresh: () -> Unit,
    onToggle: (TodoItem) -> Unit,
    onEdit: (TodoItem) -> Unit,
    onDelete: (TodoItem) -> Unit,
) {
    val tone = area.designArea
    val idPrefix = "todos.${area.wire}"
    val visible = state.visibleItems()
    val openItems = visible.filter { !it.isDone }
    val doneItems = visible.filter { it.isDone }
    val ink = DesignTokens.Colors.ink(tone)

    PullToRefreshBox(
        isRefreshing = state.isRefreshing,
        onRefresh = onRefresh,
        modifier = Modifier.fillMaxSize(),
    ) {
        LazyColumn(Modifier.fillMaxSize()) {
            state.errorMessage?.let { message ->
                item(key = "error") {
                    Text(
                        message,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                        color = DesignTokens.Colors.destructive(tone),
                        modifier = Modifier
                            .padding(horizontal = DesignTokens.Spacing.xl, vertical = DesignTokens.Spacing.sm)
                            .testTag("$idPrefix.error"),
                    )
                }
            }
            if (state.items.isEmpty()) {
                item(key = "empty") {
                    Box(
                        Modifier
                            .fillMaxWidth()
                            .padding(
                                start = DesignTokens.Spacing.xl,
                                end = DesignTokens.Spacing.xl,
                                top = area.listTopInset,
                                bottom = DesignTokens.Spacing.xxl,
                            )
                            .padding(vertical = DesignTokens.Spacing.xl),
                        contentAlignment = if (state.isLoading) Alignment.Center else Alignment.CenterStart,
                    ) {
                        if (state.isLoading) {
                            CircularProgressIndicator(color = DesignTokens.Colors.accent(tone))
                        } else {
                            Text(
                                "아직 할 일이 없습니다. 우상단 + 버튼으로 추가하세요.",
                                style = DesignTokens.Typography.style(DesignTokens.Typography.body)
                                    .copy(lineHeight = EMPTY_LINE_HEIGHT),
                                color = ink.copy(alpha = 0.55f),
                                modifier = Modifier.testTag("$idPrefix.empty.state"),
                            )
                        }
                    }
                }
            } else {
                section(
                    area = area,
                    key = "open",
                    title = if (area == TodoArea.WORK) "OPEN" else "할 일",
                    items = openItems,
                    titleOpacity = 1f,
                    isFirstSection = true,
                    isLastSection = doneItems.isEmpty(),
                    collapsible = false,
                    isCollapsed = false,
                    onToggleCollapse = onToggleCollapse,
                    onToggle = onToggle,
                    onEdit = onEdit,
                    onDelete = onDelete,
                )
                section(
                    area = area,
                    key = "done",
                    title = if (area == TodoArea.WORK) "DONE" else "완료",
                    items = doneItems,
                    titleOpacity = DONE_TITLE_OPACITY,
                    isFirstSection = openItems.isEmpty(),
                    isLastSection = true,
                    collapsible = area == TodoArea.PERSONAL,
                    isCollapsed = isDoneCollapsed,
                    onToggleCollapse = onToggleCollapse,
                    onToggle = onToggle,
                    onEdit = onEdit,
                    onDelete = onDelete,
                )
            }
        }
    }
}

private fun LazyListScope.section(
    area: TodoArea,
    key: String,
    title: String,
    items: List<TodoItem>,
    titleOpacity: Float,
    isFirstSection: Boolean,
    isLastSection: Boolean,
    collapsible: Boolean,
    isCollapsed: Boolean,
    onToggleCollapse: () -> Unit,
    onToggle: (TodoItem) -> Unit,
    onEdit: (TodoItem) -> Unit,
    onDelete: (TodoItem) -> Unit,
) {
    if (items.isEmpty()) return
    item(key = "$key.header") {
        SectionHeader(
            area = area,
            title = title,
            count = items.size,
            titleOpacity = titleOpacity,
            isFirstSection = isFirstSection,
            collapsible = collapsible,
            isCollapsed = isCollapsed,
            onToggleCollapse = onToggleCollapse,
        )
    }
    val shown = if (collapsible && isCollapsed) emptyList() else items
    itemsIndexed(shown, key = { _, item -> item.id }) { index, item ->
        val isLast = isLastSection && index == shown.lastIndex
        val shape = RoundedCornerShape(area.rowRadius)
        SwipeToDelete(
            area = area.designArea,
            shape = shape,
            onDelete = { onDelete(item) },
            modifier = Modifier.padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = DesignTokens.Spacing.cardGap / 2,
                bottom = if (isLast) DesignTokens.Spacing.xxl else DesignTokens.Spacing.cardGap / 2,
            ),
        ) {
            TodoRow(area = area, item = item, shape = shape, onToggle = { onToggle(item) }, onClick = { onEdit(item) })
        }
    }
}

@Composable
private fun SectionHeader(
    area: TodoArea,
    title: String,
    count: Int,
    titleOpacity: Float,
    isFirstSection: Boolean,
    collapsible: Boolean,
    isCollapsed: Boolean,
    onToggleCollapse: () -> Unit,
) {
    val tone = area.designArea
    val ink = DesignTokens.Colors.ink(tone)
    val label = "$title · $count"
    Row(
        Modifier
            .fillMaxWidth()
            .padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = if (isFirstSection) area.listTopInset + DesignTokens.Spacing.xs else DesignTokens.Spacing.md,
                bottom = area.sectionHeaderBottom,
            ),
        horizontalArrangement = Arrangement.spacedBy(area.sectionHeadSpacing),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        if (area.sectionBarWidth > 0.dp) {
            Box(
                Modifier
                    .width(area.sectionBarWidth)
                    .height(DesignTokens.Spacing.md)
                    .background(if (titleOpacity == 1f) DesignTokens.Colors.accent(tone) else ink.copy(alpha = titleOpacity)),
            )
        }
        Text(
            if (area.sectionTitleUppercase) label.uppercase() else label,
            style = DesignTokens.Typography.style(area.sectionTitleSize, FontWeight.Bold)
                .copy(letterSpacing = area.sectionTitleTracking),
            color = ink.copy(alpha = titleOpacity),
        )
        if (collapsible) {
            Text(
                if (isCollapsed) "펼치기" else "접기",
                style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                color = ink.copy(alpha = 0.45f),
                textAlign = TextAlign.End,
                modifier = Modifier
                    .weight(1f)
                    .clickable(role = Role.Button, onClick = onToggleCollapse)
                    .testTag("todos.${area.wire}.done.collapse"),
            )
        }
        if (area.showsSectionRule) {
            Box(Modifier.weight(1f).height(SECTION_RULE_HEIGHT).background(DesignTokens.Colors.rule(tone)))
        }
    }
}

@Composable
private fun TodoRow(area: TodoArea, item: TodoItem, shape: Shape, onToggle: () -> Unit, onClick: () -> Unit) {
    val tone = area.designArea
    val ink = DesignTokens.Colors.ink(tone)
    Box(Modifier.fillMaxWidth().background(DesignTokens.Colors.surface(tone), shape)) {
        Row(
            Modifier
                .fillMaxWidth()
                .alpha(if (item.isDone) 0.6f else 1f)
                .clip(shape)
                .background(DesignTokens.Colors.card(tone).copy(alpha = if (item.isDone) 0.5f else 1f))
                .border(BORDER_WIDTH, DesignTokens.Colors.rule(tone), shape)
                .clickable(onClick = onClick)
                .padding(ROW_PADDING),
            horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
            verticalAlignment = Alignment.Top,
        ) {
            TodoCheckBox(
                area = area,
                isDone = item.isDone,
                modifier = Modifier
                    .padding(top = ROW_CHECK_TOP)
                    .clickable(role = Role.Checkbox, onClick = onToggle)
                    .testTag("todos.${area.wire}.row.${item.id}.toggle"),
            )
            Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(ROW_TEXT_SPACING)) {
                Column {
                    if (area == TodoArea.PERSONAL && item.priority == TodoPriority.HIGH) {
                        TodoPriorityTag(area = area, item = item)
                    }
                    Text(
                        item.title,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, area.rowTitleWeight).copy(
                            lineHeight = ROW_TITLE_LINE_HEIGHT,
                            textDecoration = if (item.isDone) TextDecoration.LineThrough else null,
                        ),
                        color = ink,
                        modifier = Modifier.testTag("todos.${area.wire}.row.${item.id}"),
                    )
                }
                TodoRowMeta(area = area, item = item)
            }
        }
    }
}
