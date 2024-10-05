package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type RequestData struct {
	Expire time.Time              `json:"expire"`
	Data   map[string]interface{} `json:"data"`
}

func DecodeRequestBody(r *http.Request) (*RequestData, error) {
	var reqData RequestData

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		return nil, err
	}

	return &reqData, nil
}
