package com.dlddu.pocketaide.mail

import com.dlddu.pocketaide.api.ApiClient
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class MailMessage(
    val uid: Long = 0,
    val from: String = "",
    val subject: String = "",
    @SerialName("received_at") val receivedAt: String = "",
    val preview: String = "",
    @SerialName("account_id") val accountId: Long = 0,
    @SerialName("account_email") val accountEmail: String = "",
)

@Serializable
data class MailAccountStatus(
    @SerialName("account_id") val accountId: Long = 0,
    val status: String = "",
)

@Serializable
data class MailMessagesResponse(
    val items: List<MailMessage> = emptyList(),
    val accounts: List<MailAccountStatus> = emptyList(),
)

suspend fun ApiClient.listMailMessages(accountId: Long = 0): MailMessagesResponse {
    val path = if (accountId > 0) "/api/mail/messages?account_id=$accountId" else "/api/mail/messages"
    return get(path, MailMessagesResponse.serializer())
}
