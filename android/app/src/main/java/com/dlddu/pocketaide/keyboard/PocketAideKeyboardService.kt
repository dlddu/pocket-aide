package com.dlddu.pocketaide.keyboard

import android.content.Context
import android.graphics.drawable.GradientDrawable
import android.inputmethodservice.InputMethodService
import android.os.Build
import android.util.TypedValue
import android.view.Gravity
import android.view.View
import android.view.inputmethod.InputMethodManager
import android.widget.Button
import android.widget.FrameLayout
import com.dlddu.pocketaide.designsystem.DesignTokens

class PocketAideKeyboardService : InputMethodService() {
    override fun onCreateInputView(): View {
        val style = KeyboardStyle.of(resources)
        val gap = KeyboardStyle.px(resources, DesignTokens.Spacing.lg)
        val root = FrameLayout(this).apply {
            setBackgroundColor(style.surface)
            setPadding(gap, gap, gap, gap)
            minimumHeight = KeyboardStyle.px(resources, DesignTokens.Spacing.xxl) * KEYBOARD_HEIGHT_IN_XXL
        }
        val insert = keyButton(KeyboardCopy.INSERT_TITLE, background = style.accent, text = style.onAccent).apply {
            tag = "keyboard.insert"
            setOnClickListener { currentInputConnection?.commitText(KeyboardCopy.INSERTED_TEXT, 1) }
        }
        val next = keyButton(KeyboardCopy.NEXT_KEYBOARD, background = style.key, text = style.ink).apply {
            tag = "keyboard.next"
            setOnClickListener { switchToNextKeyboard() }
        }
        root.addView(insert, FrameLayout.LayoutParams(wrap, wrap, Gravity.TOP or Gravity.CENTER_HORIZONTAL))
        root.addView(next, FrameLayout.LayoutParams(wrap, wrap, Gravity.BOTTOM or Gravity.START))
        return root
    }

    private fun keyButton(title: String, background: Int, text: Int): Button = Button(this).apply {
        this.text = title
        isAllCaps = false
        setTextColor(text)
        setTextSize(TypedValue.COMPLEX_UNIT_SP, KeyboardStyle.textSizeSp)
        val horizontal = KeyboardStyle.px(resources, DesignTokens.Spacing.lg)
        val vertical = KeyboardStyle.px(resources, DesignTokens.Spacing.sm)
        setPadding(horizontal, vertical, horizontal, vertical)
        stateListAnimator = null
        this.background = GradientDrawable().apply {
            setColor(background)
            cornerRadius = KeyboardStyle.px(resources, DesignTokens.Radius.key).toFloat()
        }
    }

    private fun switchToNextKeyboard() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            switchToNextInputMethod(false)
        } else {
            val token = window?.window?.attributes?.token ?: return
            @Suppress("DEPRECATION")
            (getSystemService(Context.INPUT_METHOD_SERVICE) as InputMethodManager).switchToNextInputMethod(token, false)
        }
    }

    private companion object {
        const val KEYBOARD_HEIGHT_IN_XXL = 8
        val wrap = FrameLayout.LayoutParams.WRAP_CONTENT
    }
}
