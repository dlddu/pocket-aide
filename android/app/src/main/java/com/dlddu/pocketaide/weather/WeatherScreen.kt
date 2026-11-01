package com.dlddu.pocketaide.weather

import android.Manifest
import android.content.Intent
import android.net.Uri
import android.provider.Settings
import androidx.activity.compose.BackHandler
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.testTag
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.semantics.semantics
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import com.dlddu.pocketaide.designsystem.Area
import com.dlddu.pocketaide.designsystem.DesignTokens
import com.dlddu.pocketaide.designsystem.components.AreaLabel
import com.dlddu.pocketaide.designsystem.components.Card
import kotlinx.coroutines.launch
import okhttp3.OkHttpClient

private val AREA = Area.Personal
private val TITLE_SIZE = 24.sp
private val CURRENT_TEMPERATURE_SIZE = 48.sp
private val HOUR_MIN_WIDTH = 52.dp
private val DAY_LABEL_WIDTH = 76.dp
private val DAY_RANGE_WIDTH = 84.dp
private val BORDER_WIDTH = 1.dp
private const val SECONDARY_OPACITY = 0.75f
private const val TABULAR_NUMS = "tnum"

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WeatherScreen(onClose: () -> Unit) {
    val context = LocalContext.current
    val storage = remember { SharedPreferencesWeatherStorage(context) }
    val viewModel: WeatherViewModel = viewModel(key = "weather") {
        val cache = WeatherForecastCache(storage)
        WeatherViewModel(WeatherLocationStore(storage), cache, WeatherClient(OkHttpClient(), cache))
    }
    val access = remember { WeatherLocationAccess(context, WeatherLocationStore(storage)) }
    val state by viewModel.state.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()
    var locating by remember { mutableStateOf(false) }

    val refreshLocation: () -> Unit = {
        if (!locating) {
            locating = true
            scope.launch {
                try {
                    access.refresh()
                } finally {
                    locating = false
                }
                viewModel.locationDidRefresh()
            }
        }
    }

    val permissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) {
        refreshLocation()
    }

    LaunchedEffect(viewModel) {
        viewModel.load(useCache = true)
        if (!access.hasPermission()) permissionLauncher.launch(Manifest.permission.ACCESS_COARSE_LOCATION)
    }
    LifecycleEventEffect(Lifecycle.Event.ON_RESUME) {
        if (access.hasPermission()) refreshLocation()
    }
    BackHandler(onBack = onClose)

    Box(
        Modifier
            .fillMaxSize()
            .background(DesignTokens.Colors.surface(AREA))
            .testTag("weather.screen"),
    ) {
        Column(Modifier.fillMaxSize().safeDrawingPadding()) {
            Header(title = state.title, onClose = onClose)
            PullToRefreshBox(
                isRefreshing = state.isRefreshing,
                onRefresh = viewModel::refresh,
                modifier = Modifier.fillMaxSize(),
            ) {
                Column(
                    Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState())
                        .padding(
                            start = DesignTokens.Spacing.xl,
                            end = DesignTokens.Spacing.xl,
                            bottom = DesignTokens.Spacing.xxl,
                        ),
                ) {
                    when (val content = state.content) {
                        WeatherContent.Loading -> LoadingIndicator()
                        WeatherContent.NeedsLocation ->
                            if (locating) {
                                LoadingIndicator()
                            } else {
                                Notice(
                                    text = "앱에서 위치 접근을 허용해 주세요.",
                                    action = "설정 열기",
                                    tag = "weather.settings",
                                    onClick = {
                                        context.startActivity(
                                            Intent(
                                                Settings.ACTION_APPLICATION_DETAILS_SETTINGS,
                                                Uri.fromParts("package", context.packageName, null),
                                            ),
                                        )
                                    },
                                )
                            }
                        WeatherContent.Error -> Notice(
                            text = "잠시 후 다시 시도할게요.",
                            action = "다시 시도",
                            tag = "weather.retry",
                            onClick = { viewModel.load(useCache = false) },
                        )
                        is WeatherContent.Loaded -> Forecast(content.snapshot)
                    }
                }
            }
        }
    }
}

@Composable
private fun Header(title: String, onClose: () -> Unit) {
    Row(
        Modifier
            .fillMaxWidth()
            .padding(
                start = DesignTokens.Spacing.xl,
                end = DesignTokens.Spacing.xl,
                top = DesignTokens.Spacing.xl,
                bottom = DesignTokens.Spacing.md,
            ),
        verticalAlignment = Alignment.Bottom,
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
            AreaLabel(area = AREA, text = "날씨")
            Text(
                title,
                style = DesignTokens.Typography.style(TITLE_SIZE, FontWeight.Bold),
                color = DesignTokens.Colors.ink(AREA),
                modifier = Modifier.testTag("weather.place"),
            )
        }
        Spacer(Modifier.weight(1f))
        Text(
            "닫기",
            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
            color = DesignTokens.Colors.accent(AREA),
            modifier = Modifier
                .clickable(role = Role.Button, onClick = onClose)
                .testTag("weather.close"),
        )
    }
}

@Composable
private fun LoadingIndicator() {
    Box(
        Modifier.fillMaxWidth().padding(top = DesignTokens.Spacing.xxl),
        contentAlignment = Alignment.Center,
    ) {
        CircularProgressIndicator(color = DesignTokens.Colors.accent(AREA))
    }
}

@Composable
private fun Notice(text: String, action: String, tag: String, onClick: () -> Unit) {
    Column(
        Modifier.padding(top = DesignTokens.Spacing.lg),
        verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
    ) {
        Text(
            text,
            style = DesignTokens.Typography.style(DesignTokens.Typography.body),
            color = DesignTokens.Colors.ink(AREA).copy(alpha = SECONDARY_OPACITY),
            modifier = Modifier.testTag("weather.notice"),
        )
        Text(
            action,
            style = DesignTokens.Typography.style(DesignTokens.Typography.body, FontWeight.SemiBold),
            color = DesignTokens.Colors.accent(AREA),
            modifier = Modifier.clickable(role = Role.Button, onClick = onClick).testTag(tag),
        )
    }
}

@Composable
private fun Forecast(snapshot: WeatherForecastSnapshot) {
    Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.lg)) {
        Current(snapshot.forecast)
        SectionTitle("시간별 예보")
        Hourly(snapshot.forecast.hours)
        SectionTitle("주간 예보")
        Weekly(snapshot.forecast.days)
        Text(
            Weather.updatedLabel(snapshot.fetchedAt),
            style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
            color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.55f),
            modifier = Modifier.testTag("weather.updated"),
        )
    }
}

@Composable
private fun Current(forecast: WeatherForecast) {
    val ink = DesignTokens.Colors.ink(AREA)
    Card(area = AREA, modifier = Modifier.fillMaxWidth()) {
        Column(Modifier.fillMaxWidth(), verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs)) {
            Row(horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm)) {
                Text(
                    Weather.temperatureLabel(forecast.summary),
                    style = DesignTokens.Typography.style(CURRENT_TEMPERATURE_SIZE, FontWeight.Bold)
                        .copy(fontFeatureSettings = TABULAR_NUMS),
                    color = ink,
                    modifier = Modifier.alignByBaseline().testTag("weather.current.temperature"),
                )
                Text(
                    forecast.summary.condition,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.titleMd, FontWeight.SemiBold),
                    color = ink.copy(alpha = SECONDARY_OPACITY),
                    modifier = Modifier.alignByBaseline().testTag("weather.current.condition"),
                )
            }
            Text(
                Weather.apparentLabel(forecast),
                style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
                color = ink.copy(alpha = SECONDARY_OPACITY),
                modifier = Modifier.testTag("weather.current.apparent"),
            )
            Text(
                Weather.detailLabel(forecast.summary),
                style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
                color = ink.copy(alpha = SECONDARY_OPACITY),
                modifier = Modifier.testTag("weather.current.detail"),
            )
        }
    }
}

@Composable
private fun SectionTitle(text: String) {
    Text(
        text,
        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm, FontWeight.Bold),
        color = DesignTokens.Colors.ink(AREA).copy(alpha = 0.7f),
    )
}

@Composable
private fun Hourly(hours: List<WeatherHour>) {
    val ink = DesignTokens.Colors.ink(AREA)
    val accent = DesignTokens.Colors.accent(AREA)
    val shape = RoundedCornerShape(DesignTokens.Radius.card)
    val numeric = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs).copy(fontFeatureSettings = TABULAR_NUMS)
    Row(
        Modifier
            .fillMaxWidth()
            .clip(shape)
            .background(DesignTokens.Colors.card(AREA))
            .border(BORDER_WIDTH, DesignTokens.Colors.rule(AREA), shape)
            .horizontalScroll(rememberScrollState())
            .padding(DesignTokens.Spacing.md),
        horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md),
        verticalAlignment = Alignment.Top,
    ) {
        hours.forEachIndexed { index, hour ->
            Column(
                Modifier
                    .widthIn(min = HOUR_MIN_WIDTH)
                    .semantics(mergeDescendants = true) {}
                    .testTag("weather.hour.$index"),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.xs),
            ) {
                Text(
                    hour.time,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.captionXs),
                    color = ink.copy(alpha = 0.6f),
                )
                Text(
                    "${hour.temperature}°",
                    style = DesignTokens.Typography.style(DesignTokens.Typography.bodyLg, FontWeight.Bold)
                        .copy(fontFeatureSettings = TABULAR_NUMS),
                    color = ink,
                )
                Text(
                    hour.condition,
                    style = DesignTokens.Typography.style(DesignTokens.Typography.caption2xs),
                    color = ink.copy(alpha = SECONDARY_OPACITY),
                )
                Text(Weather.chanceLabel(hour.precipitationChance), style = numeric, color = accent)
            }
        }
    }
}

@Composable
private fun Weekly(days: List<WeatherDay>) {
    val ink = DesignTokens.Colors.ink(AREA)
    val accent = DesignTokens.Colors.accent(AREA)
    Card(area = AREA, modifier = Modifier.fillMaxWidth()) {
        Column(verticalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.md)) {
            days.forEachIndexed { index, day ->
                Row(
                    Modifier
                        .fillMaxWidth()
                        .semantics(mergeDescendants = true) {}
                        .testTag("weather.day.$index"),
                    horizontalArrangement = Arrangement.spacedBy(DesignTokens.Spacing.sm),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Text(
                        day.label,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm, FontWeight.SemiBold),
                        color = ink,
                        modifier = Modifier.width(DAY_LABEL_WIDTH),
                    )
                    Text(
                        day.condition,
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm),
                        color = ink.copy(alpha = SECONDARY_OPACITY),
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier.weight(1f),
                    )
                    Text(
                        Weather.chanceLabel(day.precipitationChance),
                        style = DesignTokens.Typography.style(DesignTokens.Typography.captionSm)
                            .copy(fontFeatureSettings = TABULAR_NUMS),
                        color = accent,
                    )
                    Text(
                        "${day.high}° / ${day.low}°",
                        style = DesignTokens.Typography.style(DesignTokens.Typography.bodySm, FontWeight.SemiBold)
                            .copy(fontFeatureSettings = TABULAR_NUMS),
                        color = ink,
                        textAlign = TextAlign.End,
                        modifier = Modifier.width(DAY_RANGE_WIDTH),
                    )
                }
            }
        }
    }
}
