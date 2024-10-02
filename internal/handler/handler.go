package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
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

type RequestData struct {
	Timezone string                 `json:"timezone"`
	Expire   string                 `json:"expire"`
	Data     map[string]interface{} `json:"data"`
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

	var reqData RequestData
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reqLoc, err := time.LoadLocation(reqData.Timezone)
	if err != nil {
		http.Error(w, "Invalid timezone", http.StatusBadRequest)
		return
	}

	expire, err := time.ParseInLocation(time.RFC3339, reqData.Expire, reqLoc)
	if err != nil {
		http.Error(w, "Invalid expire time", http.StatusBadRequest)
		return
	}

	localLoc, err := time.LoadLocation("Local")
	if err != nil {
		http.Error(w, "Failed to load local timezone", http.StatusInternalServerError)
		return
	}

	localExpire := expire.In(localLoc)
	key := strconv.FormatInt(localExpire.UnixNano(), 10)

	data := model.NewData(reqData.Data)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"key":           key,
		"local_expires": localExpire.Format(time.RFC3339),
		"expires":       expire.Format(time.RFC3339),
		"data":          fmt.Sprintf("%v", data.Content),
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
