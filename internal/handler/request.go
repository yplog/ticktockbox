package handler

import (
	"encoding/json"
	"errors"
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

	err := validateExpire(reqData.Expire)
	if err != nil {
		return nil, err
	}

	return &reqData, nil
}

func validateExpire(expire time.Time) error {
	if _, err := time.Parse(time.RFC3339, expire.Format(time.RFC3339)); err != nil {
		log.Printf("Expire time format is invalid: %v", expire)
		return errors.New("expire time format is invalid")
	}

	if expire.IsZero() {
		log.Printf("Expire time is not a valid time: %v", expire)
		return errors.New("expire time must be a valid time")
	}

	if expire.Before(time.Now()) {
		log.Printf("Expire time is in the past: %v", expire)
		return errors.New("expire time must be in the future")
	}

	return nil
}
