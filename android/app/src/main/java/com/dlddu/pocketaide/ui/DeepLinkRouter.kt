package com.dlddu.pocketaide.ui

import com.dlddu.pocketaide.push.PushDeepLink
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

class DeepLinkRouter {
    private val _pending = MutableStateFlow<PushDeepLink?>(null)
    val pending: StateFlow<PushDeepLink?> = _pending.asStateFlow()

    fun receive(link: PushDeepLink) {
        _pending.value = link
    }

    fun consume(link: PushDeepLink) {
        _pending.compareAndSet(link, null)
    }

    private val _weather = MutableStateFlow(false)
    val weather: StateFlow<Boolean> = _weather.asStateFlow()

    fun openWeather() {
        _weather.value = true
    }

    fun consumeWeather() {
        _weather.value = false
    }

    private val _tab = MutableStateFlow<RootTab?>(null)
    val tab: StateFlow<RootTab?> = _tab.asStateFlow()

    fun openTab(tab: RootTab) {
        _tab.value = tab
    }

    fun consumeTab(tab: RootTab) {
        _tab.compareAndSet(tab, null)
    }
}
