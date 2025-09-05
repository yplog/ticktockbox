package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/yplog/ticktockbox/internal/db"
	"github.com/yplog/ticktockbox/internal/jobs"
)

func main() {
	dbPath := getenv("SQLITE_PATH", "app.db")
	count := getenvInt("SEED_COUNT", 1000)

	sqlDB, err := db.Open(dbPath)
	must(err)
	defer sqlDB.Close()

	ctx := context.Background()
	must(db.Migrate(ctx, sqlDB))
	repo := &jobs.Repo{DB: sqlDB}

	clearExisting := getenv("CLEAR_EXISTING", "true")
	if clearExisting == "true" {
		log.Printf("Clearing existing jobs...")
		must(clearExistingJobs(ctx, repo))
	}

	log.Printf("Generating %d test jobs...", count)
	now := time.Now()

	timezones := []string{
		"Europe/Istanbul", "America/New_York", "America/Los_Angeles",
		"Europe/London", "Asia/Tokyo", "Australia/Sydney", "UTC",
		"Europe/Paris", "America/Chicago", "Asia/Shanghai",
	}

	titles := []string{
		"Team Meeting", "Client Call", "Project Deadline", "Code Review",
		"Doctor Appointment", "Birthday Party", "Conference Call", "Lunch Meeting",
		"Performance Review", "Training Session", "Product Launch", "Budget Meeting",
		"System Maintenance", "Data Backup", "Server Update", "Security Audit",
		"Customer Presentation", "Sales Review", "Marketing Campaign", "Website Deploy",
		"Database Migration", "API Release", "Bug Fix Review", "Sprint Planning",
		"Daily Standup", "Weekly Retrospective", "Monthly Report", "Quarterly Review",
	}

	for i := range count {
		randomDays := rand.Intn(30)
		randomHours := rand.Intn(24)
		randomMinutes := rand.Intn(60)

		runAt := now.Add(time.Duration(randomDays)*24*time.Hour +
			time.Duration(randomHours)*time.Hour +
			time.Duration(randomMinutes)*time.Minute)

		tz := timezones[rand.Intn(len(timezones))]

		loc, err := time.LoadLocation(tz)
		must(err)
		runAtLocal := runAt.In(loc)
		runAtUTC := runAtLocal.UTC()

		remindBefore := rand.Intn(60) + 1
		dueAtUTC := runAtUTC.Add(-time.Duration(remindBefore) * time.Minute)

		title := fmt.Sprintf("%s #%d", titles[rand.Intn(len(titles))], i+1)

		job := jobs.Job{
			Title:               title,
			TZ:                  tz,
			RunAtUTC:            runAtUTC,
			DueAtUTC:            dueAtUTC,
			RemindBeforeMinutes: remindBefore,
		}

		if i < count/10 {
			pastDays := rand.Intn(7) + 1
			pastTime := now.Add(-time.Duration(pastDays) * 24 * time.Hour)
			job.RunAtUTC = pastTime.UTC()
			job.DueAtUTC = pastTime.Add(-time.Duration(remindBefore) * time.Minute)
			job.Title = fmt.Sprintf("[PAST] %s", job.Title)
		}

		if i >= count/10 && i < count/10+count/20 {
			soonMinutes := rand.Intn(60) + 1
			soonTime := now.Add(time.Duration(soonMinutes) * time.Minute)
			job.RunAtUTC = soonTime.UTC()
			job.DueAtUTC = soonTime.Add(-time.Duration(remindBefore) * time.Minute)
			job.Title = fmt.Sprintf("[SOON] %s", job.Title)
		}

		_, err = repo.Insert(ctx, &job)
		must(err)

		if (i+1)%100 == 0 {
			log.Printf("Generated %d/%d jobs...", i+1, count)
		}
	}

	log.Printf("Successfully generated %d test jobs!", count)
	log.Printf("Updating past jobs status...")

	result, err := repo.DB.ExecContext(ctx,
		"UPDATE jobs SET status = 'completed' WHERE run_at_utc < ?",
		now)
	must(err)

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Updated %d past jobs to completed status", rowsAffected)

	stats, err := getStats(ctx, repo)
	must(err)

	log.Printf("Database Statistics:")
	log.Printf("   Total jobs: %d", stats.Total)
	log.Printf("   Pending: %d", stats.Pending)
	log.Printf("   Past jobs: %d", stats.Past)
	log.Printf("   Jobs due in next hour: %d", stats.DueSoon)
	log.Printf("   Jobs due in next 24h: %d", stats.DueToday)
}

type Stats struct {
	Total    int
	Pending  int
	Past     int
	DueSoon  int
	DueToday int
}

func clearExistingJobs(ctx context.Context, repo *jobs.Repo) error {
	_, err := repo.DB.ExecContext(ctx, "DELETE FROM jobs")
	return err
}

func getStats(ctx context.Context, repo *jobs.Repo) (*Stats, error) {
	now := time.Now()

	var stats Stats

	err := repo.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM jobs").Scan(&stats.Total)
	if err != nil {
		return nil, err
	}

	err = repo.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM jobs WHERE status = 'pending'").Scan(&stats.Pending)
	if err != nil {
		return nil, err
	}

	err = repo.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM jobs WHERE run_at_utc < ?", now).Scan(&stats.Past)
	if err != nil {
		return nil, err
	}

	err = repo.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM jobs WHERE due_at_utc BETWEEN ? AND ?",
		now, now.Add(time.Hour)).Scan(&stats.DueSoon)
	if err != nil {
		return nil, err
	}

	err = repo.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM jobs WHERE due_at_utc BETWEEN ? AND ?",
		now, now.Add(24*time.Hour)).Scan(&stats.DueToday)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getenvInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			return val
		}
	}
	return d
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
