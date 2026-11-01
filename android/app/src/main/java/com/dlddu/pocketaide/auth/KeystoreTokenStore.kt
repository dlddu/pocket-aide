package com.dlddu.pocketaide.auth

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import kotlinx.serialization.json.Json
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

class KeystoreTokenStore(context: Context) : TokenStore {
    private val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    private val lock = Any()

    override fun load(): TokenBundle? = synchronized(lock) {
        val stored = prefs.getString(KEY_BUNDLE, null) ?: return null
        runCatching {
            val raw = Base64.decode(stored, Base64.NO_WRAP)
            val iv = raw.copyOfRange(0, IV_LENGTH)
            val cipher = Cipher.getInstance(TRANSFORMATION)
            cipher.init(Cipher.DECRYPT_MODE, key(), GCMParameterSpec(TAG_BITS, iv))
            val plain = cipher.doFinal(raw, IV_LENGTH, raw.size - IV_LENGTH)
            json.decodeFromString(TokenBundle.serializer(), plain.decodeToString())
        }.getOrElse {
            prefs.edit().remove(KEY_BUNDLE).apply()
            null
        }
    }

    override fun save(bundle: TokenBundle) = synchronized(lock) {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, key())
        val sealed = cipher.iv + cipher.doFinal(json.encodeToString(TokenBundle.serializer(), bundle).encodeToByteArray())
        prefs.edit().putString(KEY_BUNDLE, Base64.encodeToString(sealed, Base64.NO_WRAP)).commit()
        Unit
    }

    override fun clear() = synchronized(lock) {
        prefs.edit().remove(KEY_BUNDLE).commit()
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
                .setKeySize(256)
                .build(),
        )
        return generator.generateKey()
    }

    private companion object {
        const val PREFS_NAME = "pocketaide.tokens"
        const val KEY_BUNDLE = "bundle"
        const val KEY_ALIAS = "pocketaide.tokens.key"
        const val ANDROID_KEYSTORE = "AndroidKeyStore"
        const val TRANSFORMATION = "AES/GCM/NoPadding"
        const val IV_LENGTH = 12
        const val TAG_BITS = 128
        val json = Json { ignoreUnknownKeys = true }
    }
}
