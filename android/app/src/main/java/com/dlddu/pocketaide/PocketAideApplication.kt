package com.dlddu.pocketaide

import android.app.Application
import com.dlddu.pocketaide.push.NotificationChannels
import com.dlddu.pocketaide.weather.WeatherLocationWorker

class PocketAideApplication : Application() {
    lateinit var container: AppContainer
        private set

    override fun onCreate() {
        super.onCreate()
        container = AppContainer(this)
        NotificationChannels.register(this)
        container.pushRegistrar.initialize()
        WeatherLocationWorker.schedule(this)
    }
}
