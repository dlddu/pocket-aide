package com.dlddu.pocketaide.sessions

import java.time.Duration
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Locale

sealed class AgentSessionTurnStatus {
    data object Running : AgentSessionTurnStatus()
    data object Ok : AgentSessionTurnStatus()
    data object Error : AgentSessionTurnStatus()
    data object Timeout : AgentSessionTurnStatus()
    data object Canceled : AgentSessionTurnStatus()
    data class Other(val raw: String) : AgentSessionTurnStatus()

    val label: String
        get() = when (this) {
            Running -> "진행 중"
            Ok -> "완료"
            Error -> "실패"
            Timeout -> "시간 초과"
            Canceled -> "취소됨"
            is Other -> raw
        }

    val isFinished: Boolean get() = this != Running

    companion object {
        fun fromMarker(value: String): AgentSessionTurnStatus = when (value) {
            "ok" -> Ok
            "error" -> Error
            "timeout" -> Timeout
            "canceled" -> Canceled
            else -> Other(value)
        }
    }
}

data class AgentSessionTurn(
    val seq: Int,
    val input: String,
    val hasInput: Boolean = true,
    val output: String = "",
    val status: AgentSessionTurnStatus = AgentSessionTurnStatus.Running,
    val startedAt: Instant? = null,
    val endedAt: Instant? = null,
) {
    val inputPreview: String get() = firstLine(input)
    val outputPreview: String get() = lastLine(output)

    companion object {
        internal fun firstLine(text: String): String = lines(text).firstOrNull { it.isNotBlank() } ?: ""

        internal fun lastLine(text: String): String = lines(text).lastOrNull { it.isNotBlank() } ?: ""

        private fun lines(text: String): List<String> = text.split('\n', '\r').filter { it.isNotEmpty() }
    }
}

data class AgentSessionTranscript(val segments: List<Segment> = emptyList()) {
    sealed class Segment {
        data class Text(val text: String) : Segment()
        data class Turn(val turn: AgentSessionTurn) : Segment()
    }

    val turns: List<AgentSessionTurn> get() = segments.mapNotNull { (it as? Segment.Turn)?.turn }
    val hasTurns: Boolean get() = segments.any { it is Segment.Turn }
    val lastSeq: Int? get() = turns.lastOrNull()?.seq

    internal class Marker(val kind: String, val seq: Int, val values: Map<String, String>) {
        val timestamp: Instant? get() = AgentSessionFormatting.date(values["at"])

        companion object {
            fun parse(header: String): Marker? {
                if (!header.startsWith(LEAD)) return null
                val fields = header.removePrefix(LEAD).split(' ').filter { it.isNotEmpty() }
                val kind = fields.firstOrNull() ?: return null
                val values = mutableMapOf<String, String>()
                fields.drop(1).forEach { field ->
                    val pair = field.split("=", limit = 2)
                    if (pair.size == 2 && pair[0].isNotEmpty() && pair[1].isNotEmpty()) values[pair[0]] = pair[1]
                }
                val seq = values["seq"]?.toIntOrNull() ?: return null
                return Marker(kind, seq, values)
            }
        }
    }

    private class Builder {
        val segments = mutableListOf<Segment>()

        fun appendText(text: String) {
            if (text.isEmpty()) return
            val last = segments.lastOrNull()
            if (last is Segment.Text) {
                segments[segments.lastIndex] = Segment.Text(last.text + text)
            } else {
                segments.add(Segment.Text(text))
            }
        }

        fun appendBody(body: String, seq: Int?) {
            if (seq == null) {
                appendText(body)
                return
            }
            updateTurn(seq) { it.copy(output = it.output + body) }
        }

        fun ensureTurn(seq: Int) {
            if (turnIndex(seq) == null) segments.add(Segment.Turn(AgentSessionTurn(seq = seq, input = "", hasInput = false)))
        }

        fun updateTurn(seq: Int, change: (AgentSessionTurn) -> AgentSessionTurn) {
            val index = turnIndex(seq) ?: return
            val segment = segments[index] as? Segment.Turn ?: return
            segments[index] = Segment.Turn(change(segment.turn))
        }

        private fun turnIndex(seq: Int): Int? {
            val index = segments.indexOfLast { it is Segment.Turn && it.turn.seq == seq }
            return index.takeIf { it >= 0 }
        }

        fun apply(marker: Marker, body: String, openTurn: Int?): Int? = when (marker.kind) {
            "input" -> {
                val prompt = body.removeSuffix("\n")
                segments.add(Segment.Turn(AgentSessionTurn(seq = marker.seq, input = prompt, startedAt = marker.timestamp)))
                null
            }
            "output" -> {
                ensureTurn(marker.seq)
                appendBody(body, marker.seq)
                marker.seq
            }
            "end" -> {
                ensureTurn(marker.seq)
                updateTurn(marker.seq) {
                    it.copy(
                        status = AgentSessionTurnStatus.fromMarker(marker.values["status"].orEmpty()),
                        endedAt = marker.timestamp,
                        output = it.output.trimEnd('\n', '\r'),
                    )
                }
                appendText(body)
                null
            }
            else -> {
                appendBody(body, openTurn)
                openTurn
            }
        }
    }

    companion object {
        const val SENTINEL = '\u001E'
        internal const val LEAD = "session-platform "
        internal const val INPUT_LEAD = "\u001Esession-platform input "

        fun parse(raw: String): AgentSessionTranscript {
            val builder = Builder()
            if (SENTINEL !in raw) {
                builder.appendText(raw)
                return AgentSessionTranscript(builder.segments.toList())
            }
            val parts = raw.split(SENTINEL)
            builder.appendText(parts[0])
            var openTurn: Int? = null
            for (part in parts.drop(1)) {
                val newline = part.indexOf('\n')
                if (newline < 0) continue
                val body = part.substring(newline + 1)
                val marker = Marker.parse(part.substring(0, newline))
                if (marker != null) {
                    openTurn = builder.apply(marker, body, openTurn)
                } else {
                    builder.appendBody(body, openTurn)
                }
            }
            return AgentSessionTranscript(builder.segments.toList())
        }

        fun keepingWholeTurns(tail: String): String {
            val start = tail.indexOf(INPUT_LEAD)
            if (start <= 0) return tail
            return tail.substring(start)
        }
    }
}

object AgentSessionTurnTiming {
    private val clock: DateTimeFormatter = DateTimeFormatter.ofPattern("HH:mm", Locale.KOREA)

    fun label(turn: AgentSessionTurn, zone: ZoneId = ZoneId.systemDefault()): String? {
        val started = turn.startedAt ?: return null
        val parts = mutableListOf(started.atZone(zone).format(clock))
        turn.endedAt?.let { ended ->
            val millis = Duration.between(started, ended).toMillis()
            parts.add(AgentSessionFormatting.duration(Math.round(millis / MILLIS_PER_SECOND)))
        }
        return parts.joinToString(" · ")
    }

    private const val MILLIS_PER_SECOND = 1000.0
}
