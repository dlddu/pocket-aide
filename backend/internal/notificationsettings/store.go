package notificationsettings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Outcomes string

const (
	OutcomesBoth    Outcomes = "both"
	OutcomesSuccess Outcomes = "success"
	OutcomesFailure Outcomes = "failure"
)

var ErrInvalidOutcomes = errors.New("outcomes must be one of both, success, failure")

// Settings is one row of the user_notification_settings table; a user without a row gets Default().
type Settings struct {
	Enabled  bool     `json:"enabled"`
	Outcomes Outcomes `json:"outcomes"`
}

type Patch struct {
	Enabled  *bool     `json:"enabled"`
	Outcomes *Outcomes `json:"outcomes"`
}

func Default() Settings {
	return Settings{Enabled: true, Outcomes: OutcomesBoth}
}

func (o Outcomes) valid() bool {
	switch o {
	case OutcomesBoth, OutcomesSuccess, OutcomesFailure:
		return true
	}
	return false
}

func IsFailureConclusion(conclusion string) bool {
	switch conclusion {
	case "failure", "timed_out", "startup_failure":
		return true
	}
	return false
}

func (s Settings) AllowsPush(conclusion string) bool {
	if !s.Enabled {
		return false
	}
	switch s.Outcomes {
	case OutcomesSuccess:
		return conclusion == "success"
	case OutcomesFailure:
		return IsFailureConclusion(conclusion)
	}
	return true
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store { return &Store{db: db} }

func (s *Store) Get(ctx context.Context, userID int64) (Settings, error) {
	out := Default()
	var enabled int
	var outcomes string
	err := s.db.QueryRowContext(ctx, `
		SELECT enabled, outcomes
		FROM user_notification_settings
		WHERE user_id = ?
	`, userID).Scan(&enabled, &outcomes)
	if errors.Is(err, sql.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("get notification settings: %w", err)
	}
	out.Enabled = enabled == 1
	out.Outcomes = Outcomes(outcomes)
	return out, nil
}

func (s *Store) Update(ctx context.Context, userID int64, p Patch) (Settings, error) {
	if p.Outcomes != nil && !p.Outcomes.valid() {
		return Settings{}, ErrInvalidOutcomes
	}
	current, err := s.Get(ctx, userID)
	if err != nil {
		return Settings{}, err
	}
	if p.Enabled != nil {
		current.Enabled = *p.Enabled
	}
	if p.Outcomes != nil {
		current.Outcomes = *p.Outcomes
	}
	enabled := 0
	if current.Enabled {
		enabled = 1
	}
	if _, err := s.db.ExecContext(ctx, `
		INSERT INTO user_notification_settings (user_id, enabled, outcomes)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id) DO UPDATE SET
			enabled = excluded.enabled,
			outcomes = excluded.outcomes,
			updated_at = unixepoch()
	`, userID, enabled, string(current.Outcomes)); err != nil {
		return Settings{}, fmt.Errorf("upsert notification settings: %w", err)
	}
	return current, nil
}
