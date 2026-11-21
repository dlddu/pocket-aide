package com.dlddu.pocketaide.chat

import org.junit.Assert.assertEquals
import org.junit.Test

class ChatLinksTest {
    private fun links(text: String) = chatLinkRanges(text).map { text.substring(it) }

    @Test
    fun findsTheCodexSignInAddress() {
        val text = "1. Open this link in your browser and sign in to your account\n" +
            "   https://auth.openai.com/codex/device\n2. Enter this one-time code (expires in 15 minutes)\n   2V6W-MKPWJ"
        assertEquals(listOf("https://auth.openai.com/codex/device"), links(text))
    }

    @Test
    fun leavesSentencePunctuationOut() {
        assertEquals(
            listOf("https://example.com/a?b=1", "http://x.io/path"),
            links("see https://example.com/a?b=1. and (http://x.io/path), done"),
        )
    }

    @Test
    fun stopsAtAnUnopenedClosingBracketBeforeAParticle() {
        assertEquals(
            listOf("https://github.com/dlddu/plugin-marketplace/pull/95"),
            links("[PR #95](https://github.com/dlddu/plugin-marketplace/pull/95)를 머지했고"),
        )
        assertEquals(listOf("https://x.io/pull/96"), links("(https://x.io/pull/96)을 올렸습니다."))
    }

    @Test
    fun keepsBracketsTheAddressOpened() {
        assertEquals(
            listOf("https://en.wikipedia.org/wiki/Foo_(bar)"),
            links("see https://en.wikipedia.org/wiki/Foo_(bar)."),
        )
    }

    @Test
    fun ignoresTextWithoutAddresses() {
        assertEquals(emptyList<String>(), links("no links here, just https:// and words"))
    }
}
