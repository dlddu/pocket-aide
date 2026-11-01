package com.dlddu.pocketaide.affirmations

import kotlin.random.Random

fun interface RandomSource {
    fun next(): ULong

    fun nextBelow(upperBound: ULong): ULong {
        require(upperBound != 0uL)
        var product = multiplyFull(next(), upperBound)
        if (product.second < upperBound) {
            val threshold = (0uL - upperBound) % upperBound
            while (product.second < threshold) {
                product = multiplyFull(next(), upperBound)
            }
        }
        return product.first
    }

    companion object {
        val system: RandomSource = RandomSource { Random.Default.nextLong().toULong() }

        internal fun multiplyFull(a: ULong, b: ULong): Pair<ULong, ULong> {
            val mask = 0xFFFF_FFFFuL
            val aLow = a and mask
            val aHigh = a shr 32
            val bLow = b and mask
            val bHigh = b shr 32
            val lowLow = aLow * bLow
            val lowHigh = aLow * bHigh
            val highLow = aHigh * bLow
            val highHigh = aHigh * bHigh
            val cross = (lowLow shr 32) + (lowHigh and mask) + highLow
            val high = highHigh + (lowHigh shr 32) + (cross shr 32)
            return high to a * b
        }
    }
}

class SeededRng(seed: ULong) : RandomSource {
    private var state: ULong = if (seed == 0uL) ZERO_SEED_REPLACEMENT else seed

    override fun next(): ULong {
        state += GOLDEN_GAMMA
        var z = state
        z = (z xor (z shr 30)) * MIX_1
        z = (z xor (z shr 27)) * MIX_2
        return z xor (z shr 31)
    }

    private companion object {
        const val ZERO_SEED_REPLACEMENT = 0xDEAD_BEEF_CAFE_F00DuL
        const val GOLDEN_GAMMA = 0x9E37_79B9_7F4A_7C15uL
        const val MIX_1 = 0xBF58_476D_1CE4_E5B9uL
        const val MIX_2 = 0x94D0_49BB_1331_11EBuL
    }
}

class RotationSelector(
    private val weights: Map<AffirmationPriority, Int> = DEFAULT_WEIGHTS,
) {
    fun pick(pool: List<Affirmation>, random: RandomSource): Affirmation? {
        if (pool.isEmpty()) return null
        val total = pool.sumOf { weight(it.priority) }
        if (total <= 0) return pool.first()
        val target = random.nextBelow(total.toULong()).toLong()
        var running = 0L
        for (item in pool) {
            running += weight(item.priority)
            if (target < running) return item
        }
        return pool.last()
    }

    private fun weight(priority: AffirmationPriority): Int = maxOf(weights[priority] ?: 0, 0)

    companion object {
        val DEFAULT_WEIGHTS: Map<AffirmationPriority, Int> = mapOf(
            AffirmationPriority.HIGH to 3,
            AffirmationPriority.NORMAL to 2,
            AffirmationPriority.LOW to 1,
        )
    }
}
