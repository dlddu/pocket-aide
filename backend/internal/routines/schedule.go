package routines

import (
	"errors"
	"math/bits"
	"time"
)

type Cadence string

const (
	CadenceDaily    Cadence = "daily"
	CadenceWeekdays Cadence = "weekdays"
	CadenceWeekly   Cadence = "weekly"
	CadenceMonthly  Cadence = "monthly"
)

const dayLayout = "2006-01-02"

const allWeekdays = 1<<7 - 1

var (
	ErrInvalidDay      = errors.New("day must be YYYY-MM-DD")
	ErrInvalidSchedule = errors.New("invalid schedule")
)

type Schedule struct {
	Cadence  Cadence `json:"cadence"`
	Weekdays int     `json:"weekdays"`
	MonthDay int     `json:"month_day"`
}

func ParseDay(s string) (time.Time, error) {
	t, err := time.Parse(dayLayout, s)
	if err != nil || t.Format(dayLayout) != s {
		return time.Time{}, ErrInvalidDay
	}
	return t, nil
}

func WeekdayBit(d time.Weekday) int { return 1 << int(d) }

func (s Schedule) Normalized() (Schedule, error) {
	switch s.Cadence {
	case CadenceDaily:
		return Schedule{Cadence: CadenceDaily}, nil
	case CadenceWeekdays:
		if s.Weekdays <= 0 || s.Weekdays > allWeekdays {
			return Schedule{}, ErrInvalidSchedule
		}
		return Schedule{Cadence: CadenceWeekdays, Weekdays: s.Weekdays}, nil
	case CadenceWeekly:
		if s.Weekdays <= 0 || s.Weekdays > allWeekdays || bits.OnesCount(uint(s.Weekdays)) != 1 {
			return Schedule{}, ErrInvalidSchedule
		}
		return Schedule{Cadence: CadenceWeekly, Weekdays: s.Weekdays}, nil
	case CadenceMonthly:
		if s.MonthDay < 1 || s.MonthDay > 31 {
			return Schedule{}, ErrInvalidSchedule
		}
		return Schedule{Cadence: CadenceMonthly, MonthDay: s.MonthDay}, nil
	default:
		return Schedule{}, ErrInvalidSchedule
	}
}

func (s Schedule) Matches(day time.Time) bool {
	switch s.Cadence {
	case CadenceDaily:
		return true
	case CadenceWeekdays, CadenceWeekly:
		return s.Weekdays&WeekdayBit(day.Weekday()) != 0
	case CadenceMonthly:
		last := time.Date(day.Year(), day.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
		target := s.MonthDay
		if target > last {
			target = last
		}
		return day.Day() == target
	default:
		return false
	}
}
