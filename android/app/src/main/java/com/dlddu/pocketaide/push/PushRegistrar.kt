package com.dlddu.pocketaide.push

import android.Manifest
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import android.util.Log
import androidx.core.content.ContextCompat
import com.dlddu.pocketaide.BuildConfig
import com.dlddu.pocketaide.api.ApiClient
import com.google.firebase.FirebaseApp
import com.google.firebase.FirebaseOptions
import com.google.firebase.messaging.FirebaseMessaging
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlin.coroutines.resume
import kotlin.coroutines.resumeWithException

class PushRegistrar(private val context: Context) {
    val isConfigured: Boolean
        get() = BuildConfig.FIREBASE_PROJECT_ID.isNotEmpty() &&
            BuildConfig.FIREBASE_APP_ID.isNotEmpty() &&
            BuildConfig.FIREBASE_API_KEY.isNotEmpty()

    fun initialize() {
        if (!isConfigured || FirebaseApp.getApps(context).isNotEmpty()) return
        val options = FirebaseOptions.Builder()
            .setProjectId(BuildConfig.FIREBASE_PROJECT_ID)
            .setApplicationId(BuildConfig.FIREBASE_APP_ID)
            .setApiKey(BuildConfig.FIREBASE_API_KEY)
            .setGcmSenderId(BuildConfig.FIREBASE_SENDER_ID)
            .build()
        FirebaseApp.initializeApp(context, options)
    }

    fun notificationsPermitted(): Boolean =
        Build.VERSION.SDK_INT < Build.VERSION_CODES.TIRAMISU ||
            ContextCompat.checkSelfPermission(context, Manifest.permission.POST_NOTIFICATIONS) ==
            PackageManager.PERMISSION_GRANTED

    suspend fun register(api: ApiClient) {
        if (!isConfigured) return
        initialize()
        try {
            FirebaseMessaging.getInstance().isAutoInitEnabled = true
            api.registerDeviceToken(currentToken())
        } catch (e: Exception) {
            Log.w(TAG, "push registration failed", e)
        }
    }

    fun unregister() {
        if (!isConfigured || FirebaseApp.getApps(context).isEmpty()) return
        FirebaseMessaging.getInstance().isAutoInitEnabled = false
        FirebaseMessaging.getInstance().deleteToken()
    }

    private suspend fun currentToken(): String = suspendCancellableCoroutine { cont ->
        FirebaseMessaging.getInstance().token
            .addOnSuccessListener { cont.resume(it) }
            .addOnFailureListener { cont.resumeWithException(it) }
    }

    private companion object {
        const val TAG = "PushRegistrar"
    }
}
