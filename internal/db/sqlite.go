package db

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	// Busy timeout, WAL ve foreign_keys
	dsn := "file:" + path + "?cache=shared&mode=rwc&_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=foreign_keys=ON"
	return sql.Open("sqlite", dsn)
}

func Migrate(ctx context.Context, sqlDB *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS jobs(
		   id INTEGER PRIMARY KEY AUTOINCREMENT,
		   title TEXT NOT NULL,
		   tz TEXT NOT NULL,
		   run_at_utc TIMESTAMP NOT NULL,
		   due_at_utc TIMESTAMP NOT NULL, -- run_at - remind_before
		   remind_before_minutes INTEGER NOT NULL DEFAULT 0,
           status TEXT NOT NULL DEFAULT 'pending', -- pending|enqueued|cancelled
		   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		 );`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_status_due ON jobs(status, due_at_utc);`,
	}

	for _, s := range stmts {
		if _, err := sqlDB.ExecContext(ctx, s); err != nil {
			return err
		}
	}

	return nil
}
