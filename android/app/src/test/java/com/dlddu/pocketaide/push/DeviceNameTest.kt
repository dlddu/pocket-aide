package com.dlddu.pocketaide.push

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class DeviceNameTest {
    @Test
    fun normalizeTrimsAndCapsLength() {
        assertEquals("Pixel 8", DeviceName.normalize("  Pixel 8  "))
        assertEquals("가".repeat(DeviceName.MAX_LENGTH), DeviceName.normalize("가".repeat(DeviceName.MAX_LENGTH + 5)))
        assertEquals("", DeviceName.normalize("   "))
    }

    @Test
    fun registrationOmitsNameUntilOneWasSaved() {
        assertNull(DeviceName.forRegistration(null))
        assertNull(DeviceName.forRegistration(""))
        assertEquals("출근용 폰", DeviceName.forRegistration("출근용 폰"))
    }

    @Test
    fun inMemoryStoreDistinguishesSkippedFromNotAsked() {
        val store = InMemoryDeviceNameStore()
        assertNull(store.load())
        store.save("")
        assertEquals("", store.load())
    }
}
