package com.dlddu.pocketaide.widget

import android.Manifest
import android.annotation.SuppressLint
import android.content.ContentUris
import android.content.Context
import android.content.pm.PackageManager
import android.provider.CalendarContract
import java.time.Instant
import java.time.ZoneId

data class CalendarSnapshot(val authorized: Boolean, val events: List<UpcomingEvent>) {
    fun state(at: Instant): WidgetCalendarState =
        if (!authorized) {
            WidgetCalendarState.NeedsPermission
        } else {
            WidgetCalendarState.Loaded(UpcomingEvents.summarize(events, at))
        }

    companion object {
        const val PERMISSION = Manifest.permission.READ_CALENDAR

        private val PROJECTION = arrayOf(
            CalendarContract.Instances.TITLE,
            CalendarContract.Instances.BEGIN,
            CalendarContract.Instances.END,
            CalendarContract.Instances.ALL_DAY,
        )
        private val SELECTION = "${CalendarContract.Instances.VISIBLE} = 1"
        private val SORT = "${CalendarContract.Instances.BEGIN} ASC"

        fun hasPermission(context: Context): Boolean =
            context.checkSelfPermission(PERMISSION) == PackageManager.PERMISSION_GRANTED

        @SuppressLint("MissingPermission")
        fun load(context: Context, from: Instant, zone: ZoneId = ZoneId.systemDefault()): CalendarSnapshot {
            if (!hasPermission(context)) return CalendarSnapshot(authorized = false, events = emptyList())
            val uri = CalendarContract.Instances.CONTENT_URI.buildUpon().also {
                ContentUris.appendId(it, from.toEpochMilli())
                ContentUris.appendId(it, from.plus(UpcomingEvents.LOOKAHEAD).toEpochMilli())
            }.build()
            val events = mutableListOf<UpcomingEvent>()
            try {
                context.contentResolver.query(uri, PROJECTION, SELECTION, null, SORT)?.use { cursor ->
                    while (cursor.moveToNext()) {
                        val allDay = cursor.getInt(COLUMN_ALL_DAY) != 0
                        val begin = cursor.getLong(COLUMN_BEGIN)
                        val end = cursor.getLong(COLUMN_END)
                        events += UpcomingEvent(
                            title = cursor.getString(COLUMN_TITLE).orEmpty(),
                            start = if (allDay) UpcomingEvents.allDayInstant(begin, zone) else Instant.ofEpochMilli(begin),
                            end = if (allDay) UpcomingEvents.allDayInstant(end, zone) else Instant.ofEpochMilli(end),
                            isAllDay = allDay,
                        )
                    }
                }
            } catch (e: SecurityException) {
                return CalendarSnapshot(authorized = false, events = emptyList())
            }
            return CalendarSnapshot(authorized = true, events = events)
        }

        private const val COLUMN_TITLE = 0
        private const val COLUMN_BEGIN = 1
        private const val COLUMN_END = 2
        private const val COLUMN_ALL_DAY = 3
    }
}
