package model

import (
	"fmt"
	"time"
)

type Data struct {
	Content map[string]interface{} `json:"content"`
}

func NewData(content map[string]interface{}) *Data {
	return &Data{
		Content: content,
	}
}

func MakeKey(expire time.Time) []byte {
	utcExpireTime := expire.UTC()
	fmt.Println(utcExpireTime)

	return []byte(utcExpireTime.Format(time.RFC3339))
}
