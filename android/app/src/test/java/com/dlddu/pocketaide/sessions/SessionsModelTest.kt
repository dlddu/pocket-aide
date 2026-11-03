package com.dlddu.pocketaide.sessions

import com.dlddu.pocketaide.api.ApiClient
import com.dlddu.pocketaide.api.ApiException
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.io.IOException
import java.time.Instant

class SessionsModelTest {
    private fun result(payload: String, next: Long) = AgentSessionReadResult(payload = payload, nextOffset = next)

    @Test
    fun outputAppendsIncrementally() {
        val first = AgentSessionOutput().merge(result("abc", 3), 0)
        assertEquals(AgentSessionOutput.Outcome.GREW, first.outcome)

        val second = first.output.merge(result("def", 6), 3)
        assertEquals(AgentSessionOutput.Outcome.GREW, second.outcome)
        assertEquals("abcdef", second.output.text)
        assertEquals(6L, second.output.nextOffset)

        val idle = second.output.merge(result("", 6), 6)
        assertEquals(AgentSessionOutput.Outcome.UNCHANGED, idle.outcome)
    }

    @Test
    fun outputRequestsFullRereadWhenOffsetGoesBackwards() {
        val output = AgentSessionOutput(text = "abc", nextOffset = 10)

        val applied = output.merge(result("x", 2), 10)

        assertEquals(AgentSessionOutput.Outcome.NEEDS_FULL_REREAD, applied.outcome)
        assertEquals(output, applied.output)
    }

    @Test
    fun fullReadOfSameTextIsUnchanged() {
        val output = AgentSessionOutput(text = "abc", nextOffset = 3)

        assertEquals(AgentSessionOutput.Outcome.UNCHANGED, output.merge(result("abc", 3), 0).outcome)
    }

    @Test
    fun outputTrimsToDisplayLimit() {
        val big = "x".repeat(AgentSessionOutput.DISPLAY_LIMIT + 10)

        val applied = AgentSessionOutput().merge(result(big, big.length.toLong()), 0)

        assertTrue(applied.output.isTrimmed)
        assertEquals(AgentSessionOutput.DISPLAY_LIMIT, applied.output.text.length)
    }

    @Test
    fun phaseTransitions() {
        assertEquals(AgentSessionResponsePhase.STREAMING, AgentSessionResponsePhase.QUEUED.next(true, 0))
        assertEquals(AgentSessionResponsePhase.QUEUED, AgentSessionResponsePhase.QUEUED.next(false, 60_000))
        assertEquals(AgentSessionResponsePhase.QUIET, AgentSessionResponsePhase.STREAMING.next(false, 10_000))
        assertEquals(AgentSessionResponsePhase.STREAMING, AgentSessionResponsePhase.STREAMING.next(false, 9_999))
        assertEquals(AgentSessionResponsePhase.NONE, AgentSessionResponsePhase.NONE.next(true, 0))
        assertNull(AgentSessionResponsePhase.NONE.label)
    }

    @Test
    fun errorMapping() {
        assertEquals(AgentSessionError.BUSY, AgentSessionError.from(ApiException.BadStatus(409, "")))
        assertEquals(AgentSessionError.QUOTA_FULL, AgentSessionError.from(ApiException.BadStatus(507, "")))
        assertEquals(AgentSessionError.UNKNOWN, AgentSessionError.from(ApiException.BadStatus(500, "")))
        assertEquals(AgentSessionError.UNREACHABLE, AgentSessionError.from(ApiException.Transport(IOException("x"))))
        assertEquals(AgentSessionError.UNKNOWN, AgentSessionError.from(ApiException.Decoding(IllegalArgumentException("x"))))
        assertEquals(AgentSessionError.DISABLED, AgentSessionError.forSnapshot(ApiException.BadStatus(503, "")))
        assertEquals(AgentSessionError.UNREACHABLE, AgentSessionError.from(ApiException.BadStatus(503, "")))
    }

    @Test
    fun promptLimitCountsUtf8Bytes() {
        assertFalse(AgentSessionPrompt.exceedsLimit("a".repeat(AgentSessionPrompt.MAX_BYTES)))
        assertTrue(AgentSessionPrompt.exceedsLimit("가".repeat(AgentSessionPrompt.MAX_BYTES / 3 + 1)))
    }

    @Test
    fun formatting() {
        val now = Instant.parse("2026-10-23T12:00:00Z")
        assertEquals("—", AgentSessionFormatting.relative(null, now))
        assertEquals("방금", AgentSessionFormatting.relative("2026-10-23T11:59:30Z", now))
        assertEquals("5분 전", AgentSessionFormatting.relative("2026-10-23T11:55:00.123Z", now))
        assertEquals("2시간 전", AgentSessionFormatting.relative("2026-10-23T10:00:00Z", now))
        assertEquals("3일 전", AgentSessionFormatting.relative("2026-10-20T12:00:00Z", now))
        assertEquals("1분 30초", AgentSessionFormatting.duration(90))
        assertEquals("2시간", AgentSessionFormatting.duration(7_200))
        assertEquals("1시간 5분", AgentSessionFormatting.duration(3_900))
        assertEquals("12 KB", AgentSessionFormatting.size(12_000))
        assertEquals("1.5 MB", AgentSessionFormatting.size(1_500_000))
        assertEquals("512 bytes", AgentSessionFormatting.size(512))
        assertNull(AgentSessionFormatting.size(null))
    }

    @Test
    fun stateDetail() {
        val now = Instant.parse("2026-10-23T12:00:00Z")
        val snapshot = AgentSession(
            id = "s1",
            name = "n",
            state = AgentSessionState.SNAPSHOT,
            createdAt = "2026-10-01T00:00:00Z",
            lastAccess = "2026-10-01T00:00:00Z",
            checkpoint = AgentSessionCheckpoint(sizeBytes = 2_000_000, createdAt = "2026-10-23T11:00:00Z"),
        )
        assertEquals("아카이브 1시간 전 · 2 MB · 환경 반납됨", AgentSessionFormatting.stateDetail(snapshot, now))
        val active = snapshot.copy(state = AgentSessionState.ACTIVE, lastAccess = "2026-10-23T11:59:59Z")
        assertEquals("환경이 떠 있어요 · 마지막 사용 방금", AgentSessionFormatting.stateDetail(active, now))
    }

    @Test
    fun terminalCleaning() {
        assertEquals("red ok\nnext", AgentSessionTerminal.clean("\u001B[31mred\u001B[0m \u001B]0;title\u0007ok\r\nnext"))
        assertEquals("ls\n", AgentSessionTerminal.command("ls"))
        assertEquals("ls\n", AgentSessionTerminal.command("ls\n"))
    }

    @Test
    fun kindLabel() {
        val session = AgentSession(id = "1", name = "n", state = AgentSessionState.IDLE, model = "opus")
        val type = AgentSessionWorkloadType(type = "claude-code", label = "Claude Code")
        assertEquals("Claude Code · opus", session.kindLabel(type))
        assertEquals("claude-code", session.copy(model = "").kindLabel(null))
        assertEquals("platform-default", session.copy(model = null).modelLabel)
    }

    @Test
    fun decodesSessionAndUnknownState() {
        val session = ApiClient.json.decodeFromString(
            AgentSession.serializer(),
            """{"id":"a","name":"n","workloadType":"shell","state":"snapshot","createdAt":"c","lastAccess":"l","checkpoint":{"sizeBytes":10}}""",
        )
        assertEquals(AgentSessionState.SNAPSHOT, session.state)
        assertEquals(10L, session.checkpoint?.sizeBytes)
        val unknown = ApiClient.json.decodeFromString(
            AgentSession.serializer(),
            """{"id":"a","name":"n","state":"booting","createdAt":"c","lastAccess":"l"}""",
        )
        assertEquals(AgentSessionState.IDLE, unknown.state)
        assertEquals(AgentSession.DEFAULT_WORKLOAD_TYPE, unknown.workloadType)
    }

    @Test
    fun decodesWorkloadConfig() {
        val config = ApiClient.json.decodeFromString(
            AgentSessionConfigBody.serializer(),
            """{"claudeCode":{"defaultModel":"x","models":[]},"workloadTypes":[
                {"type":"shell","label":"Shell","description":"","takesModel":false,"input":"terminal","defaultModel":null,"models":null},
                {"type":"claude-code","label":"","takesModel":true,"input":"prompt","defaultModel":"opus","models":["sonnet","opus"]},
                {"type":"odd","takesModel":true,"input":"mystery"}
            ]}""",
        ).toConfig()

        assertEquals(listOf("shell", "claude-code", "odd"), config.workloadTypes.map { it.id })
        assertEquals(AgentSessionInput.TERMINAL, config.type("shell")?.input)
        assertNull(config.type("shell")?.model)
        assertEquals("claude-code", config.initialType?.id)
        assertEquals("claude-code", config.type("claude-code")?.displayName)
        assertEquals("opus", config.type("claude-code")?.model?.initialPick)
        assertNull(config.type("odd")?.model?.initialPick)
        assertEquals(AgentSessionInput.PROMPT, config.type("odd")?.input)
    }

    @Test
    fun legacyClaudeCodeConfig() {
        val config = ApiClient.json.decodeFromString(
            AgentSessionConfigBody.serializer(),
            """{"claudeCode":{"defaultModel":"haiku","models":["sonnet"]}}""",
        ).toConfig()

        val type = config.workloadTypes.single()
        assertEquals(AgentSession.DEFAULT_WORKLOAD_TYPE, type.type)
        assertEquals("sonnet", type.model?.initialPick)
    }
}
