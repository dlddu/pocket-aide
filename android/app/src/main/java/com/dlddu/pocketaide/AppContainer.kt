package com.dlddu.pocketaide

import android.content.Context
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.auth.KeystoreTokenStore
import com.dlddu.pocketaide.auth.OidcClient
import com.dlddu.pocketaide.auth.OidcRedirectBus
import com.dlddu.pocketaide.auth.TokenStore
import com.dlddu.pocketaide.push.PushRegistrar
import com.dlddu.pocketaide.ui.DeepLinkRouter
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.Dispatchers
import okhttp3.OkHttpClient
import java.util.concurrent.TimeUnit

class AppContainer(context: Context) {
    val appScope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    val tokenStore: TokenStore = KeystoreTokenStore(context)
    val redirectBus = OidcRedirectBus()
    val deepLinks = DeepLinkRouter()
    val pushRegistrar = PushRegistrar(context)

    private val http = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(60, TimeUnit.SECONDS)
        .build()

    val backendConfigured: Boolean = BuildConfig.BACKEND_BASE_URL.isNotBlank()

    val api: ApiClient = ApiClient(
        baseUrl = BuildConfig.BACKEND_BASE_URL.ifBlank { UNCONFIGURED_BASE_URL },
        tokenStore = tokenStore,
        http = http,
    )

    val oidc: OidcClient = OidcClient(api, tokenStore, redirectBus, http).also { api.tokenRefresher = it }

    private companion object {
        const val UNCONFIGURED_BASE_URL = "http://backend.invalid"
    }
}
