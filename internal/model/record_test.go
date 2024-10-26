package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewRecord(t *testing.T) {
	expire := time.Now().Add(24 * time.Hour)
	data := map[string]interface{}{"key": "value"}
	record := NewRecord(expire, data)

	if record.Expire != expire {
		t.Errorf("expected %v, got %v", expire, record.Expire)
	}
	if record.Data["key"] != "value" {
		t.Errorf("expected %v, got %v", "value", record.Data["key"])
	}
}

func TestNewRecordWithID(t *testing.T) {
	id := 1
	expire := time.Now().Add(24 * time.Hour)
	data := map[string]interface{}{"key": "value"}
	record := NewRecordWithID(id, expire, data)

	if record.ID != id {
		t.Errorf("expected %v, got %v", id, record.ID)
	}
	if record.Expire != expire {
		t.Errorf("expected %v, got %v", expire, record.Expire)
	}
	if record.Data["key"] != "value" {
		t.Errorf("expected %v, got %v", "value", record.Data["key"])
	}
}

func TestRecord_ToJSON(t *testing.T) {
	record := NewRecordWithID(1, time.Now().Add(24*time.Hour), map[string]interface{}{"key": "value"})
	jsonData := record.ToJSON()

	if jsonData["id"] != 1 {
		t.Errorf("expected %v, got %v", 1, jsonData["id"])
	}
	if jsonData["data"].(map[string]interface{})["key"] != "value" {
		t.Errorf("expected %v, got %v", "value", jsonData["data"].(map[string]interface{})["key"])
	}
}

func TestRecord_SerializeData(t *testing.T) {
	record := NewRecord(time.Now().Add(24*time.Hour), map[string]interface{}{"key": "value"})
	jsonString, err := record.SerializeData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonString), &data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data["key"] != "value" {
		t.Errorf("expected %v, got %v", "value", data["key"])
	}
}

func TestRecord_DeserializeData(t *testing.T) {
	record := NewRecord(time.Now().Add(24*time.Hour), nil)
	jsonString := `{"key": "value"}`
	err := record.DeserializeData(jsonString)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if record.Data["key"] != "value" {
		t.Errorf("expected %v, got %v", "value", record.Data["key"])
	}
}
