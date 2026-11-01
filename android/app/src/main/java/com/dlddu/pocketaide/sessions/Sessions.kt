package com.dlddu.pocketaide.sessions

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.KSerializer
import kotlinx.serialization.Serializable
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.encoding.Decoder
import kotlinx.serialization.encoding.Encoder
import java.time.Duration
import java.time.Instant
import java.time.OffsetDateTime
import java.time.format.DateTimeParseException
import java.util.Locale
import kotlin.math.roundToLong

@Serializable(with = AgentSessionStateSerializer::class)
enum class AgentSessionState(val wire: String, val label: String) {
    ACTIVE("active", "실행 중"),
    IDLE("idle", "대기"),
    SNAPSHOT("snapshot", "아카이브됨");

    companion object {
        fun fromWire(raw: String): AgentSessionState = entries.firstOrNull { it.wire == raw } ?: IDLE
    }
}

object AgentSessionStateSerializer : KSerializer<AgentSessionState> {
    override val descriptor: SerialDescriptor = PrimitiveSerialDescriptor("AgentSessionState", PrimitiveKind.STRING)
    override fun serialize(encoder: Encoder, value: AgentSessionState) = encoder.encodeString(value.wire)
    override fun deserialize(decoder: Decoder): AgentSessionState = AgentSessionState.fromWire(decoder.decodeString())
}

@Serializable
data class AgentSessionCheckpoint(
    val sizeBytes: Long? = null,
    val createdAt: String? = null,
)

@Serializable
data class AgentSession(
    val id: String,
    val name: String,
    val workloadType: String = DEFAULT_WORKLOAD_TYPE,
    val state: AgentSessionState,
    val model: String? = null,
    val createdAt: String = "",
    val lastAccess: String = "",
    val checkpoint: AgentSessionCheckpoint? = null,
) {
    val modelLabel: String get() = model?.takeIf { it.isNotEmpty() } ?: "platform-default"

    fun kindLabel(type: AgentSessionWorkloadType?): String {
        val name = type?.displayName ?: workloadType
        val model = model?.takeIf { it.isNotEmpty() } ?: return name
        return "$name · $model"
    }

    companion object {
        const val DEFAULT_WORKLOAD_TYPE = "claude-code"
    }
}

@Serializable
data class AgentSessionReadResult(
    val session: AgentSession? = null,
    val payload: String = "",
    val nextOffset: Long = 0,
)

@Serializable
data class AgentSessionWriteResult(val session: AgentSession? = null)

data class AgentSessionOutput(
    val text: String = "",
    val nextOffset: Long = 0,
    val isTrimmed: Boolean = false,
) {
    enum class Outcome { GREW, UNCHANGED, NEEDS_FULL_REREAD }

    data class Applied(val output: AgentSessionOutput, val outcome: Outcome)

    fun merge(result: AgentSessionReadResult, requestedOffset: Long): Applied {
        if (result.nextOffset < requestedOffset) return Applied(this, Outcome.NEEDS_FULL_REREAD)
        if (requestedOffset == 0L) {
            val next = AgentSessionOutput(nextOffset = result.nextOffset).appending(result.payload)
            return Applied(next, if (next.text == text) Outcome.UNCHANGED else Outcome.GREW)
        }
        if (result.payload.isEmpty()) return Applied(copy(nextOffset = result.nextOffset), Outcome.UNCHANGED)
        return Applied(copy(nextOffset = result.nextOffset).appending(result.payload), Outcome.GREW)
    }

    private fun appending(payload: String): AgentSessionOutput {
        val combined = text + payload
        if (combined.length <= DISPLAY_LIMIT) return copy(text = combined)
        return copy(text = AgentSessionTranscript.keepingWholeTurns(combined.takeLast(DISPLAY_LIMIT)), isTrimmed = true)
    }

    companion object {
        const val DISPLAY_LIMIT = 120_000
    }
}

enum class AgentSessionResponsePhase(val label: String?) {
    NONE(null),
    QUEUED("전송됨 · 응답 대기 중"),
    STREAMING("응답 진행 중"),
    QUIET("새 출력이 멈췄어요 · 응답이 끝났을 수 있어요"),
    DONE("응답 완료");

    fun next(outputGrew: Boolean, millisSinceLastOutput: Long): AgentSessionResponsePhase = when (this) {
        NONE -> NONE
        QUEUED -> if (outputGrew) STREAMING else QUEUED
        STREAMING, QUIET -> when {
            outputGrew -> STREAMING
            millisSinceLastOutput >= QUIET_AFTER_MILLIS -> QUIET
            else -> this
        }
        DONE -> if (outputGrew) STREAMING else DONE
    }

    companion object {
        const val QUIET_AFTER_MILLIS = 10_000L

        fun fromTurns(transcript: AgentSessionTranscript, awaitingThrough: Int?): AgentSessionResponsePhase? {
            val last = transcript.turns.lastOrNull() ?: return null
            if (!last.status.isFinished) return STREAMING
            if (awaitingThrough == null) return NONE
            return if (last.seq >= awaitingThrough) DONE else QUEUED
        }
    }
}

object AgentSessionPrompt {
    const val MAX_BYTES = 1 shl 20

    fun exceedsLimit(prompt: String): Boolean = prompt.toByteArray(Charsets.UTF_8).size > MAX_BYTES
}

enum class AgentSessionError(val message: String) {
    UNAUTHORIZED("로그인이 만료됐어요. 다시 로그인해 주세요."),
    NOT_FOUND("세션을 찾을 수 없어요. 목록을 새로고침해 주세요."),
    BUSY("세션이 다른 작업 중이에요. 잠시 후 다시 시도해 주세요."),
    TOO_LARGE("프롬프트가 너무 커요. 1 MiB 이하로 줄여 주세요."),
    SNAPSHOTTED("스냅샷 상태인 세션이에요. 복원한 뒤 이어서 쓸 수 있어요."),
    QUEUE_FULL("앞선 프롬프트가 아직 처리 중이에요. 끝난 뒤 다시 보내 주세요."),
    DISABLED("이 환경에서는 아카이브가 꺼져 있어요. 세션은 그대로 유지되고 데이터는 사라지지 않아요."),
    QUOTA_FULL("세션 출력 한도가 찼어요. 기존 출력은 볼 수 있지만 새 프롬프트는 보낼 수 없어요."),
    UNREACHABLE("세션 플랫폼에 연결할 수 없어요. 네트워크를 확인하고 다시 시도해 주세요."),
    UNKNOWN("세션 플랫폼에서 오류가 났어요. 잠시 후 다시 시도해 주세요.");

    companion object {
        private val byStatus = mapOf(
            401 to UNAUTHORIZED,
            404 to NOT_FOUND,
            409 to BUSY,
            413 to TOO_LARGE,
            422 to SNAPSHOTTED,
            429 to QUEUE_FULL,
            503 to UNREACHABLE,
            507 to QUOTA_FULL,
            502 to UNREACHABLE,
            504 to UNREACHABLE,
        )

        fun fromStatus(status: Int): AgentSessionError = byStatus[status] ?: UNKNOWN

        fun from(error: Throwable): AgentSessionError = when (error) {
            is ApiException.BadStatus -> fromStatus(error.status)
            is ApiException.Decoding -> UNKNOWN
            else -> UNREACHABLE
        }

        fun forSnapshot(error: Throwable): AgentSessionError =
            if (error is ApiException.BadStatus && error.status == 503) DISABLED else from(error)
    }
}

object AgentSessionFormatting {
    private const val KILOBYTE = 1000.0
    private val FRACTION = Regex("\\.\\d+")

    fun date(raw: String?): Instant? {
        if (raw.isNullOrEmpty()) return null
        return try {
            OffsetDateTime.parse(raw.replace(FRACTION, "")).toInstant()
        } catch (e: DateTimeParseException) {
            null
        }
    }

    fun relative(raw: String?, now: Instant = Instant.now()): String {
        val date = date(raw) ?: return "—"
        val seconds = Duration.between(date, now).seconds
        if (seconds < MINUTE) return "방금"
        val amount = when {
            seconds < HOUR -> "${seconds / MINUTE}분"
            seconds < DAY -> "${seconds / HOUR}시간"
            seconds < WEEK -> "${seconds / DAY}일"
            seconds < MONTH -> "${seconds / WEEK}주"
            seconds < YEAR -> "${seconds / MONTH}개월"
            else -> "${seconds / YEAR}년"
        }
        return "$amount 전"
    }

    fun size(bytes: Long?): String? {
        bytes ?: return null
        if (bytes < KILOBYTE) return "$bytes bytes"
        val units = listOf("KB", "MB", "GB", "TB")
        var value = bytes / KILOBYTE
        var index = 0
        while (value >= KILOBYTE && index < units.lastIndex) {
            value /= KILOBYTE
            index += 1
        }
        val digits = index.coerceAtMost(2)
        val formatted = if (digits == 0) {
            value.roundToLong().toString()
        } else {
            String.format(Locale.US, "%.${digits}f", value).trimEnd('0').trimEnd('.')
        }
        return "$formatted ${units[index]}"
    }

    fun stateDetail(session: AgentSession, now: Instant = Instant.now()): String = when (session.state) {
        AgentSessionState.ACTIVE -> "환경이 떠 있어요 · 마지막 사용 ${relative(session.lastAccess, now)}"
        AgentSessionState.IDLE -> "환경은 떠 있지만 쉬는 중 · 마지막 사용 ${relative(session.lastAccess, now)}"
        AgentSessionState.SNAPSHOT -> listOfNotNull(
            "아카이브 ${relative(session.checkpoint?.createdAt, now)}",
            size(session.checkpoint?.sizeBytes),
            "환경 반납됨",
        ).joinToString(" · ")
    }

    fun duration(seconds: Long): String {
        val total = seconds.coerceAtLeast(0)
        if (total < MINUTE) return "${total}초"
        if (total < HOUR) {
            val rest = total % MINUTE
            return if (rest == 0L) "${total / MINUTE}분" else "${total / MINUTE}분 ${rest}초"
        }
        val minutes = (total % HOUR) / MINUTE
        return if (minutes == 0L) "${total / HOUR}시간" else "${total / HOUR}시간 ${minutes}분"
    }

    private const val MINUTE = 60L
    private const val HOUR = 3_600L
    private const val DAY = 86_400L
    private const val WEEK = 604_800L
    private const val MONTH = 2_592_000L
    private const val YEAR = 31_536_000L
}

object AgentSessionTerminal {
    private val escapes = Regex(
        "\u001B\\][^\u0007\u001B]*(?:\u0007|\u001B\\\\)|\u001B\\[[0-?]*[ -/]*[@-~]|\u001B[0-?@-Z\\\\-_]",
    )

    fun clean(raw: String): String = escapes.replace(raw, "").replace("\r\n", "\n").replace("\r", "")

    fun command(input: String): String = if (input.endsWith("\n")) input else input + "\n"
}

@Serializable
internal data class AgentSessionListResponse(val sessions: List<AgentSession> = emptyList())

@Serializable
internal data class AgentSessionCreatePayload(
    val name: String,
    val workloadType: String,
    val model: String? = null,
)

@Serializable
internal data class AgentSessionReadPayload(val offset: Long)

@Serializable
internal data class AgentSessionWritePayload(val payload: String)

@Serializable
internal class AgentSessionEmptyPayload

suspend fun ApiClient.agentSessionConfig(): AgentSessionConfig =
    get("/api/sessions/config", AgentSessionConfigBody.serializer()).toConfig()

suspend fun ApiClient.listAgentSessions(): List<AgentSession> =
    get("/api/sessions", AgentSessionListResponse.serializer()).sessions

suspend fun ApiClient.createAgentSession(name: String, workload: AgentSessionWorkloadType, model: String?): AgentSession {
    val trimmedModel = if (workload.takesModel) model?.trim() else null
    val payload = AgentSessionCreatePayload(
        name = name.trim(),
        workloadType = workload.type,
        model = trimmedModel?.takeIf { it.isNotEmpty() },
    )
    return post("/api/sessions", payload, AgentSessionCreatePayload.serializer(), AgentSession.serializer())
}

suspend fun ApiClient.agentSession(id: String): AgentSession =
    get("/api/sessions/$id", AgentSession.serializer())

suspend fun ApiClient.readAgentSession(id: String, offset: Long): AgentSessionReadResult =
    post(
        "/api/sessions/$id/read",
        AgentSessionReadPayload(offset),
        AgentSessionReadPayload.serializer(),
        AgentSessionReadResult.serializer(),
    )

suspend fun ApiClient.writeAgentSession(id: String, prompt: String): AgentSessionWriteResult =
    post(
        "/api/sessions/$id/write",
        AgentSessionWritePayload(prompt),
        AgentSessionWritePayload.serializer(),
        AgentSessionWriteResult.serializer(),
    )

suspend fun ApiClient.restoreAgentSession(id: String): AgentSession =
    post(
        "/api/sessions/$id/switch",
        AgentSessionEmptyPayload(),
        AgentSessionEmptyPayload.serializer(),
        AgentSession.serializer(),
    )

suspend fun ApiClient.snapshotAgentSession(id: String): AgentSession =
    post(
        "/api/sessions/$id/snapshot",
        AgentSessionEmptyPayload(),
        AgentSessionEmptyPayload.serializer(),
        AgentSession.serializer(),
    )

suspend fun ApiClient.deleteAgentSession(id: String) {
    delete("/api/sessions/$id")
}
