package jobs

import (
	"context"
	"database/sql"
	"time"
)

type Job struct {
	ID                  int64
	Title               string
	TZ                  string
	RunAtUTC            time.Time
	DueAtUTC            time.Time
	RemindBeforeMinutes int
	Status              string
	CreatedAt           time.Time
}

type JobFilter struct {
	Status string
	Page   int
	Limit  int
}

type JobPage struct {
	Jobs       []Job
	Total      int
	Page       int
	Limit      int
	TotalPages int
	HasNext    bool
	HasPrev    bool
}

type Repo struct{ DB *sql.DB }

func (r *Repo) Insert(ctx context.Context, j *Job) (int64, error) {
	res, err := r.DB.ExecContext(ctx, `
	  INSERT INTO jobs(title, tz, run_at_utc, due_at_utc, remind_before_minutes, status)
	  VALUES (?, ?, ?, ?, ?, 'pending')`,
		j.Title, j.TZ, j.RunAtUTC, j.DueAtUTC, j.RemindBeforeMinutes)

	if err != nil {
		return 0, err
	}

	id, _ := res.LastInsertId()

	return id, nil
}

func (r *Repo) MarkEnqueued(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE jobs SET status='enqueued' WHERE id=?`, id)

	return err
}

func (r *Repo) MarkDone(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE jobs SET status='done' WHERE id=?`, id)

	return err
}

func (r *Repo) Cancel(ctx context.Context, id int64) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE jobs SET status='cancelled' WHERE id=?`, id)

	return err
}

func (r *Repo) GetUpcoming(ctx context.Context, limit int) ([]Job, error) {
	rows, err := r.DB.QueryContext(ctx, `
	  SELECT id, title, tz, run_at_utc, due_at_utc, remind_before_minutes, status, created_at
	  FROM jobs
	  WHERE status IN ('pending','enqueued')
	  ORDER BY due_at_utc ASC
	  LIMIT ?`, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []Job

	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Title, &j.TZ, &j.RunAtUTC, &j.DueAtUTC, &j.RemindBeforeMinutes, &j.Status, &j.CreatedAt); err != nil {
			return nil, err
		}

		res = append(res, j)
	}

	return res, rows.Err()
}

func (r *Repo) GetJobsPaginated(ctx context.Context, filter JobFilter) (*JobPage, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 50
	}

	offset := (filter.Page - 1) * filter.Limit

	var statusCondition string
	var args []any

	if filter.Status == "" || filter.Status == "all" {
		statusCondition = "1=1"
	} else {
		statusCondition = "status = ?"
		args = append(args, filter.Status)
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM jobs WHERE ` + statusCondition
	err := r.DB.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT id, title, tz, run_at_utc, due_at_utc, remind_before_minutes, status, created_at
		FROM jobs
		WHERE ` + statusCondition + `
		ORDER BY due_at_utc ASC
		LIMIT ? OFFSET ?`

	args = append(args, filter.Limit, offset)
	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Title, &j.TZ, &j.RunAtUTC, &j.DueAtUTC, &j.RemindBeforeMinutes, &j.Status, &j.CreatedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit

	return &JobPage{
		Jobs:       jobs,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
		HasNext:    filter.Page < totalPages,
		HasPrev:    filter.Page > 1,
	}, nil
}

func (r *Repo) LoadPendingSince(ctx context.Context, since time.Time) ([]Job, error) {
	rows, err := r.DB.QueryContext(ctx, `
	  SELECT id, title, tz, run_at_utc, due_at_utc, remind_before_minutes, status, created_at
	  FROM jobs
	  WHERE status = 'pending' AND due_at_utc >= ?
	  ORDER BY due_at_utc ASC`, since)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var res []Job

	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.Title, &j.TZ, &j.RunAtUTC, &j.DueAtUTC, &j.RemindBeforeMinutes, &j.Status, &j.CreatedAt); err != nil {
			return nil, err
		}

		res = append(res, j)
	}

	return res, rows.Err()
}
