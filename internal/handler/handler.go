package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/model"
	"github.com/yplog/ticktockbox/internal/notifier"
)

type Handler struct {
	cfg      *config.Config
	db       *database.Database
	notifier *notifier.Notifier
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Replace with actual origin check
		return true
	},
}

func NewHandler(cfg *config.Config, db *database.Database, notifier *notifier.Notifier) *Handler {
	return &Handler{
		cfg:      cfg,
		db:       db,
		notifier: notifier,
	}
}

func (h *Handler) Healthcheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *Handler) AddItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		Data       string    `json:"data"`
		ExpireDate time.Time `json:"expireDate"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.Data == "" || requestBody.ExpireDate.IsZero() {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	item := model.NewItem(requestBody.ExpireDate, requestBody.Data)

	err = h.db.SetItem(*item)
	if err != nil {
		http.Error(w, "Failed to save item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "Item added successfully", "id": item.ID})
}

func (h *Handler) GetItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing item ID", http.StatusBadRequest)
		return
	}

	item, err := h.db.GetItem(id)
	if err != nil {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(item)
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusInternalServerError)
		return
	}

	h.notifier.AddWebSocketClient(conn)
}

func (h *Handler) CheckExpiredItems() {
	now := time.Now().UTC()
	log.Printf("Checking expired items at %s", now.Format(time.RFC3339))

	expiredItems, err := h.db.GetExpiredItems(now)

	log.Println("Expired items:", expiredItems)

	if err != nil {
		log.Println("Error getting expired items:", err)
		return
	}

	for _, item := range expiredItems {
		h.notifier.NotifyExpiredItem(item)
		h.db.DeleteItem(item.ID)
	}
}
