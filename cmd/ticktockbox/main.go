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

func main() {
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

	http.HandleFunc("/", h.Healthcheck)
	http.HandleFunc("/create", h.CreateHandler)
	if cfg.Notifier.UseWebSocket {
		http.HandleFunc("/ws", h.WebSocketHandler)
	}

	go func() {
		ticker := time.NewTicker(time.Duration(cfg.Database.Checker) * time.Second)
		for range ticker.C {
			h.GetExpireRecordsHandler()
		}
	}()

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Printf("Server starting on %s", serverAddr)

		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if err := server.Close(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
