package com.dlddu.pocketaide.weather

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.time.Instant

sealed interface WeatherContent {
    data object Loading : WeatherContent
    data object NeedsLocation : WeatherContent
    data object Error : WeatherContent
    data class Loaded(val snapshot: WeatherForecastSnapshot) : WeatherContent
}

data class WeatherUiState(
    val content: WeatherContent = WeatherContent.Loading,
    val placeName: String? = null,
    val isRefreshing: Boolean = false,
) {
    val title: String get() = placeName?.takeIf { it.isNotEmpty() } ?: "날씨"
}

class WeatherViewModel(
    private val store: WeatherLocationStore,
    private val cache: WeatherForecastCache,
    private val client: WeatherClient,
    private val clock: () -> Instant = Instant::now,
) : ViewModel() {
    private val _state = MutableStateFlow(WeatherUiState())
    val state: StateFlow<WeatherUiState> = _state.asStateFlow()

    private var shownLocation: WeatherCoordinate? = null

    fun load(useCache: Boolean) {
        viewModelScope.launch { performLoad(useCache) }
    }

    fun refresh() {
        viewModelScope.launch {
            _state.update { it.copy(isRefreshing = true) }
            try {
                performLoad(useCache = false)
            } finally {
                _state.update { it.copy(isRefreshing = false) }
            }
        }
    }

    fun locationDidRefresh() {
        viewModelScope.launch { performLocationDidRefresh() }
    }

    internal suspend fun performLoad(useCache: Boolean) {
        val location = store.load()
        if (location == null) {
            shownLocation = null
            _state.update { it.copy(content = WeatherContent.NeedsLocation, placeName = null) }
            return
        }
        shownLocation = location
        _state.update { it.copy(placeName = location.placeName) }
        if (useCache) {
            cache.load(location, clock())?.let { cached ->
                _state.update { it.copy(content = WeatherContent.Loaded(cached)) }
            }
        }
        val content = try {
            WeatherContent.Loaded(client.fetchForecastToCompletion(location))
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            WeatherContent.Error
        }
        _state.update { it.copy(content = content) }
    }

    internal suspend fun performLocationDidRefresh() {
        if (store.load() == shownLocation) return
        performLoad(useCache = true)
    }
}
