package routines

import "testing"

func TestScheduleMatches(t *testing.T) {
	sunday := WeekdayBit(0)
	monWedFri := WeekdayBit(1) | WeekdayBit(3) | WeekdayBit(5)
	cases := []struct {
		name     string
		schedule Schedule
		day      string
		want     bool
	}{
		{"daily", Schedule{Cadence: CadenceDaily}, "2026-10-01", true},
		{"weekdays hit", Schedule{Cadence: CadenceWeekdays, Weekdays: monWedFri}, "2026-09-30", true},
		{"weekdays miss", Schedule{Cadence: CadenceWeekdays, Weekdays: monWedFri}, "2026-10-01", false},
		{"weekly sunday", Schedule{Cadence: CadenceWeekly, Weekdays: sunday}, "2026-10-04", true},
		{"weekly not sunday", Schedule{Cadence: CadenceWeekly, Weekdays: sunday}, "2026-10-05", false},
		{"monthly", Schedule{Cadence: CadenceMonthly, MonthDay: 15}, "2026-10-15", true},
		{"monthly other day", Schedule{Cadence: CadenceMonthly, MonthDay: 15}, "2026-10-16", false},
		{"monthly 31 in short month", Schedule{Cadence: CadenceMonthly, MonthDay: 31}, "2026-09-30", true},
		{"monthly 31 in february", Schedule{Cadence: CadenceMonthly, MonthDay: 31}, "2026-02-28", true},
	}
	for _, c := range cases {
		d, err := ParseDay(c.day)
		if err != nil {
			t.Fatalf("%s: parse %s: %v", c.name, c.day, err)
		}
		if got := c.schedule.Matches(d); got != c.want {
			t.Errorf("%s: Matches(%s) = %v, want %v", c.name, c.day, got, c.want)
		}
	}
}

func TestScheduleNormalized(t *testing.T) {
	invalid := []Schedule{
		{Cadence: "hourly"},
		{Cadence: CadenceWeekdays},
		{Cadence: CadenceWeekdays, Weekdays: 128},
		{Cadence: CadenceWeekly, Weekdays: WeekdayBit(0) | WeekdayBit(1)},
		{Cadence: CadenceMonthly},
		{Cadence: CadenceMonthly, MonthDay: 32},
	}
	for _, s := range invalid {
		if _, err := s.Normalized(); err != ErrInvalidSchedule {
			t.Errorf("Normalized(%+v) = %v, want ErrInvalidSchedule", s, err)
		}
	}
	got, err := Schedule{Cadence: CadenceDaily, Weekdays: 5, MonthDay: 3}.Normalized()
	if err != nil || got != (Schedule{Cadence: CadenceDaily}) {
		t.Errorf("daily must drop weekdays and month_day: got %+v, %v", got, err)
	}
}

func TestScheduledOnRespectsStartDay(t *testing.T) {
	r := Routine{StartDay: "2026-10-01", Schedule: Schedule{Cadence: CadenceDaily}}
	before, _ := ParseDay("2026-09-30")
	on, _ := ParseDay("2026-10-01")
	if r.ScheduledOn(before) || !r.ScheduledOn(on) {
		t.Errorf("start day: before=%v on=%v", r.ScheduledOn(before), r.ScheduledOn(on))
	}
}

func TestParseDayRejectsLooseFormats(t *testing.T) {
	for _, s := range []string{"2026-1-01", "2026/10/01", "2026-02-30", "today", ""} {
		if _, err := ParseDay(s); err != ErrInvalidDay {
			t.Errorf("ParseDay(%q) = %v, want ErrInvalidDay", s, err)
		}
	}
}
