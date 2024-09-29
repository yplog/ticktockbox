package model

import (
	"encoding/binary"
	"encoding/json"
	"time"
)

type Item struct {
	Data string `json:"data"`
}

func NewItem(data string) *Item {
	return &Item{
		Data: data,
	}
}

func (i *Item) Value() ([]byte, error) {
	return json.Marshal(i)
}

func ItemFromValue(value []byte) (*Item, error) {
	var item Item
	err := json.Unmarshal(value, &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func MakeKey(expireTime time.Time) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, uint64(expireTime.UnixNano()))
	return key
}
