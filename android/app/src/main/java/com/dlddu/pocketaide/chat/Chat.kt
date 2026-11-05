package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.api.SseEvent
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Locale

@Serializable
enum class ChatRole {
    @SerialName("user")
    USER,

    @SerialName("assistant")
    ASSISTANT,
}

@Serializable
data class ChatSession(
    val id: Long,
    val title: String = "",
    @SerialName("created_at") val createdAt: Long = 0,
    @SerialName("updated_at") val updatedAt: Long = 0,
)

@Serializable
data class ChatMessage(
    val id: Long,
    val role: ChatRole,
    val content: String,
    @SerialName("created_at") val createdAt: Long = 0,
)

data class ChatConversation(val session: ChatSession, val messages: List<ChatMessage>)

@Serializable
internal data class ChatSessionListResponse(val items: List<ChatSession>? = null)

@Serializable
internal data class ChatSessionDetailResponse(val session: ChatSession, val messages: List<ChatMessage>? = null)

@Serializable
internal class ChatCreatePayload

@Serializable
internal data class ChatSendPayload(val content: String)

@Serializable
internal data class ChatDeltaPayload(val delta: String = "")

@Serializable
internal data class ChatDonePayload(val message: ChatMessage? = null)

@Serializable
internal data class ChatErrorPayload(val message: String = "")

sealed interface ChatStreamEvent {
    data class Delta(val text: String) : ChatStreamEvent

    data class Done(val message: ChatMessage?) : ChatStreamEvent

    data class Failed(val message: String) : ChatStreamEvent
}

object ChatStreamEvents {
    fun decode(event: SseEvent): ChatStreamEvent? = try {
        when (event.event) {
            "message" -> ChatStreamEvent.Delta(ApiClient.json.decodeFromString<ChatDeltaPayload>(event.data).delta)
            "done" -> ChatStreamEvent.Done(ApiClient.json.decodeFromString<ChatDonePayload>(event.data).message)
            "error" -> ChatStreamEvent.Failed(ApiClient.json.decodeFromString<ChatErrorPayload>(event.data).message)
            else -> null
        }
    } catch (e: IllegalArgumentException) {
        null
    }
}

suspend fun ApiClient.listChatSessions(): List<ChatSession> =
    get("/api/chat/sessions", ChatSessionListResponse.serializer()).items.orEmpty()

suspend fun ApiClient.createChatSession(): ChatSession =
    post("/api/chat/sessions", ChatCreatePayload(), ChatCreatePayload.serializer(), ChatSession.serializer())

suspend fun ApiClient.getChatConversation(id: Long): ChatConversation {
    val detail = get("/api/chat/sessions/$id", ChatSessionDetailResponse.serializer())
    return ChatConversation(detail.session, detail.messages.orEmpty())
}

suspend fun ApiClient.sendChatMessage(sessionId: Long, content: String, onEvent: (ChatStreamEvent) -> Unit) {
    postStream("/api/chat/sessions/$sessionId/messages", ChatSendPayload(content), ChatSendPayload.serializer()) { raw ->
        ChatStreamEvents.decode(raw)?.let(onEvent)
    }
}

object ChatCopy {
    const val UNCONFIGURED = "채팅이 아직 설정되지 않았어요. 서버에 AI 키가 등록되면 쓸 수 있어요."
    const val REPLY_FAILED = "답변을 끝까지 받지 못했어요. 다시 보내 주세요."
    const val UNTITLED = "새 대화"

    fun sessionTitle(session: ChatSession): String = session.title.ifBlank { UNTITLED }

    fun messageCount(count: Int): String = "${count}개의 메시지"

    fun errorMessage(e: Exception): String = when {
        e is ApiException.BadStatus && e.status == 503 -> UNCONFIGURED
        else -> e.message ?: e.toString()
    }
}

object ChatTimes {
    private val timeFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("HH:mm", Locale.KOREA)

    fun dayLabel(epochSeconds: Long, today: LocalDate, zone: ZoneId = ZoneId.systemDefault()): String {
        val day = Instant.ofEpochSecond(epochSeconds).atZone(zone).toLocalDate()
        return when (day) {
            today -> "오늘"
            today.minusDays(1) -> "어제"
            else -> "${day.monthValue}월 ${day.dayOfMonth}일"
        }
    }

    fun timeLabel(epochSeconds: Long, zone: ZoneId = ZoneId.systemDefault()): String =
        Instant.ofEpochSecond(epochSeconds).atZone(zone).format(timeFormatter)

    fun stamp(epochSeconds: Long, today: LocalDate, zone: ZoneId = ZoneId.systemDefault()): String =
        "${dayLabel(epochSeconds, today, zone)} · ${timeLabel(epochSeconds, zone)}"
}
