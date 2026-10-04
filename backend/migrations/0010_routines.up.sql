CREATE TABLE IF NOT EXISTS routines (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    cadence TEXT NOT NULL DEFAULT 'daily' CHECK(cadence IN ('daily', 'weekdays', 'weekly', 'monthly')),
    weekdays INTEGER NOT NULL DEFAULT 0 CHECK(weekdays BETWEEN 0 AND 127),
    month_day INTEGER NOT NULL DEFAULT 0 CHECK(month_day BETWEEN 0 AND 31),
    start_day TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_routines_user_id ON routines(user_id);

CREATE TABLE IF NOT EXISTS routine_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    routine_id INTEGER NOT NULL REFERENCES routines(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    title TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
);

CREATE INDEX IF NOT EXISTS idx_routine_steps_routine_id ON routine_steps(routine_id);

CREATE TABLE IF NOT EXISTS routine_step_checks (
    step_id INTEGER NOT NULL REFERENCES routine_steps(id) ON DELETE CASCADE,
    day TEXT NOT NULL,
    checked_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (step_id, day)
);
