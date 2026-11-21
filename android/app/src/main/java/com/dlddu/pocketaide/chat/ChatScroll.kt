package com.dlddu.pocketaide.chat

import androidx.compose.foundation.interaction.DragInteraction
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.flow.distinctUntilChanged

private val FOLLOW_SLACK = 48.dp
private const val MAX_FOLLOW_STEPS = 64

/**
 * Keeps the newest text of a streaming reply in view.
 *
 * The list's end is pinned to the bottom of the viewport rather than scrolled
 * to the last item, which shows only a long reply's first lines while the text
 * being appended grows past the fold. Pinning is instant — a reply changes
 * every poll and an animation restarts on each change instead of arriving.
 * A drag hands the position over until a gesture settles back at the end.
 */
@Composable
internal fun FollowLatestMessage(listState: LazyListState, count: Int, pendingContent: String?) {
    val slack = with(LocalDensity.current) { FOLLOW_SLACK.roundToPx() }
    val follows = remember { mutableStateOf(true) }
    val lastIndex by rememberUpdatedState(count - 1)

    LaunchedEffect(listState) {
        listState.interactionSource.interactions.collect { interaction ->
            if (interaction is DragInteraction.Start) follows.value = false
        }
    }
    LaunchedEffect(listState, slack) {
        snapshotFlow { listState.isScrollInProgress }
            .distinctUntilChanged()
            .collect { scrolling ->
                if (!scrolling && !follows.value) follows.value = listState.isAtListEnd(slack)
            }
    }
    LaunchedEffect(listState, count, pendingContent, follows.value) {
        if (follows.value) listState.pinToListEnd(lastIndex)
    }
    LaunchedEffect(listState) {
        snapshotFlow { listState.layoutInfo.viewportSize }
            .distinctUntilChanged()
            .collect { if (follows.value) listState.pinToListEnd(lastIndex) }
    }
}

internal suspend fun LazyListState.pinToListEnd(lastIndex: Int) {
    if (lastIndex < 0) return
    val viewport = layoutInfo.viewportSize.height.toFloat()
    if (viewport <= 0f) return
    if (layoutInfo.visibleItemsInfo.lastOrNull()?.index != lastIndex) scrollToItem(lastIndex)
    scroll {
        for (step in 0 until MAX_FOLLOW_STEPS) {
            if (scrollBy(viewport) < viewport) break
        }
    }
}

internal fun LazyListState.isAtListEnd(slackPx: Int): Boolean {
    if (!canScrollForward) return true
    val info = layoutInfo
    val last = info.visibleItemsInfo.lastOrNull() ?: return true
    return isWithinListEnd(
        lastVisibleIndex = last.index,
        lastVisibleEndPx = last.offset + last.size,
        totalCount = info.totalItemsCount,
        viewportEndPx = info.viewportEndOffset,
        slackPx = slackPx,
    )
}

internal fun isWithinListEnd(
    lastVisibleIndex: Int,
    lastVisibleEndPx: Int,
    totalCount: Int,
    viewportEndPx: Int,
    slackPx: Int,
): Boolean = lastVisibleIndex == totalCount - 1 && lastVisibleEndPx - viewportEndPx <= slackPx
