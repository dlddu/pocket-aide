package com.dlddu.pocketaide.chat

import androidx.compose.foundation.interaction.DragInteraction
import androidx.compose.foundation.lazy.LazyListState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.State
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.flow.drop

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
internal fun FollowLatestMessage(
    listState: LazyListState,
    count: Int,
    pendingContent: String?,
    follows: MutableState<Boolean>,
) {
    val slack = with(LocalDensity.current) { FOLLOW_SLACK.roundToPx() }
    val lastIndex by rememberUpdatedState(count - 1)

    LaunchedEffect(listState) {
        listState.interactionSource.interactions.collect { interaction ->
            if (interaction is DragInteraction.Start) follows.value = false
        }
    }
    LaunchedEffect(listState, slack) {
        snapshotFlow { listState.isScrollInProgress }
            .distinctUntilChanged()
            // The first value is the state as it stands, not a scroll that
            // ended. Acting on it would re-follow before the list has been
            // measured, where "at the end" cannot be told from "no items yet",
            // and a restored reading position would be thrown away at once.
            .drop(1)
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

/**
 * The message the reader left at the top of the viewport, and how far the list
 * is scrolled past that message's start.
 *
 * The anchor is the message's id and not its place in the list, because the
 * list changes between visits — a message is deleted, a sync brings history
 * the visit never saw, a new session starts the conversation over. An index
 * would then put the reader in front of a different message than the one they
 * were reading, which is worse than sending them to the end.
 */
internal data class ChatScrollAnchor(val messageId: Long, val offset: Int)

/**
 * Where the reader left the list, kept outside the composition: moving to
 * another tab drops the chat screen (RootScreen swaps tabs in a `when`, so
 * nothing holds a departed tab's state) and takes its LazyListState with it,
 * while the ChatViewModel lives on the activity and stays. `null` means the
 * end — a first visit, a visit that ended at the newest message, and a process
 * death that took the view model with it all land there.
 */
internal class ChatScrollMemory {
    var anchor: ChatScrollAnchor? = null
}

/** -1 when nothing is anchored or the anchored message is gone — read as "the end". */
internal fun anchoredIndex(messages: List<ChatMessage>, anchor: ChatScrollAnchor?): Int {
    val anchored = anchor ?: return -1
    return messages.indexOfFirst { it.id == anchored.messageId }
}

/**
 * Whether the list follows its end. A visit that left a reading position
 * behind starts away from the end, so the reader lands back where they were
 * instead of at the newest message.
 */
@Composable
internal fun rememberFollowsListEnd(memory: ChatScrollMemory): MutableState<Boolean> =
    remember(memory) { mutableStateOf(memory.anchor == null) }

/**
 * Puts the reader back where the last visit left them and keeps that position
 * current while they are away from the end. Following the end leaves nothing
 * to remember — the next visit belongs at the newest message.
 */
@Composable
internal fun KeepReadingPosition(
    listState: LazyListState,
    messages: List<ChatMessage>,
    memory: ChatScrollMemory,
    follows: MutableState<Boolean>,
) {
    val current by rememberUpdatedState(messages)
    val restored = remember(listState) { mutableStateOf(false) }

    LaunchedEffect(listState) {
        val anchor = memory.anchor
        if (anchor != null) {
            val index = anchoredIndex(current, anchor)
            // The message being read is gone, so there is no position to go
            // back to and the end is the one place that still means something.
            if (index < 0) follows.value = true else listState.scrollToItem(index, anchor.offset)
        }
        restored.value = true
    }
    // Waits for the restore: this would otherwise record the start of the list
    // as the reading position before the restore has had a chance to run.
    LaunchedEffect(listState, follows.value, restored.value) {
        if (!restored.value) return@LaunchedEffect
        if (follows.value) {
            memory.anchor = null
            return@LaunchedEffect
        }
        snapshotFlow { listState.firstVisibleItemIndex to listState.firstVisibleItemScrollOffset }
            .collect { (index, offset) ->
                memory.anchor = current.getOrNull(index)?.let { ChatScrollAnchor(it.id, offset) }
            }
    }
}

/**
 * Whether to offer the way back to the newest message. It asks the same
 * question the following does, so the button is on screen exactly while the
 * list has handed the position over to the reader.
 */
@Composable
internal fun rememberShowsListEndButton(listState: LazyListState): State<Boolean> {
    val slack = with(LocalDensity.current) { FOLLOW_SLACK.roundToPx() }
    return remember(listState, slack) { derivedStateOf { !listState.isAtListEnd(slack) } }
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
