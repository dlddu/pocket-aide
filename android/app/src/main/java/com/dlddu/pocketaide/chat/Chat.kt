package com.dlddu.pocketaide.chat

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import com.dlddu.pocketaide.api.isAuthPoolExhausted
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
data class ChatQueued(
    val content: String = "",
    @SerialName("queued_at") val queuedAt: Long = 0,
)

@Serializable
data class ChatView(
    val conversation: ChatConversation,
    val messages: List<ChatMessage> = emptyList(),
    val pending: ChatPending? = null,
    val queued: ChatQueued? = null,
)

/** A send either went out at once or, with a reply already in progress, is
 * held to go right after (PRD-1 AC1, 2026-11-09 개정). */
sealed class ChatSendOutcome {
    data class Sent(val message: ChatMessage) : ChatSendOutcome()
    data class Queued(val queued: ChatQueued) : ChatSendOutcome()
}

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
internal data class ChatSendResponse(val message: ChatMessage? = null, val queued: ChatQueued? = null)

@Serializable
internal class ChatEmptyPayload

enum class ChatError(val message: String) {
    BUSY("이전 답이 끝난 뒤에 다시 해 주세요."),
    NO_CONVERSATION("먼저 메시지를 한 번 보내 주세요."),
    TOO_LARGE("메시지가 너무 깁니다."),
    UNREACHABLE("세션 플랫폼에 연결하지 못했어요. 잠시 뒤 다시 시도해 주세요."),
    SLOTS_BUSY("로그인 슬롯이 모두 사용 중이에요. 다른 세션이 끝난 뒤 다시 시도해 주세요."),
    FAILED("요청을 처리하지 못했어요."),
    ;

    companion object {
        fun from(e: Exception, refreshing: Boolean = false): ChatError = when (e) {
            is ApiException.BadStatus -> when (e.status) {
                409 -> if (refreshing && e.body.contains("send a message first")) NO_CONVERSATION else BUSY
                413 -> TOO_LARGE
                503 -> if (e.isAuthPoolExhausted()) SLOTS_BUSY else UNREACHABLE
                502, 504 -> UNREACHABLE
                else -> FAILED
            }
            is ApiException.Transport -> UNREACHABLE
            else -> FAILED
        }
    }
}

const val CHAT_MAX_MESSAGE_BYTES = 16 * 1024

private val CHAT_URL = Regex("""https?://[^\s<>"'`]+""")
private const val CHAT_URL_TRAILING = ".,;:!?'\""
private val CHAT_URL_CLOSERS = mapOf(')' to '(', ']' to '[', '}' to '{')

/**
 * Where the web addresses in a message are, so the screen can make them
 * tappable — the ChatGPT sign-in a codex session asks for arrives as one.
 * Punctuation that ends a sentence is not part of the address, and neither is
 * a closing bracket the address never opened — so `(https://x.io/a)를` and
 * `[PR](https://x.io/pull/1)을` stop before the `)`, while
 * `https://x.io/wiki/A_(b)` keeps its own pair.
 */
fun chatLinkRanges(text: String): List<IntRange> =
    CHAT_URL.findAll(text).map { match -> match.range.first..chatLinkEnd(text, match.range) }
        .filter { it.last - it.first + 1 > "https://".length }.toList()

private fun chatLinkEnd(text: String, match: IntRange): Int {
    val open = mutableMapOf<Char, Int>()
    var end = match.last
    for (i in match) {
        val c = text[i]
        val opener = CHAT_URL_CLOSERS[c]
        if (c in CHAT_URL_CLOSERS.values) {
            open[c] = (open[c] ?: 0) + 1
        } else if (opener != null) {
            val depth = open[opener] ?: 0
            if (depth == 0) {
                end = i - 1
                break
            }
            open[opener] = depth - 1
        }
    }
    while (end > match.first && text[end] in CHAT_URL_TRAILING) end--
    return end
}

suspend fun ApiClient.chat(): ChatView = get("/api/chat", ChatView.serializer())

/** Always safe to call whether or not a reply is in progress: the server
 * sends at once when it can, and otherwise holds the message to send right
 * after (PRD-1 AC1, 2026-11-09 개정). */
suspend fun ApiClient.sendChatMessage(content: String): ChatSendOutcome {
    val response = post(
        "/api/chat/messages",
        ChatSendPayload(content),
        ChatSendPayload.serializer(),
        ChatSendResponse.serializer(),
    )
    response.message?.let { return ChatSendOutcome.Sent(it) }
    response.queued?.let { return ChatSendOutcome.Queued(it) }
    error("chat send answered with neither a message nor a queued one")
}

/** Drops the message held for after the reply in progress, if it has not
 * already been sent. */
suspend fun ApiClient.cancelQueuedChatMessage() {
    delete("/api/chat/queue")
}

suspend fun ApiClient.refreshChatPlugins(): ChatPluginRefresh =
    post(
        "/api/chat/refresh-plugins",
        ChatEmptyPayload(),
        ChatEmptyPayload.serializer(),
        ChatPluginRefresh.serializer(),
    )
