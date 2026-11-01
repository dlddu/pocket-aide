package com.dlddu.pocketaide.affirmations

import com.dlddu.pocketaide.api.ApiClient
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
enum class AffirmationPriority(val wire: String, val displayName: String) {
    @SerialName("high")
    HIGH("high", "높음"),

    @SerialName("normal")
    NORMAL("normal", "보통"),

    @SerialName("low")
    LOW("low", "가끔"),
}

@Serializable
data class Affirmation(
    val id: Long,
    val text: String,
    val priority: AffirmationPriority,
    @SerialName("created_at") val createdAt: Long,
    @SerialName("updated_at") val updatedAt: Long,
)

@Serializable
internal data class AffirmationPayload(val text: String, val priority: AffirmationPriority)

suspend fun ApiClient.updateAffirmation(id: Long, text: String, priority: AffirmationPriority): Affirmation =
    patch(
        "/api/affirmations/$id",
        AffirmationPayload(text = text, priority = priority),
        AffirmationPayload.serializer(),
        Affirmation.serializer(),
    )
