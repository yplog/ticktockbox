package main

import (
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

	log.Printf("Server port: %d", cfg.Server.Port)

	db, err := database.NewDatabase(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	notifier := notifier.NewNotifier(cfg.Notifier)

	h := handler.NewHandler(cfg, db, notifier)
	http.HandleFunc("/", h.Healthcheck)
	http.HandleFunc("/add", h.AddItemHandler)
	http.HandleFunc("/get", h.GetItemHandler)
	if cfg.Notifier.UseWebSocket {
		http.HandleFunc("/ws", h.WebSocketHandler)
	}

	go func() {
		// TODO: Interval should be configurable
		ticker := time.NewTicker(time.Duration(1) * time.Second)
		for range ticker.C {
			h.CheckExpiredItems()
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for range ticker.C {
			if err := db.RunGC(); err != nil {
				log.Printf("Error running database GC: %v", err)
			}
		}
	}()

	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Printf("Server starting on %s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
