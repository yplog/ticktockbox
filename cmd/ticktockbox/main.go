package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/handler"
	"github.com/yplog/ticktockbox/internal/notifier"
)

type App struct {
	cfg      *config.Config
	db       *database.Database
	notifier *notifier.Notifier
	handler  *handler.Handler
	server   *http.Server
}

func NewApp() *App {
	cfg := config.Load()
	if cfg == nil {
		log.Fatalf("Failed to load configuration")
	}

	cfg.Log()

	db, err := database.InitDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	n := notifier.NewNotifier(cfg.Notifier)
	h := handler.NewHandler(cfg, db, n)

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: http.DefaultServeMux,
	}

	return &App{
		cfg:      cfg,
		db:       db,
		notifier: n,
		handler:  h,
		server:   server,
	}
}

func (app *App) Run() {
	http.HandleFunc("/", app.handler.Healthcheck)
	http.HandleFunc("/create", app.handler.CreateHandler)
	if app.cfg.Notifier.UseWebSocket {
		http.HandleFunc("/ws", app.handler.WebSocketHandler)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(app.cfg.Database.CheckerInterval) * time.Second)
		for range ticker.C {
			app.handler.GetExpireRecordsHandler()
		}
	}()

	go func() {
		log.Printf("Server starting on %s", app.server.Addr)

		if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if err := app.server.Close(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

func main() {
	app := NewApp()
	app.Run()
}
