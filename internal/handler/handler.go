package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/model"
	"github.com/yplog/ticktockbox/internal/notifier"
	"log"
	"net/http"
	"time"
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

func (h *Handler) CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	reqData, err := DecodeRequestBody(r)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	data := model.NewRecord(reqData.Expire, reqData.Data)

	record, err := h.db.CreateRecord(data)
	if err != nil {
		log.Printf("Failed to create data: %v", err)
		http.Error(w, "Failed to create data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record.ToJSON())
}

func (h *Handler) GetExpireRecordsHandler() {
	expiredRecords, err := h.db.GetExpireRecords()
	if err != nil {
		log.Printf("Failed to get expire records: %v", err)
		return
	}

	log.Printf("Found %d expired items", len(expiredRecords))

	for _, record := range expiredRecords {
		log.Printf("Processing expired item: Data=%s, ExpireTime=%s (UTC)",
			record.Data, record.Expire.UTC().Format(time.RFC3339))

		h.notifier.NotifyExpiredItem(record)
		log.Printf("Successfully deleted, verified, and notified about item: %s", record.Data)
	}

	var ids []int
	for _, record := range expiredRecords {
		ids = append(ids, record.ID)
	}

	_, err = h.db.DeleteRecords(ids)
	if err != nil {
		log.Printf("Failed to delete records: %v", err)
		return
	}

	log.Printf("Successfully deleted records: %v", ids)
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Could not open websocket connection: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusInternalServerError)
		return
	}

	h.notifier.AddWebSocketClient(conn)
}
