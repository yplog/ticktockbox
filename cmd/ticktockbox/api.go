package main

import (
	"errors"
	"fmt"
	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/handler"
	"github.com/yplog/ticktockbox/internal/notifier"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type API struct {
	cfg      *config.Config
	db       *database.Database
	notifier *notifier.Notifier
	handler  *handler.Handler
	server   *http.Server
}

func NewAPI() *API {
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

	return &API{
		cfg:      cfg,
		db:       db,
		notifier: n,
		handler:  h,
		server:   server,
	}
}

func (api *API) Run() {
	http.HandleFunc("/", api.handler.Healthcheck)
	http.HandleFunc("/create", api.handler.CreateHandler)
	if api.cfg.Notifier.UseWebSocket {
		http.HandleFunc("/ws", api.handler.WebSocketHandler)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(api.cfg.Database.CheckerInterval) * time.Second)
		for range ticker.C {
			api.handler.GetExpireRecordsHandler()
		}
	}()

	go func() {
		log.Printf("Server starting on %s", api.server.Addr)

		if err := api.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if err := api.server.Close(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
