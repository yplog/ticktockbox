package model

import (
	"encoding/json"
	"log"
	"time"
)

type ExpireData struct {
	ExpireTime uint64 `json:"expire_time"`
}

func NewExpireData(expireTime time.Time) *ExpireData {
	return &ExpireData{
		ExpireTime: uint64(expireTime.UnixNano()),
	}
}

func (e *ExpireData) ToBytes() []byte {
	data, err := json.Marshal(e)
	if err != nil {
		log.Printf("Failed to marshal ExpireData: %v", err)
		return nil
	}
	return data
}

func (e *ExpireData) ToTime() time.Time {
	return time.Unix(0, int64(e.ExpireTime))
}

func ExpireDataFromBytes(data []byte) *ExpireData {
	var e ExpireData

	err := json.Unmarshal(data, &e)
	if err != nil {
		log.Printf("Failed to unmarshal ExpireData: %v", err)
		return nil
	}

	return &e
}
