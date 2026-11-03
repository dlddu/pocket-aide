package com.dlddu.pocketaide.scratchpad

import com.dlddu.pocketaide.affirmations.Affirmation
import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.todos.TodoItem
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonElement
import java.time.Clock
import java.time.Instant
import java.time.LocalDate
import java.time.ZoneId
import java.time.format.DateTimeFormatter
import java.util.Locale

@Serializable
enum class ScratchpadSource(val wire: String, val displayName: String) {
    @SerialName("text")
    TEXT("text", "탭 내 입력"),

    @SerialName("voice")
    VOICE("voice", "음성"),

    @SerialName("shortcut")
    SHORTCUT("shortcut", "숏컷 · 음성"),
}

@Serializable
data class ScratchpadItem(
    val id: Long,
    val text: String,
    val source: ScratchpadSource = ScratchpadSource.TEXT,
    @SerialName("captured_at") val capturedAt: Long,
    @SerialName("created_at") val createdAt: Long = 0,
)

@Serializable
enum class ScratchpadMoveTarget(val wire: String, val chipLabel: String) {
    @SerialName("personal")
    PERSONAL("personal", "→ 개인"),

    @SerialName("work")
    WORK("work", "→ 회사"),

    @SerialName("affirmation")
    AFFIRMATION("affirmation", "→ 다짐"),

    @SerialName("routine")
    ROUTINE("routine", "→ 루틴"),
}

@Serializable
data class ScratchpadMoveResult(
    val target: ScratchpadMoveTarget,
    val todo: TodoItem? = null,
    val affirmation: Affirmation? = null,
    val routine: JsonElement? = null,
)

data class ScratchpadSection(val title: String, val items: List<ScratchpadItem>) {
    val id: String get() = title
}

object ScratchpadSections {
    private val timeFormatter: DateTimeFormatter = DateTimeFormatter.ofPattern("HH:mm", Locale.KOREA)

    fun group(items: List<ScratchpadItem>, clock: Clock = Clock.systemDefaultZone()): List<ScratchpadSection> {
        val zone = clock.zone
        val today = LocalDate.now(clock)
        val sorted = items.sortedWith(
            compareByDescending<ScratchpadItem> { it.capturedAt }.thenByDescending { it.id },
        )
        val sections = mutableListOf<ScratchpadSection>()
        var currentDay: LocalDate? = null
        var bucket = mutableListOf<ScratchpadItem>()
        for (item in sorted) {
            val day = Instant.ofEpochSecond(item.capturedAt).atZone(zone).toLocalDate()
            val open = currentDay
            if (open != null && day != open) {
                sections.add(ScratchpadSection(title(open, today), bucket))
                bucket = mutableListOf()
            }
            currentDay = day
            bucket.add(item)
        }
        currentDay?.let { sections.add(ScratchpadSection(title(it, today), bucket)) }
        return sections
    }

    fun timeLabel(item: ScratchpadItem, zone: ZoneId = ZoneId.systemDefault()): String =
        Instant.ofEpochSecond(item.capturedAt).atZone(zone).format(timeFormatter)

    internal fun title(day: LocalDate, today: LocalDate): String = when (day) {
        today -> "오늘"
        today.minusDays(1) -> "어제"
        else -> "${day.monthValue}월 ${day.dayOfMonth}일"
    }
}

object WidgetScratchpad {
    fun countText(unclassified: Int): String =
        if (unclassified > 0) "미분류 ${unclassified}개" else "정리할 항목 없음"
}

@Serializable
internal data class ScratchpadListResponse(val items: List<ScratchpadItem>)

@Serializable
internal data class ScratchpadCreatePayload(val text: String, val source: ScratchpadSource)

@Serializable
internal data class ScratchpadMovePayload(val target: ScratchpadMoveTarget)

suspend fun ApiClient.listScratchpad(): List<ScratchpadItem> =
    get("/api/scratchpad", ScratchpadListResponse.serializer()).items

suspend fun ApiClient.createScratchpadItem(text: String, source: ScratchpadSource = ScratchpadSource.TEXT): ScratchpadItem =
    post(
        "/api/scratchpad",
        ScratchpadCreatePayload(text = text, source = source),
        ScratchpadCreatePayload.serializer(),
        ScratchpadItem.serializer(),
    )

suspend fun ApiClient.deleteScratchpadItem(id: Long) {
    delete("/api/scratchpad/$id")
}

suspend fun ApiClient.moveScratchpadItem(id: Long, target: ScratchpadMoveTarget): ScratchpadMoveResult =
    post(
        "/api/scratchpad/$id/move",
        ScratchpadMovePayload(target = target),
        ScratchpadMovePayload.serializer(),
        ScratchpadMoveResult.serializer(),
    )
