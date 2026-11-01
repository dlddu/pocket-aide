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

enum class AffirmationSortOrder(val wire: String, val label: String) {
    PRIORITY("priority", "우선순위 순"),
    NEWEST("newest", "최신순"),
    ;

    fun sorted(items: List<Affirmation>): List<Affirmation> = when (this) {
        PRIORITY -> items.sortedWith(compareBy<Affirmation> { rank(it.priority) }.then(NEWER_FIRST))
        NEWEST -> items.sortedWith(NEWER_FIRST)
    }

    private companion object {
        val NEWER_FIRST: Comparator<Affirmation> =
            compareByDescending<Affirmation> { it.createdAt }.thenByDescending { it.id }

        fun rank(priority: AffirmationPriority): Int = when (priority) {
            AffirmationPriority.HIGH -> 0
            AffirmationPriority.NORMAL -> 1
            AffirmationPriority.LOW -> 2
        }
    }
}

@Serializable
internal data class AffirmationsListResponse(val items: List<Affirmation>)

@Serializable
internal data class AffirmationPayload(val text: String, val priority: AffirmationPriority)

suspend fun ApiClient.listAffirmations(): List<Affirmation> =
    get("/api/affirmations", AffirmationsListResponse.serializer()).items

suspend fun ApiClient.createAffirmation(text: String, priority: AffirmationPriority): Affirmation =
    post(
        "/api/affirmations",
        AffirmationPayload(text = text, priority = priority),
        AffirmationPayload.serializer(),
        Affirmation.serializer(),
    )

suspend fun ApiClient.updateAffirmation(id: Long, text: String, priority: AffirmationPriority): Affirmation =
    patch(
        "/api/affirmations/$id",
        AffirmationPayload(text = text, priority = priority),
        AffirmationPayload.serializer(),
        Affirmation.serializer(),
    )

suspend fun ApiClient.deleteAffirmation(id: Long) {
    delete("/api/affirmations/$id")
}
