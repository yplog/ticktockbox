package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/yplog/ticktockbox/internal/db"
	httpx "github.com/yplog/ticktockbox/internal/http"
	"github.com/yplog/ticktockbox/internal/jobs"
	"github.com/yplog/ticktockbox/internal/rmq"
	"github.com/yplog/ticktockbox/internal/twheel"
	"github.com/yplog/ticktockbox/public"
	"github.com/yplog/ticktockbox/templates"

	// Embed IANA timezone database to support all time zones
	_ "time/tzdata"
)

func main() {
	// ENV
	addr := getenv("ADDR", ":3000")
	dbPath := getenv("SQLITE_PATH", "app.db")
	rmqURL := getenv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	queue := getenv("RABBITMQ_QUEUE", "reminders.due")

	// DB
	sqlDB, err := db.Open(dbPath)
	must(err)
	ctx := context.Background()
	must(db.Migrate(ctx, sqlDB))
	repo := &jobs.Repo{DB: sqlDB}

	// RabbitMQ
	pub, err := rmq.NewPublisher(rmqURL, queue)
	must(err)
	defer pub.Close()

	// Wheel
	wheel := twheel.New(1*time.Second, 512)
	wheel.Start()
	defer wheel.Stop(context.Background())

	// Scheduler
	sched := jobs.NewScheduler(repo, pub, wheel)
	must(sched.Warmup(ctx))

	// HTTP
	admin := &httpx.AdminHandlers{
		Repo:        repo,
		Scheduler:   sched,
		TemplatesFS: templates.TemplateFiles,
		Assets:      public.PublicFiles,
		Validate:    validator.New(validator.WithRequiredStructEnabled()),
	}
	srv := httpx.NewServer(admin)

	log.Printf("listening on %s", addr)
	must(http.ListenAndServe(addr, srv.R))
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}

	return d
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
