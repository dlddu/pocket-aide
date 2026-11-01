package com.dlddu.pocketaide.keyboard

import org.junit.Assert.assertEquals
import org.junit.Test

class KeyboardCopyTest {
    @Test
    fun matchesTheIosKeyboardExtension() {
        assertEquals("Insert hello from pocket-aide", KeyboardCopy.INSERT_TITLE)
        assertEquals("Hello from pocket-aide", KeyboardCopy.INSERTED_TEXT)
        assertEquals("next ⇄", KeyboardCopy.NEXT_KEYBOARD)
    }
}
