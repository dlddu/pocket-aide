package com.dlddu.pocketaide.push

import android.content.Context

/** 푸시 기기 이름. `load()` 가 null 이면 아직 묻지 않았고, "" 이면 건너뛰었다. */
interface DeviceNameStore {
    fun load(): String?
    fun save(name: String)
}

class PrefsDeviceNameStore(context: Context) : DeviceNameStore {
    private val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)

    override fun load(): String? = prefs.getString(KEY_NAME, null)

    override fun save(name: String) {
        prefs.edit().putString(KEY_NAME, name).apply()
    }

    private companion object {
        const val PREFS = "pocketaide.device"
        const val KEY_NAME = "name"
    }
}

class InMemoryDeviceNameStore(private var name: String? = null) : DeviceNameStore {
    override fun load(): String? = name

    override fun save(name: String) {
        this.name = name
    }
}

object DeviceName {
    const val MAX_LENGTH = 40

    fun normalize(raw: String): String = raw.trim().take(MAX_LENGTH).trim()

    /** 서버로 보낼 값 — 건너뛴 경우("")와 묻기 전(null)은 보내지 않는다. */
    fun forRegistration(stored: String?): String? = stored?.takeIf { it.isNotEmpty() }
}
