package com.dlddu.pocketaide.api

data class SseEvent(val event: String, val data: String)

class SseParser {
    private var event = ""
    private val data = StringBuilder()
    private var hasData = false

    fun feed(line: String): SseEvent? {
        if (line.isEmpty()) return flush()
        if (line.startsWith(":")) return null
        val colon = line.indexOf(':')
        val field = if (colon < 0) line else line.substring(0, colon)
        val value = if (colon < 0) "" else line.substring(colon + 1).removePrefix(" ")
        when (field) {
            "event" -> event = value
            "data" -> {
                if (hasData) data.append('\n')
                data.append(value)
                hasData = true
            }
        }
        return null
    }

    private fun flush(): SseEvent? {
        val out = if (hasData) SseEvent(event.ifEmpty { "message" }, data.toString()) else null
        event = ""
        data.clear()
        hasData = false
        return out
    }
}
