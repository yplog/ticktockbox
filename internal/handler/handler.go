package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/model"
	"github.com/yplog/ticktockbox/internal/notifier"
	"log"
	"net/http"
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

	id, err := h.db.GetNextID()
	if err != nil {
		log.Printf("Failed to get next ID: %v", err)
		http.Error(w, "Failed to get next ID", http.StatusInternalServerError)
		return
	}

	fmt.Println("id", id)

	data := model.NewData(reqData.Data)
	expireData := model.NewExpireData(reqData.Expire)

	err = h.db.SetExpireData(&id, expireData)
	if err != nil {
		log.Printf("Failed to set expire data: %v", err)
		http.Error(w, "Failed to get next ID", http.StatusInternalServerError)
		return
	}

	err = h.db.SetData(&id, data)
	if err != nil {
		log.Printf("Failed to set data: %v", err)
		http.Error(w, "Failed to set data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":     id.ToUint64(),
		"expires": expireData.ToTime(),
		"data":    data.Content,
	})
}

/*func (h *Handler) ShowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	expireDateStr := r.URL.Query().Get("expireDate")
	if expireDateStr == "" {
		http.Error(w, "Missing expireDate", http.StatusBadRequest)
		return
	}

	expireDate, err := time.Parse(time.RFC3339, expireDateStr)
	if err != nil {
		http.Error(w, "Invalid expireDate format", http.StatusBadRequest)
		return
	}

	item, err := h.db.GetItem(expireDate)
	if err != nil {
		log.Printf("Failed to get item: %v", err)
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(item)
}*/

func (h *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetAllItem")
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

/*func (h *Handler) CheckExpiredItems() {
	now := time.Now().UTC()
	log.Printf("Checking expired items at %s (UTC)", now.Format(time.RFC3339))

	expiredItems, err := h.db.GetExpiredItems(now)
	if err != nil {
		log.Printf("Error getting expired items: %v", err)
		return
	}

	log.Printf("Found %d expired items", len(expiredItems))

	for _, item := range expiredItems {
		log.Printf("Processing expired item: Data=%s, ExpireTime=%s (UTC)",
			item.Data, item.ExpireTime.UTC().Format(time.RFC3339))

		key := model.MakeKey(item.ExpireTime.UTC())
		if err := h.db.DeleteAndVerify(key); err != nil {
			log.Printf("Error deleting and verifying expired item: %v", err)
		} else {
			h.notifier.NotifyExpiredItem(item)
			log.Printf("Successfully deleted, verified, and notified about item: %s", item.Data)
		}
	}
}*/
