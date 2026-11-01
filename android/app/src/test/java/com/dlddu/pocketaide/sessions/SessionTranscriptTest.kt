package com.dlddu.pocketaide.sessions

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.time.Instant
import java.time.ZoneOffset

class SessionTranscriptTest {
    private val s = AgentSessionTranscript.SENTINEL

    @Test
    fun plainTextWithoutSentinelIsOneSegment() {
        val transcript = AgentSessionTranscript.parse("hello\nworld")

        assertEquals(listOf(AgentSessionTranscript.Segment.Text("hello\nworld")), transcript.segments)
        assertFalse(transcript.hasTurns)
    }

    @Test
    fun parsesInputOutputEndTurns() {
        val raw = "boot\n" +
            "${s}session-platform input seq=1 at=2026-10-01T00:00:00Z\nfix the test\n" +
            "${s}session-platform output seq=1\nworking\n" +
            "${s}session-platform output seq=1\ndone\n\n" +
            "${s}session-platform end seq=1 status=ok at=2026-10-01T00:01:05Z\ntrailer"

        val transcript = AgentSessionTranscript.parse(raw)

        assertEquals(3, transcript.segments.size)
        assertEquals(AgentSessionTranscript.Segment.Text("boot\n"), transcript.segments[0])
        val turn = transcript.turns.single()
        assertEquals(1, turn.seq)
        assertEquals("fix the test", turn.input)
        assertEquals("working\ndone", turn.output)
        assertEquals(AgentSessionTurnStatus.Ok, turn.status)
        assertEquals(Instant.parse("2026-10-01T00:00:00Z"), turn.startedAt)
        assertEquals("done", turn.outputPreview)
        assertEquals("fix the test", turn.inputPreview)
        assertEquals(AgentSessionTranscript.Segment.Text("trailer"), transcript.segments[2])
        assertEquals("00:00 · 1분 5초", AgentSessionTurnTiming.label(turn, ZoneOffset.UTC))
    }

    @Test
    fun outputWithoutInputCreatesPlaceholderTurn() {
        val raw = "${s}session-platform output seq=4\npartial\n${s}session-platform end seq=4 status=timeout\n"

        val turn = AgentSessionTranscript.parse(raw).turns.single()

        assertFalse(turn.hasInput)
        assertEquals("partial", turn.output)
        assertEquals(AgentSessionTurnStatus.Timeout, turn.status)
        assertNull(AgentSessionTurnTiming.label(turn))
    }

    @Test
    fun runningTurnAndUnknownMarkers() {
        val raw = "${s}session-platform input seq=2\nhi\n${s}session-platform output seq=2\nchunk\n${s}other header\nmore\n"

        val turn = AgentSessionTranscript.parse(raw).turns.single()

        assertEquals(AgentSessionTurnStatus.Running, turn.status)
        assertEquals("chunk\nmore\n", turn.output)
        assertEquals(2, AgentSessionTranscript.parse(raw).lastSeq)
    }

    @Test
    fun unknownStatusKeepsRawLabel() {
        assertEquals("weird", AgentSessionTurnStatus.fromMarker("weird").label)
        assertTrue(AgentSessionTurnStatus.fromMarker("canceled").isFinished)
    }

    @Test
    fun keepingWholeTurnsDropsPartialHead() {
        val tail = "tail of old\n${AgentSessionTranscript.INPUT_LEAD}seq=3\nnext"

        assertEquals("${AgentSessionTranscript.INPUT_LEAD}seq=3\nnext", AgentSessionTranscript.keepingWholeTurns(tail))
        assertEquals("no marker", AgentSessionTranscript.keepingWholeTurns("no marker"))
    }

    @Test
    fun phaseFromTurns() {
        val done = AgentSessionTranscript.parse("${s}session-platform input seq=1\nx\n${s}session-platform end seq=1 status=ok\n")
        val running = AgentSessionTranscript.parse("${s}session-platform input seq=1\nx\n")

        assertEquals(AgentSessionResponsePhase.STREAMING, AgentSessionResponsePhase.fromTurns(running, null))
        assertEquals(AgentSessionResponsePhase.NONE, AgentSessionResponsePhase.fromTurns(done, null))
        assertEquals(AgentSessionResponsePhase.DONE, AgentSessionResponsePhase.fromTurns(done, 1))
        assertEquals(AgentSessionResponsePhase.QUEUED, AgentSessionResponsePhase.fromTurns(done, 2))
        assertNull(AgentSessionResponsePhase.fromTurns(AgentSessionTranscript(), 1))
    }
}
