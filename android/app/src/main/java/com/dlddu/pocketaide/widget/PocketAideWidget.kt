package com.dlddu.pocketaide.widget

import android.content.Context
import android.content.Intent
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.produceState
import androidx.compose.runtime.remember
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.core.net.toUri
import androidx.glance.GlanceId
import androidx.glance.GlanceModifier
import androidx.glance.LocalContext
import androidx.glance.action.Action
import androidx.glance.action.clickable
import androidx.glance.appwidget.GlanceAppWidget
import androidx.glance.appwidget.GlanceAppWidgetReceiver
import androidx.glance.appwidget.action.actionStartActivity
import androidx.glance.appwidget.cornerRadius
import androidx.glance.appwidget.provideContent
import androidx.glance.background
import androidx.glance.currentState
import androidx.glance.layout.Alignment
import androidx.glance.layout.Box
import androidx.glance.layout.Column
import androidx.glance.layout.Row
import androidx.glance.layout.Spacer
import androidx.glance.layout.fillMaxHeight
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.height
import androidx.glance.layout.padding
import androidx.glance.layout.size
import androidx.glance.layout.width
import androidx.glance.text.FontFamily
import androidx.glance.text.FontWeight
import androidx.glance.text.Text
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import com.dlddu.pocketaide.MainActivity
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabelDefaults
import com.dlddu.pocketaide.weather.Weather
import com.dlddu.pocketaide.weather.WeatherDeepLink

class PocketAideWidget : GlanceAppWidget() {
    override suspend fun provideGlance(context: Context, id: GlanceId) {
        val loader = PocketAideWidgetLoader(context)
        val initial = loader.load()
        provideContent {
            val revision = currentState(WidgetRefresher.REVISION) ?: 0L
            val firstRevision = remember { revision }
            val entry by produceState(initial, revision) {
                if (revision != firstRevision) value = loader.load()
            }
            WidgetContent(entry)
        }
    }
}

class PocketAideWidgetReceiver : GlanceAppWidgetReceiver() {
    override val glanceAppWidget: GlanceAppWidget = PocketAideWidget()
}

private val RULE_WIDTH = 1.dp
private val LINE_GAP = 2.dp
private const val SECONDARY_OPACITY = 0.75f
private const val NOTICE_OPACITY = 0.6f
private const val MORE_OPACITY = 0.55f
private const val PLACEHOLDER_OPACITY = 0.55f
private const val AFFIRMATION_MAX_LINES = 5

@Composable
private fun WidgetContent(entry: PocketAideWidgetEntry) {
    val context = LocalContext.current
    Column(
        GlanceModifier
            .fillMaxSize()
            .background(WidgetColors.surface)
            .padding(DesignTokens.Spacing.lg)
            .clickable(actionStartActivity(Intent(context, MainActivity::class.java))),
    ) {
        Row(GlanceModifier.fillMaxWidth().padding(bottom = DesignTokens.Spacing.md)) {
            WeatherSection(entry.weather, GlanceModifier.defaultWeight())
            Spacer(GlanceModifier.width(DesignTokens.Spacing.md))
            Box(GlanceModifier.width(RULE_WIDTH).fillMaxHeight().background(WidgetColors.rule)) {}
            Spacer(GlanceModifier.width(DesignTokens.Spacing.md))
            CalendarSection(entry.calendar, entry, GlanceModifier.defaultWeight())
        }
        HorizontalRule()
        AffirmationSection(entry.affirmation)
        HorizontalRule()
        NotificationSection(entry.notifications)
        HorizontalRule()
        ScratchpadSection(entry.scratchpad)
    }
}

@Composable
private fun HorizontalRule() {
    Box(GlanceModifier.fillMaxWidth().height(RULE_WIDTH).background(WidgetColors.rule)) {}
}

@Composable
private fun WeatherSection(state: WidgetWeatherState, modifier: GlanceModifier) {
    val area = Area.Personal
    Column(modifier.clickable(open(LocalContext.current, WeatherDeepLink.URL))) {
        SectionLabel(area, WidgetCopy.weatherLabel(state))
        Spacer(GlanceModifier.height(DesignTokens.Spacing.xs))
        if (state is WidgetWeatherState.Loaded) {
            Row(verticalAlignment = Alignment.Bottom) {
                Text(
                    Weather.temperatureLabel(state.summary),
                    style = style(DesignTokens.Typography.h1, WidgetColors.ink(area), FontWeight.Bold),
                    maxLines = 1,
                )
                Spacer(GlanceModifier.width(DesignTokens.Spacing.xs))
                Text(
                    state.summary.condition,
                    style = style(DesignTokens.Typography.captionXs, WidgetColors.ink(area, SECONDARY_OPACITY)),
                    maxLines = 1,
                )
            }
            Spacer(GlanceModifier.height(LINE_GAP))
            Text(
                Weather.detailLabel(state.summary),
                style = style(DesignTokens.Typography.caption2xs, WidgetColors.ink(area, SECONDARY_OPACITY)),
                maxLines = 1,
            )
        } else {
            Notice(area, WidgetCopy.weatherNotice(state).orEmpty(), maxLines = 2)
        }
    }
}

@Composable
private fun CalendarSection(state: WidgetCalendarState, entry: PocketAideWidgetEntry, modifier: GlanceModifier) {
    val area = Area.Work
    val notice = WidgetCopy.calendarNotice(state)
    Column(modifier.clickable(open(LocalContext.current, WidgetDeepLink.CALENDAR))) {
        SectionLabel(area, WidgetCopy.CALENDAR_LABEL)
        Spacer(GlanceModifier.height(DesignTokens.Spacing.xs))
        if (state is WidgetCalendarState.Loaded && notice == null) {
            val summary = state.summary
            val first = summary.events.first()
            Text(
                first.title,
                style = style(DesignTokens.Typography.bodySm, WidgetColors.ink(area), FontWeight.Bold),
                maxLines = 2,
            )
            Spacer(GlanceModifier.height(LINE_GAP))
            Text(
                UpcomingEvents.rangeLabel(first, entry.date),
                style = style(DesignTokens.Typography.captionXs, WidgetColors.ink(area, SECONDARY_OPACITY)),
                maxLines = 1,
            )
            summary.events.drop(1).forEach { event ->
                Spacer(GlanceModifier.height(LINE_GAP))
                Text(
                    "${UpcomingEvents.startLabel(event, entry.date)} ${event.title}",
                    style = style(DesignTokens.Typography.caption2xs, WidgetColors.ink(area, SECONDARY_OPACITY)),
                    maxLines = 2,
                )
            }
            WidgetCopy.moreLabel(summary.moreCount)?.let { MoreLine(area, it) }
        } else {
            Notice(area, notice.orEmpty(), maxLines = 2)
        }
    }
}

@Composable
private fun AffirmationSection(state: WidgetAffirmationState) {
    val area = Area.Affirmations
    val alpha = if (state is WidgetAffirmationState.Loaded) 1f else PLACEHOLDER_OPACITY
    Column(
        GlanceModifier
            .fillMaxWidth()
            .padding(vertical = DesignTokens.Spacing.md)
            .clickable(open(LocalContext.current, WidgetDeepLink.AFFIRMATIONS)),
    ) {
        SectionLabel(area, WidgetCopy.AFFIRMATION_LABEL, showsDot = true)
        Spacer(GlanceModifier.height(DesignTokens.Spacing.sm))
        Text(
            WidgetCopy.affirmationMessage(state),
            style = style(DesignTokens.Typography.bodyLg, WidgetColors.ink(area, alpha), family = FontFamily.Serif),
            maxLines = AFFIRMATION_MAX_LINES,
        )
    }
}

@Composable
private fun NotificationSection(state: WidgetNotificationState) {
    val area = Area.Scratchpad
    val summary = (state as? WidgetNotificationState.Loaded)?.summary
    val latest = summary?.latest
    val modifier = GlanceModifier
        .fillMaxWidth()
        .padding(vertical = DesignTokens.Spacing.md)
        .clickable(open(LocalContext.current, WidgetDeepLink.NOTIFICATIONS))
    if (summary != null && latest != null) {
        Column(modifier) {
            SectionLabel(area, WidgetCopy.NOTIFICATION_LABEL)
            Spacer(GlanceModifier.height(DesignTokens.Spacing.xs))
            Text(
                latest.title,
                style = style(DesignTokens.Typography.captionSm, WidgetColors.ink(area), FontWeight.Bold),
                maxLines = 1,
            )
            if (latest.body.isNotEmpty()) {
                Spacer(GlanceModifier.height(LINE_GAP))
                Text(
                    latest.body,
                    style = style(DesignTokens.Typography.caption2xs, WidgetColors.ink(area, SECONDARY_OPACITY)),
                    maxLines = 1,
                )
            }
            WidgetCopy.moreLabel(summary.moreCount)?.let { MoreLine(area, it) }
        }
    } else {
        Row(modifier, verticalAlignment = Alignment.CenterVertically) {
            SectionLabel(area, WidgetCopy.NOTIFICATION_LABEL)
            Spacer(GlanceModifier.width(DesignTokens.Spacing.sm))
            Spacer(GlanceModifier.defaultWeight())
            Notice(area, WidgetCopy.notificationNotice(state).orEmpty(), maxLines = 1)
        }
    }
}

@Composable
private fun ScratchpadSection(state: WidgetScratchpadState) {
    val area = Area.Scratchpad
    Row(
        GlanceModifier
            .fillMaxWidth()
            .padding(top = DesignTokens.Spacing.md)
            .clickable(open(LocalContext.current, WidgetDeepLink.SCRATCHPAD)),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        SectionLabel(area, WidgetCopy.SCRATCHPAD_LABEL)
        Spacer(GlanceModifier.width(DesignTokens.Spacing.sm))
        Spacer(GlanceModifier.defaultWeight())
        if (WidgetCopy.scratchpadHasItems(state)) {
            Text(
                WidgetCopy.scratchpadText(state),
                style = style(DesignTokens.Typography.captionSm, WidgetColors.ink(area), FontWeight.Medium),
                maxLines = 1,
            )
        } else {
            Notice(area, WidgetCopy.scratchpadText(state), maxLines = 1)
        }
    }
}

@Composable
private fun SectionLabel(area: Area, text: String, showsDot: Boolean = false) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        if (showsDot) {
            Box(
                GlanceModifier
                    .size(AreaLabelDefaults.dotSize)
                    .cornerRadius(AreaLabelDefaults.dotSize / 2)
                    .background(WidgetColors.accent(area)),
            ) {}
            Spacer(GlanceModifier.width(AreaLabelDefaults.dotGap))
        }
        Text(
            text.uppercase(),
            style = style(DesignTokens.Typography.captionXs, WidgetColors.accent(area), FontWeight.Bold),
            maxLines = 1,
        )
    }
}

@Composable
private fun Notice(area: Area, text: String, maxLines: Int) {
    Text(
        text,
        style = style(DesignTokens.Typography.captionXs, WidgetColors.ink(area, NOTICE_OPACITY)),
        maxLines = maxLines,
    )
}

@Composable
private fun MoreLine(area: Area, text: String) {
    Spacer(GlanceModifier.height(LINE_GAP))
    Text(
        text,
        style = style(DesignTokens.Typography.caption2xs, WidgetColors.ink(area, MORE_OPACITY)),
        maxLines = 1,
    )
}

private fun style(
    size: TextUnit,
    color: ColorProvider,
    weight: FontWeight = FontWeight.Normal,
    family: FontFamily? = null,
): TextStyle = TextStyle(color = color, fontSize = size, fontWeight = weight, fontFamily = family)

private fun open(context: Context, url: String): Action =
    actionStartActivity(Intent(Intent.ACTION_VIEW, url.toUri(), context, MainActivity::class.java))
