package com.dlddu.pocketaide.widget

import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.weather.Weather
import java.time.Instant
import kotlin.math.ceil
import kotlin.math.max

/**
 * 위젯 상단(날씨 | 다음 일정) 사이 세로 구분선의 높이를 정한다.
 *
 * iOS 는 상단 HStack 에 `.fixedSize(horizontal: false, vertical: true)` 를 걸어
 * 구분선이 두 섹션 중 더 긴 쪽 내용 높이에 맞춰진다. Glance(RemoteViews)에는
 * 내용 높이(intrinsic size)로 맞추는 수단이 없어서, 구분선에 `fillMaxHeight()` 를 주면
 * 위젯에 남은 높이 전체로 늘어나 아래 섹션들을 밀어낸다.
 *
 * 그래서 [PocketAideWidget] 의 WeatherSection / CalendarSection 이 그리는 텍스트 블록을
 * 그대로 본떠 각 섹션의 내용 높이를 추정하고, 더 긴 쪽 높이를 구분선에 직접 준다.
 * 섹션 구성을 바꾸면 [weatherBlocks] / [calendarBlocks] 도 같이 바꿔야 한다.
 */
object WidgetHeaderLayout {
    const val NOTICE_MAX_LINES = 2
    const val EVENT_TITLE_MAX_LINES = 2
    const val EVENT_LINE_MAX_LINES = 2

    sealed interface Block {
        data class Text(val text: String, val sizeSp: Float, val maxLines: Int = 1) : Block
        data class Gap(val dp: Float) : Block
    }

    /** 위젯 전체 너비에서 상단 한 섹션이 쓸 수 있는 너비(dp). */
    fun sectionWidth(widgetWidthDp: Float): Float {
        val used = DesignTokens.Spacing.lg.value * 2 + DesignTokens.Spacing.md.value * 2 + RULE_WIDTH.value
        return max(0f, (widgetWidthDp - used) / 2)
    }

    fun ruleHeight(entry: PocketAideWidgetEntry, sectionWidthDp: Float, fontScale: Float): Float = max(
        height(weatherBlocks(entry.weather), sectionWidthDp, fontScale),
        height(calendarBlocks(entry.calendar, entry.date), sectionWidthDp, fontScale),
    )

    fun weatherBlocks(state: WidgetWeatherState): List<Block> = buildList {
        add(label(WidgetCopy.weatherLabel(state)))
        add(Block.Gap(DesignTokens.Spacing.xs.value))
        if (state is WidgetWeatherState.Loaded) {
            // 기온(h1)과 상태(captionXs)는 한 줄에 나란히 놓이므로 더 큰 기온 줄 높이만 센다.
            add(Block.Text(Weather.temperatureLabel(state.summary), DesignTokens.Typography.h1.value))
            add(Block.Gap(LINE_GAP.value))
            add(Block.Text(Weather.detailLabel(state.summary), DesignTokens.Typography.caption2xs.value))
        } else {
            add(notice(WidgetCopy.weatherNotice(state).orEmpty()))
        }
    }

    fun calendarBlocks(state: WidgetCalendarState, now: Instant): List<Block> = buildList {
        add(label(WidgetCopy.CALENDAR_LABEL))
        add(Block.Gap(DesignTokens.Spacing.xs.value))
        val notice = WidgetCopy.calendarNotice(state)
        if (state is WidgetCalendarState.Loaded && notice == null) {
            val summary = state.summary
            val first = summary.events.first()
            add(Block.Text(first.title, DesignTokens.Typography.bodySm.value, EVENT_TITLE_MAX_LINES))
            add(Block.Gap(LINE_GAP.value))
            add(Block.Text(UpcomingEvents.rangeLabel(first, now), DesignTokens.Typography.captionXs.value))
            summary.events.drop(1).forEach { event ->
                add(Block.Gap(LINE_GAP.value))
                add(
                    Block.Text(
                        "${UpcomingEvents.startLabel(event, now)} ${event.title}",
                        DesignTokens.Typography.caption2xs.value,
                        EVENT_LINE_MAX_LINES,
                    ),
                )
            }
            WidgetCopy.moreLabel(summary.moreCount)?.let {
                add(Block.Gap(LINE_GAP.value))
                add(Block.Text(it, DesignTokens.Typography.caption2xs.value))
            }
        } else {
            add(notice(notice.orEmpty()))
        }
    }

    /** 블록들을 위에서부터 쌓았을 때의 높이(dp, 올림). */
    fun height(blocks: List<Block>, widthDp: Float, fontScale: Float): Float = ceil(
        blocks.sumOf { block ->
            when (block) {
                is Block.Gap -> block.dp.toDouble()
                is Block.Text -> {
                    val lines = lineCount(block.text, block.sizeSp, fontScale, widthDp, block.maxLines)
                    (lines * lineHeight(block.text, block.sizeSp, fontScale)).toDouble()
                }
            }
        },
    ).toFloat()

    /**
     * TextView 한 줄 높이 추정(dp). 한글이 섞인 줄은 대체 글꼴(CJK)의 위·아래 여백 때문에
     * 라틴 글꼴만 쓰는 줄보다 높다.
     */
    fun lineHeight(text: String, sizeSp: Float, fontScale: Float): Float {
        val ratio = if (text.any(::isWide)) CJK_LINE_HEIGHT else LATIN_LINE_HEIGHT
        return sizeSp * fontScale * ratio
    }

    /** 글자 폭을 근사해 [maxLines] 안에서 몇 줄로 접힐지 추정한다. */
    fun lineCount(text: String, sizeSp: Float, fontScale: Float, widthDp: Float, maxLines: Int): Int {
        if (maxLines <= 1 || widthDp <= 0f) return 1
        val em = sizeSp * fontScale
        val textWidth = text.sumOf { ch ->
            val ratio = when {
                isWide(ch) -> WIDE_CHAR
                ch.isWhitespace() -> SPACE_CHAR
                else -> NARROW_CHAR
            }
            (ratio * em).toDouble()
        }
        return ceil(textWidth / widthDp).toInt().coerceIn(1, maxLines)
    }

    private fun label(text: String) = Block.Text(text.uppercase(), DesignTokens.Typography.captionXs.value)

    private fun notice(text: String) = Block.Text(text, DesignTokens.Typography.captionXs.value, NOTICE_MAX_LINES)

    private fun isWide(ch: Char): Boolean {
        val block = Character.UnicodeBlock.of(ch)
        return block == Character.UnicodeBlock.HANGUL_SYLLABLES ||
            block == Character.UnicodeBlock.HANGUL_JAMO ||
            block == Character.UnicodeBlock.HANGUL_COMPATIBILITY_JAMO ||
            block == Character.UnicodeBlock.CJK_UNIFIED_IDEOGRAPHS ||
            block == Character.UnicodeBlock.HIRAGANA ||
            block == Character.UnicodeBlock.KATAKANA
    }

    private const val LATIN_LINE_HEIGHT = 1.17f
    private const val CJK_LINE_HEIGHT = 1.45f
    private const val WIDE_CHAR = 0.95f
    private const val NARROW_CHAR = 0.56f
    private const val SPACE_CHAR = 0.28f
}
