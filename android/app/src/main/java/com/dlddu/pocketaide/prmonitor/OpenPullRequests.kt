package com.dlddu.pocketaide.prmonitor

import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.SerializationException
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.IOException
import java.time.Instant
import java.time.OffsetDateTime
import java.time.format.DateTimeParseException

enum class PullRequestRole(val label: String) {
    AUTHOR("작성자"),
    REVIEWER("리뷰어"),
}

enum class PullRequestCiStatus(val label: String) {
    IN_PROGRESS("진행 중"),
    SUCCESS("성공"),
    FAILURE("실패"),
    NO_CHECKS("상태 없음");

    companion object {
        fun fromRollup(state: String?): PullRequestCiStatus = when (state) {
            "SUCCESS" -> SUCCESS
            "FAILURE", "ERROR" -> FAILURE
            "PENDING", "EXPECTED" -> IN_PROGRESS
            else -> NO_CHECKS
        }
    }
}

data class OpenPullRequest(
    val repository: String,
    val number: Int,
    val title: String,
    val author: String,
    val url: String?,
    val updatedAt: Instant,
    val role: PullRequestRole,
    val ciStatus: PullRequestCiStatus,
) {
    val id: String get() = "$repository#$number"
}

enum class OpenPullRequestFilter(val wire: String, val label: String) {
    ALL("all", "전체"),
    MINE("mine", "내 PR만"),
    REVIEW_REQUESTED("reviewRequested", "리뷰 요청만"),
    CI_FAILING("ciFailing", "CI 실패만");

    fun matches(pullRequest: OpenPullRequest): Boolean = when (this) {
        ALL -> true
        MINE -> pullRequest.role == PullRequestRole.AUTHOR
        REVIEW_REQUESTED -> pullRequest.role == PullRequestRole.REVIEWER
        CI_FAILING -> pullRequest.ciStatus == PullRequestCiStatus.FAILURE
    }

    fun apply(pullRequests: List<OpenPullRequest>): List<OpenPullRequest> = pullRequests.filter(::matches)

    companion object {
        fun fromWire(raw: String?): OpenPullRequestFilter? = entries.firstOrNull { it.wire == raw }
    }
}

data class OpenPullRequestsResult(
    val pullRequests: List<OpenPullRequest>,
    val inaccessibleCount: Int,
)

sealed class GitHubError(val description: String) : Exception(description) {
    data object Unauthorized : GitHubError("GitHub가 토큰을 거부했습니다(만료·폐기·오타). 토큰을 바꿔 다시 연결하세요.")
    data object Forbidden : GitHubError("GitHub가 이 토큰의 요청을 거부했습니다(권한 부족).")
    data class RateLimited(val resetAt: Instant?) : GitHubError("GitHub 요청 한도에 도달했습니다.")
    data class BadStatus(val status: Int) : GitHubError("GitHub 응답 오류 (HTTP $status)")
    data class GraphQl(val detail: String) : GitHubError("GitHub 조회 오류: $detail")
    data class Decoding(val detail: String) : GitHubError("GitHub 응답을 해석하지 못했습니다: $detail")
    data class Transport(val detail: String) : GitHubError("GitHub에 연결하지 못했습니다: $detail")

    companion object {
        fun classify(
            statusCode: Int,
            remaining: String? = null,
            reset: String? = null,
            retryAfter: String? = null,
            now: Instant = Instant.now(),
        ): GitHubError? {
            if (statusCode in 200..299) return null
            if (statusCode == 401) return Unauthorized
            if (statusCode == 403 || statusCode == 429) {
                retryAfter?.toDoubleOrNull()?.let { seconds ->
                    return RateLimited(now.plusMillis((seconds * MILLIS_PER_SECOND).toLong()))
                }
                if (remaining == "0" || statusCode == 429) {
                    val resetAt = reset?.toDoubleOrNull()?.let { Instant.ofEpochMilli((it * MILLIS_PER_SECOND).toLong()) }
                    return RateLimited(resetAt)
                }
                return Forbidden
            }
            return BadStatus(statusCode)
        }

        private const val MILLIS_PER_SECOND = 1000.0
    }
}

sealed class GitHubAlert {
    data object TokenRejected : GitHubAlert()
    data class InsufficientAccess(val hiddenCount: Int) : GitHubAlert()
    data class RateLimited(val resetAt: Instant?) : GitHubAlert()

    val kind: String
        get() = when (this) {
            TokenRejected -> "token"
            is InsufficientAccess -> "access"
            is RateLimited -> "ratelimit"
        }

    val title: String
        get() = when (this) {
            TokenRejected -> "GitHub 토큰이 거부되었습니다"
            is InsufficientAccess -> "접근 권한이 부족합니다"
            is RateLimited -> "GitHub 요청 한도에 도달했습니다"
        }

    fun message(formatTime: (Instant) -> String): String = when (this) {
        TokenRejected -> "토큰이 만료·폐기되었거나 잘못 입력되었습니다. 새 토큰으로 다시 연결하세요."
        is InsufficientAccess -> {
            val lead = if (hiddenCount > 0) {
                "PR ${hiddenCount}개는 이 토큰으로 볼 수 없어 표시하지 않았습니다."
            } else {
                "이 토큰으로는 GitHub 조회가 허용되지 않습니다."
            }
            "$lead 조직 SSO 승인이나 토큰 범위(classic은 repo)를 확인한 뒤 토큰을 바꾸세요."
        }
        is RateLimited -> resetAt?.let { "${formatTime(it)} 이후에 다시 시도하세요." } ?: "잠시 후 다시 시도하세요."
    }

    val actionTitle: String
        get() = when (this) {
            TokenRejected -> "토큰 다시 연결"
            is InsufficientAccess -> "토큰 바꾸기"
            is RateLimited -> "다시 시도"
        }

    companion object {
        fun make(error: GitHubError?, inaccessibleCount: Int): GitHubAlert? = when (error) {
            GitHubError.Unauthorized -> TokenRejected
            GitHubError.Forbidden -> InsufficientAccess(inaccessibleCount)
            is GitHubError.RateLimited -> RateLimited(error.resetAt)
            else -> if (inaccessibleCount > 0) InsufficientAccess(inaccessibleCount) else null
        }
    }
}

object OpenPullRequests {
    val query = """
        query(${'$'}authored: String!, ${'$'}requested: String!, ${'$'}reviewed: String!) {
          authored: search(query: ${'$'}authored, type: ISSUE, first: 50) { nodes { ...OpenPR } }
          requested: search(query: ${'$'}requested, type: ISSUE, first: 50) { nodes { ...OpenPR } }
          reviewed: search(query: ${'$'}reviewed, type: ISSUE, first: 50) { nodes { ...OpenPR } }
        }
        fragment OpenPR on PullRequest {
          number
          title
          url
          updatedAt
          repository { nameWithOwner }
          author { login }
          commits(last: 1) { nodes { commit { statusCheckRollup { state } } } }
        }
    """.trimIndent()

    fun variables(login: String): Map<String, String> {
        val base = "is:pr is:open archived:false"
        return mapOf(
            "authored" to "$base author:$login",
            "requested" to "$base review-requested:$login",
            "reviewed" to "$base reviewed-by:$login -author:$login",
        )
    }

    fun decode(raw: String): OpenPullRequestsResult {
        val response = try {
            json.decodeFromString(SearchResponse.serializer(), raw)
        } catch (e: SerializationException) {
            throw GitHubError.Decoding(e.message ?: e.toString())
        } catch (e: IllegalArgumentException) {
            throw GitHubError.Decoding(e.message ?: e.toString())
        }
        val payload = response.data
        if (payload == null) {
            if (response.errors.orEmpty().any { it.type == "RATE_LIMITED" }) throw GitHubError.RateLimited(null)
            throw GitHubError.GraphQl(response.errors?.firstOrNull()?.message ?: "data 없음")
        }
        val sources = listOf(
            payload.authored to PullRequestRole.AUTHOR,
            payload.requested to PullRequestRole.REVIEWER,
            payload.reviewed to PullRequestRole.REVIEWER,
        )
        val seen = mutableSetOf<String>()
        val merged = mutableListOf<OpenPullRequest>()
        var inaccessible = 0
        for ((connection, role) in sources) {
            for (node in connection?.nodes.orEmpty()) {
                val pullRequest = node?.toPullRequest(role)
                if (pullRequest == null) {
                    inaccessible += 1
                    continue
                }
                if (seen.add(pullRequest.id)) merged.add(pullRequest)
            }
        }
        return OpenPullRequestsResult(merged.sortedByDescending { it.updatedAt }, inaccessible)
    }

    internal fun parseInstant(raw: String?): Instant? {
        if (raw.isNullOrEmpty()) return null
        return try {
            OffsetDateTime.parse(raw).toInstant()
        } catch (e: DateTimeParseException) {
            null
        }
    }

    internal val json = Json {
        ignoreUnknownKeys = true
        explicitNulls = false
    }
}

@Serializable
internal data class SearchResponse(
    val data: Payload? = null,
    val errors: List<GraphQlErrorBody>? = null,
) {
    @Serializable
    data class Payload(
        val authored: Connection? = null,
        val requested: Connection? = null,
        val reviewed: Connection? = null,
    )

    @Serializable
    data class Connection(val nodes: List<Node?> = emptyList())

    @Serializable
    data class Node(
        val number: Int? = null,
        val title: String? = null,
        val url: String? = null,
        val updatedAt: String? = null,
        val repository: Repository? = null,
        val author: Actor? = null,
        val commits: Commits? = null,
    ) {
        fun toPullRequest(role: PullRequestRole): OpenPullRequest? {
            val number = number ?: return null
            val title = title ?: return null
            val repository = repository ?: return null
            val updated = OpenPullRequests.parseInstant(updatedAt) ?: return null
            val rollup = commits?.nodes?.filterNotNull()?.lastOrNull()?.commit?.statusCheckRollup?.state
            return OpenPullRequest(
                repository = repository.nameWithOwner,
                number = number,
                title = title,
                author = author?.login ?: "ghost",
                url = url,
                updatedAt = updated,
                role = role,
                ciStatus = PullRequestCiStatus.fromRollup(rollup),
            )
        }
    }

    @Serializable
    data class Repository(val nameWithOwner: String)

    @Serializable
    data class Actor(val login: String)

    @Serializable
    data class Commits(val nodes: List<CommitNode?>? = null)

    @Serializable
    data class CommitNode(val commit: Commit? = null)

    @Serializable
    data class Commit(val statusCheckRollup: Rollup? = null)

    @Serializable
    data class Rollup(val state: String? = null)
}

@Serializable
internal data class GraphQlErrorBody(val type: String? = null, val message: String? = null)

@Serializable
internal data class GitHubViewer(val login: String)

@Serializable
internal data class GraphQlRequest(val query: String, val variables: Map<String, String>)

class GitHubClient(
    baseUrl: String = DEFAULT_BASE_URL,
    private val http: OkHttpClient = OkHttpClient(),
) {
    val baseUrl: HttpUrl = baseUrl.trimEnd('/').toHttpUrl()

    suspend fun viewerLogin(token: String): String {
        val raw = send(request("user", token).get().build())
        return try {
            OpenPullRequests.json.decodeFromString(GitHubViewer.serializer(), raw).login
        } catch (e: SerializationException) {
            throw GitHubError.Decoding(e.message ?: e.toString())
        } catch (e: IllegalArgumentException) {
            throw GitHubError.Decoding(e.message ?: e.toString())
        }
    }

    suspend fun openPullRequests(token: String, login: String): OpenPullRequestsResult {
        val body = OpenPullRequests.json.encodeToString(
            GraphQlRequest.serializer(),
            GraphQlRequest(OpenPullRequests.query, OpenPullRequests.variables(login)),
        )
        val request = request("graphql", token).post(body.toRequestBody(JSON_MEDIA_TYPE)).build()
        return OpenPullRequests.decode(send(request))
    }

    private fun request(path: String, token: String): Request.Builder =
        Request.Builder()
            .url(baseUrl.newBuilder().addPathSegment(path).build())
            .header("Authorization", "Bearer $token")
            .header("Accept", "application/vnd.github+json")
            .header("X-GitHub-Api-Version", API_VERSION)

    private suspend fun send(request: Request): String = withContext(Dispatchers.IO) {
        val response = try {
            http.newCall(request).execute()
        } catch (e: IOException) {
            throw GitHubError.Transport(e.message ?: e.toString())
        }
        response.use {
            GitHubError.classify(
                statusCode = it.code,
                remaining = it.header("X-RateLimit-Remaining"),
                reset = it.header("X-RateLimit-Reset"),
                retryAfter = it.header("Retry-After"),
            )?.let { error -> throw error }
            try {
                it.body?.string() ?: ""
            } catch (e: IOException) {
                throw GitHubError.Transport(e.message ?: e.toString())
            }
        }
    }

    companion object {
        const val DEFAULT_BASE_URL = "https://api.github.com"
        private const val API_VERSION = "2022-11-28"
        private val JSON_MEDIA_TYPE = "application/json".toMediaType()
    }
}
