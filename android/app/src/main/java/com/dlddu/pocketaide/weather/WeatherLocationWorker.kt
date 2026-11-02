package com.dlddu.pocketaide.weather

import android.content.Context
import androidx.work.Constraints
import androidx.work.CoroutineWorker
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.WorkerParameters
import com.dlddu.pocketaide.widget.WidgetRefresher
import kotlinx.coroutines.CancellationException
import java.time.Duration

/**
 * 앱을 열지 않아도 위젯 날씨가 현재 위치를 따라가도록, 주기적으로 위치를 새로 잡아
 * 저장하고 위치가 바뀌었으면 위젯을 다시 그린다.
 *
 * 백그라운드 위치 권한("항상 허용")이 없으면 아무것도 하지 않는다. 그 경우 위치는
 * 지금처럼 앱을 열 때만 갱신된다. 예보 자체는 위젯 주기 갱신(30분)이 저장된 위치로 받아온다.
 */
class WeatherLocationWorker(context: Context, params: WorkerParameters) : CoroutineWorker(context, params) {
    override suspend fun doWork(): Result {
        val store = WeatherLocationStore(SharedPreferencesWeatherStorage(applicationContext))
        val access = WeatherLocationAccess(applicationContext, store)
        if (!access.hasBackgroundPermission()) return Result.success()
        return try {
            val before = store.load()
            access.refresh()
            if (store.load() != before) WidgetRefresher.reloadNow(applicationContext)
            Result.success()
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            Result.retry()
        }
    }

    companion object {
        private const val WORK_NAME = "weather.location.refresh"
        val INTERVAL: Duration = Duration.ofHours(3)

        fun schedule(context: Context) {
            val request = PeriodicWorkRequestBuilder<WeatherLocationWorker>(INTERVAL)
                // 지명(Geocoder)과 새 위치의 예보를 받으려면 네트워크가 필요하다.
                .setConstraints(Constraints.Builder().setRequiredNetworkType(NetworkType.CONNECTED).build())
                .build()
            WorkManager.getInstance(context)
                .enqueueUniquePeriodicWork(WORK_NAME, ExistingPeriodicWorkPolicy.UPDATE, request)
        }
    }
}
