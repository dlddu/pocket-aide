package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// PRD-1: the chat tab is one long conversation with the user's codex session
// on session-platform. The backend answers a send at once (202) and the reply
// arrives through GET /api/chat, so the screen polls while one is pending.

@Serializable
data class ChatConversation(
    val id: Long,
    val title: String = "",
    @SerialName("created_at") val createdAt: Long = 0,
    @SerialName("updated_at") val updatedAt: Long = 0,
)

@Serializable
data class ChatMessage(
    val id: Long,
    val role: String,
    val content: String,
    @SerialName("created_at") val createdAt: Long = 0,
) {
    val isMine: Boolean get() = role == ROLE_USER

    companion object {
        const val ROLE_USER = "user"
    }
}

@Serializable
data class ChatPending(
    val content: String = "",
    @SerialName("started_at") val startedAt: Long = 0,
)

@Serializable
data class ChatView(
    val conversation: ChatConversation,
    val messages: List<ChatMessage> = emptyList(),
    val pending: ChatPending? = null,
)

@Serializable
data class ChatPluginRefresh(
    val state: String = "",
    val revision: String? = null,
    val added: List<String> = emptyList(),
    val updated: List<String> = emptyList(),
    val removed: List<String> = emptyList(),
) {
    val changedCount: Int get() = added.size + updated.size + removed.size
}

@Serializable
internal data class ChatSendPayload(val content: String)

@Serializable
internal data class ChatSendResponse(val message: ChatMessage)

@Serializable
internal class ChatEmptyPayload

enum class ChatError(val message: String) {
    BUSY("이전 답이 끝난 뒤에 다시 해 주세요."),
    NO_CONVERSATION("먼저 메시지를 한 번 보내 주세요."),
    TOO_LARGE("메시지가 너무 깁니다."),
    UNREACHABLE("세션 플랫폼에 연결하지 못했어요. 잠시 뒤 다시 시도해 주세요."),
    FAILED("요청을 처리하지 못했어요."),
    ;

    companion object {
        fun from(e: Exception, refreshing: Boolean = false): ChatError = when (e) {
            is ApiException.BadStatus -> when (e.status) {
                409 -> if (refreshing && e.body.contains("send a message first")) NO_CONVERSATION else BUSY
                413 -> TOO_LARGE
                502, 503, 504 -> UNREACHABLE
                else -> FAILED
            }
            is ApiException.Transport -> UNREACHABLE
            else -> FAILED
        }
    }
}

const val CHAT_MAX_MESSAGE_BYTES = 16 * 1024

suspend fun ApiClient.chat(): ChatView = get("/api/chat", ChatView.serializer())

suspend fun ApiClient.sendChatMessage(content: String): ChatMessage =
    post(
        "/api/chat/messages",
        ChatSendPayload(content),
        ChatSendPayload.serializer(),
        ChatSendResponse.serializer(),
    ).message

suspend fun ApiClient.refreshChatPlugins(): ChatPluginRefresh =
    post(
        "/api/chat/refresh-plugins",
        ChatEmptyPayload(),
        ChatEmptyPayload.serializer(),
        ChatPluginRefresh.serializer(),
    )
