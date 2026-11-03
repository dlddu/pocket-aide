package com.dlddu.pocketaide.sessions

import kotlinx.serialization.Serializable

enum class AgentSessionInput(val wire: String) {
    PROMPT("prompt"),
    TERMINAL("terminal");

    companion object {
        fun fromWire(raw: String?): AgentSessionInput = entries.firstOrNull { it.wire == raw } ?: PROMPT
    }
}

data class AgentSessionModelCatalog(val defaultModel: String, val models: List<String>) {
    val initialPick: String?
        get() {
            if (models.isEmpty()) return null
            return if (defaultModel in models) defaultModel else models.first()
        }
}

data class AgentSessionWorkloadType(
    val type: String,
    val label: String? = null,
    val description: String? = null,
    val input: AgentSessionInput = AgentSessionInput.PROMPT,
    val model: AgentSessionModelCatalog? = null,
) {
    val id: String get() = type
    val displayName: String get() = label?.takeIf { it.isNotEmpty() } ?: type
    val takesModel: Boolean get() = model != null
}

data class AgentSessionConfig(val workloadTypes: List<AgentSessionWorkloadType>) {
    fun type(id: String): AgentSessionWorkloadType? = workloadTypes.firstOrNull { it.type == id }

    val initialType: AgentSessionWorkloadType?
        get() = workloadTypes.firstOrNull { it.input == AgentSessionInput.PROMPT } ?: workloadTypes.firstOrNull()
}

@Serializable
internal data class AgentSessionModelCatalogBody(
    val defaultModel: String? = null,
    val models: List<String>? = null,
)

@Serializable
internal data class AgentSessionWorkloadTypeBody(
    val type: String,
    val label: String? = null,
    val description: String? = null,
    val input: String? = null,
    val takesModel: Boolean = false,
    val defaultModel: String? = null,
    val models: List<String>? = null,
) {
    fun toWorkloadType(): AgentSessionWorkloadType = AgentSessionWorkloadType(
        type = type,
        label = label,
        description = description,
        input = AgentSessionInput.fromWire(input),
        model = if (takesModel) AgentSessionModelCatalog(defaultModel.orEmpty(), models.orEmpty()) else null,
    )
}

@Serializable
internal data class AgentSessionConfigBody(
    val workloadTypes: List<AgentSessionWorkloadTypeBody>? = null,
    val claudeCode: AgentSessionModelCatalogBody? = null,
) {
    fun toConfig(): AgentSessionConfig {
        val listed = workloadTypes.orEmpty().map { it.toWorkloadType() }
        val legacy = claudeCode
        if (listed.isEmpty() && legacy != null) {
            val catalog = AgentSessionModelCatalog(legacy.defaultModel.orEmpty(), legacy.models.orEmpty())
            return AgentSessionConfig(listOf(AgentSessionWorkloadType(type = AgentSession.DEFAULT_WORKLOAD_TYPE, model = catalog)))
        }
        return AgentSessionConfig(listed)
    }
}
