package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/rabbitmq"
)

type Handler struct {
	db  *database.QuestDB
	rmq *rabbitmq.RabbitMQ
}

type CreateMessageRequest struct {
	Message  string    `json:"message"`
	ExpireAt time.Time `json:"expire_at"`
}

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewHandler(db *database.QuestDB, rmq *rabbitmq.RabbitMQ) *Handler {
	return &Handler{
		db:  db,
		rmq: rmq,
	}
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var req CreateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		h.sendError(w, "Message is required", http.StatusBadRequest)
		return
	}

	if req.ExpireAt.IsZero() {
		h.sendError(w, "ExpireAt is required", http.StatusBadRequest)
		return
	}

	if req.ExpireAt.Before(time.Now()) {
		h.sendError(w, "ExpireAt must be in the future", http.StatusBadRequest)
		return
	}

	if err := h.db.InsertMessage(req.Message, req.ExpireAt); err != nil {
		h.sendError(w, "Failed to create message", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, "Message created successfully", nil)
}

func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.db.GetAllMessages()
	if err != nil {
		h.sendError(w, "Failed to get messages", http.StatusInternalServerError)
		return
	}

	h.sendSuccess(w, "", messages)
}

func (h *Handler) sendSuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := Response{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: false,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
