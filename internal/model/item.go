package model

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/klauspost/compress/zstd"
)

const (
	PrefixExpire = "expire:"
	PrefixLength = len(PrefixExpire)
)

var encoder, _ = zstd.NewWriter(nil)

type Item struct {
	Data       string    `json:"data"`
	ExpireTime time.Time `json:"expireTime"`
}

func NewItem(data string, expireTime time.Time) *Item {
	return &Item{
		Data:       data,
		ExpireTime: expireTime,
	}
}

func (i *Item) Value() ([]byte, error) {
	jsonData, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	compressed := encoder.EncodeAll(jsonData, nil)
	value := make([]byte, 4+len(compressed))
	binary.BigEndian.PutUint32(value[:4], uint32(len(compressed)))
	copy(value[4:], compressed)

	return value, nil
}

func ItemFromValue(value []byte) (*Item, error) {
	var item Item
	err := json.Unmarshal(value, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal item: %v", err)
	}
	item.ExpireTime = item.ExpireTime.UTC()
	return &item, nil
}

func MakeKey(expireTime time.Time) []byte {
	utcExpireTime := expireTime.UTC()
	key := make([]byte, PrefixLength+8)
	copy(key, PrefixExpire)
	binary.BigEndian.PutUint64(key[PrefixLength:], uint64(utcExpireTime.UnixNano()))
	return key
}
