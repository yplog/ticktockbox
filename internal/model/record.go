package model

import (
	"encoding/json"
	"time"
)

type Record struct {
	ID     int                    `json:"id"`
	Expire time.Time              `json:"expire"`
	Data   map[string]interface{} `json:"data"`
}

func NewRecord(expire time.Time, data map[string]interface{}) *Record {
	return &Record{
		Expire: expire,
		Data:   data,
	}
}

func NewRecordWithID(id int, expire time.Time, data map[string]interface{}) *Record {
	return &Record{
		ID:     id,
		Expire: expire,
		Data:   data,
	}
}

func (r *Record) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"id":     r.ID,
		"expire": r.Expire,
		"data":   r.Data,
	}
}

func (r *Record) SerializeData() (string, error) {
	jsonData, err := json.Marshal(r.Data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (r *Record) DeserializeData(jsonString string) error {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		return err
	}
	r.Data = data
	return nil
}
