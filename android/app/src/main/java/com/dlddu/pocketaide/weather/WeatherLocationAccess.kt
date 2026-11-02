package com.dlddu.pocketaide.weather

import android.Manifest
import android.annotation.SuppressLint
import android.content.Context
import android.content.SharedPreferences
import android.content.pm.PackageManager
import android.location.Address
import android.location.Geocoder
import android.location.Location
import android.location.LocationListener
import android.location.LocationManager
import android.os.Build
import android.os.Bundle
import android.os.CancellationSignal
import android.os.Looper
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import kotlinx.coroutines.withTimeoutOrNull
import java.io.IOException
import java.util.concurrent.Executor
import kotlin.coroutines.resume

class SharedPreferencesWeatherStorage(private val prefs: SharedPreferences) : WeatherStorage {
    constructor(context: Context) : this(context.applicationContext.getSharedPreferences(FILE, Context.MODE_PRIVATE))

    override fun getString(key: String): String? = prefs.getString(key, null)

    override fun putString(key: String, value: String) {
        prefs.edit().putString(key, value).apply()
    }

    override fun remove(key: String) {
        prefs.edit().remove(key).apply()
    }

    private companion object {
        const val FILE = "weather"
    }
}

class WeatherLocationAccess(context: Context, private val store: WeatherLocationStore) {
    private val appContext = context.applicationContext
    private val locationManager = appContext.getSystemService(Context.LOCATION_SERVICE) as LocationManager

    fun hasPermission(): Boolean = PERMISSIONS.any {
        appContext.checkSelfPermission(it) == PackageManager.PERMISSION_GRANTED
    }

    fun hasPrecisePermission(): Boolean =
        appContext.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) == PackageManager.PERMISSION_GRANTED

    // Android 9 이하에는 백그라운드 위치 권한이 따로 없고 위치 권한이 곧 백그라운드 권한이다.
    fun hasBackgroundPermission(): Boolean {
        if (!hasPermission()) return false
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.Q) return true
        return appContext.checkSelfPermission(Manifest.permission.ACCESS_BACKGROUND_LOCATION) ==
            PackageManager.PERMISSION_GRANTED
    }

    suspend fun refresh() {
        if (!hasPermission()) {
            store.clear()
            return
        }
        val location = withTimeoutOrNull(TIMEOUT_MILLIS) { currentLocation() } ?: return
        val place = placeName(location)
        val coordinate = WeatherCoordinate.of(location.latitude, location.longitude, place)
        if (coordinate != store.load()) store.save(coordinate)
    }

    private suspend fun currentLocation(): Location? {
        if (appContext.checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION) != PackageManager.PERMISSION_GRANTED &&
            appContext.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) != PackageManager.PERMISSION_GRANTED
        ) {
            return null
        }
        val provider = provider() ?: return lastKnown()
        return try {
            requestSingle(provider) ?: lastKnown()
        } catch (e: SecurityException) {
            null
        } catch (e: IllegalArgumentException) {
            lastKnown()
        }
    }

    private fun provider(): String? {
        val candidates = buildList {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) add(LocationManager.FUSED_PROVIDER)
            add(LocationManager.NETWORK_PROVIDER)
            add(LocationManager.GPS_PROVIDER)
        }
        val available = locationManager.allProviders
        return candidates.firstOrNull { it in available && locationManager.isProviderEnabled(it) }
    }

    @SuppressLint("MissingPermission")
    private fun lastKnown(): Location? {
        if (appContext.checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION) != PackageManager.PERMISSION_GRANTED &&
            appContext.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) != PackageManager.PERMISSION_GRANTED
        ) {
            return null
        }
        var newest: Location? = null
        for (provider in locationManager.allProviders) {
            val candidate = try {
                locationManager.getLastKnownLocation(provider)
            } catch (e: SecurityException) {
                null
            } catch (e: IllegalArgumentException) {
                null
            }
            if (candidate != null && (newest == null || candidate.time > newest.time)) newest = candidate
        }
        return newest
    }

    @SuppressLint("MissingPermission")
    @Suppress("DEPRECATION")
    private suspend fun requestSingle(provider: String): Location? = suspendCancellableCoroutine { continuation ->
        if (appContext.checkSelfPermission(Manifest.permission.ACCESS_COARSE_LOCATION) != PackageManager.PERMISSION_GRANTED &&
            appContext.checkSelfPermission(Manifest.permission.ACCESS_FINE_LOCATION) != PackageManager.PERMISSION_GRANTED
        ) {
            continuation.resume(null)
            return@suspendCancellableCoroutine
        }
        try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
                val signal = CancellationSignal()
                continuation.invokeOnCancellation { signal.cancel() }
                val executor = Executor { it.run() }
                locationManager.getCurrentLocation(provider, signal, executor) { location ->
                    if (continuation.isActive) continuation.resume(location)
                }
            } else {
                val listener = object : LocationListener {
                    override fun onLocationChanged(location: Location) {
                        if (continuation.isActive) continuation.resume(location)
                    }

                    override fun onStatusChanged(provider: String?, status: Int, extras: Bundle?) = Unit

                    override fun onProviderEnabled(provider: String) = Unit

                    override fun onProviderDisabled(provider: String) {
                        if (continuation.isActive) continuation.resume(null)
                    }
                }
                continuation.invokeOnCancellation { locationManager.removeUpdates(listener) }
                locationManager.requestSingleUpdate(provider, listener, Looper.getMainLooper())
            }
        } catch (e: SecurityException) {
            if (continuation.isActive) continuation.resume(null)
        }
    }

    @Suppress("DEPRECATION")
    private suspend fun placeName(location: Location): String? {
        if (!Geocoder.isPresent()) return null
        return withContext(Dispatchers.IO) {
            try {
                Geocoder(appContext).getFromLocation(location.latitude, location.longitude, 1)
                    ?.firstOrNull()
                    ?.let(::localityName)
            } catch (e: IOException) {
                null
            } catch (e: IllegalArgumentException) {
                null
            }
        }
    }

    private fun localityName(address: Address): String? =
        address.locality ?: address.subAdminArea ?: address.adminArea

    companion object {
        val PERMISSIONS = listOf(Manifest.permission.ACCESS_COARSE_LOCATION, Manifest.permission.ACCESS_FINE_LOCATION)
        private const val TIMEOUT_MILLIS = 20_000L
    }
}
