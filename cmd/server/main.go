package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"github.com/yplog/ticktockbox/internal/api"
	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/rabbitmq"
	"github.com/yplog/ticktockbox/internal/scheduler"
	"github.com/yplog/ticktockbox/internal/websocket"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	db, err := database.NewQuestDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to QuestDB:", err)
	}
	defer db.Close()

	rmq, err := rabbitmq.New(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer rmq.Close()

	wsHub := websocket.NewHub()
	go wsHub.Run()

	sched := scheduler.New(db, rmq, wsHub)
	go sched.Start(10 * time.Second)

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	apiHandler := api.NewHandler(db, rmq)
	r.Route("/api", func(r chi.Router) {
		r.Post("/messages", apiHandler.CreateMessage)
		r.Get("/messages", apiHandler.GetMessages)
	})

	r.Get("/ws", wsHub.HandleWebSocket)

	log.Printf("Server starting on port %s", cfg.Port)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")
}
