package com.dlddu.pocketaide.prmonitor

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

@Serializable
data class GitHubCredential(val token: String, val login: String)

interface GitHubCredentialStore {
    fun load(): GitHubCredential?
    fun save(credential: GitHubCredential)
    fun clear()
}

class InMemoryGitHubCredentialStore(initial: GitHubCredential? = null) : GitHubCredentialStore {
    @Volatile
    private var credential: GitHubCredential? = initial

    override fun load(): GitHubCredential? = credential

    override fun save(credential: GitHubCredential) {
        this.credential = credential
    }

    override fun clear() {
        credential = null
    }
}

class KeystoreGitHubCredentialStore(context: Context) : GitHubCredentialStore {
    private val prefs = context.applicationContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    private val lock = Any()

    override fun load(): GitHubCredential? = synchronized(lock) {
        val stored = prefs.getString(KEY_CREDENTIAL, null) ?: return null
        runCatching {
            val raw = Base64.decode(stored, Base64.NO_WRAP)
            val iv = raw.copyOfRange(0, IV_LENGTH)
            val cipher = Cipher.getInstance(TRANSFORMATION)
            cipher.init(Cipher.DECRYPT_MODE, key(), GCMParameterSpec(TAG_BITS, iv))
            val plain = cipher.doFinal(raw, IV_LENGTH, raw.size - IV_LENGTH)
            json.decodeFromString(GitHubCredential.serializer(), plain.decodeToString())
        }.getOrElse {
            prefs.edit().remove(KEY_CREDENTIAL).apply()
            null
        }
    }

    override fun save(credential: GitHubCredential) = synchronized(lock) {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, key())
        val plain = json.encodeToString(GitHubCredential.serializer(), credential).encodeToByteArray()
        val sealed = cipher.iv + cipher.doFinal(plain)
        prefs.edit().putString(KEY_CREDENTIAL, Base64.encodeToString(sealed, Base64.NO_WRAP)).commit()
        Unit
    }

    override fun clear() = synchronized(lock) {
        prefs.edit().remove(KEY_CREDENTIAL).commit()
        Unit
    }

    private fun key(): SecretKey {
        val keyStore = KeyStore.getInstance(ANDROID_KEYSTORE).apply { load(null) }
        (keyStore.getEntry(KEY_ALIAS, null) as? KeyStore.SecretKeyEntry)?.let { return it.secretKey }
        val generator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, ANDROID_KEYSTORE)
        generator.init(
            KeyGenParameterSpec.Builder(KEY_ALIAS, KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT)
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setKeySize(KEY_SIZE)
                .build(),
        )
        return generator.generateKey()
    }

    private companion object {
        const val PREFS_NAME = "pocketaide.github"
        const val KEY_CREDENTIAL = "credential"
        const val KEY_ALIAS = "pocketaide.github.key"
        const val ANDROID_KEYSTORE = "AndroidKeyStore"
        const val TRANSFORMATION = "AES/GCM/NoPadding"
        const val IV_LENGTH = 12
        const val TAG_BITS = 128
        const val KEY_SIZE = 256
        val json = Json { ignoreUnknownKeys = true }
    }
}

interface OpenPullRequestFilterStore {
    fun load(): OpenPullRequestFilter?
    fun save(filter: OpenPullRequestFilter)
}

class InMemoryOpenPullRequestFilterStore(private var filter: OpenPullRequestFilter? = null) : OpenPullRequestFilterStore {
    override fun load(): OpenPullRequestFilter? = filter

    override fun save(filter: OpenPullRequestFilter) {
        this.filter = filter
    }
}

class SharedPreferencesOpenPullRequestFilterStore(context: Context) : OpenPullRequestFilterStore {
    private val prefs = context.applicationContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)

    override fun load(): OpenPullRequestFilter? = OpenPullRequestFilter.fromWire(prefs.getString(FILTER_KEY, null))

    override fun save(filter: OpenPullRequestFilter) {
        prefs.edit().putString(FILTER_KEY, filter.wire).apply()
    }

    private companion object {
        const val PREFS_NAME = "pocketaide.prmonitor"
        const val FILTER_KEY = "openprs.filter"
    }
}
