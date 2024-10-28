package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yplog/ticktockbox/internal/model"
)

func setupTestDB(t *testing.T) (*Database, func()) {
	t.Helper()

	// Create a temporary database file
	dbPath := filepath.Join(os.TempDir(), "test_db.sqlite")
	db, err := InitDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	cleanup := func() {
		db.DB.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestInitDatabase(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	if db.DB == nil {
		t.Fatal("Database not initialized; DB object is nil")
	}
}

func TestCreateRecord(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	record := &model.Record{
		Expire: time.Now().Add(10 * time.Minute),
		Data:   map[string]interface{}{"key": "value"},
	}

	createdRecord, err := db.CreateRecord(record)
	if err != nil {
		t.Fatalf("Error creating record: %v", err)
	}

	if createdRecord.ID == 0 {
		t.Error("Record ID was not set as expected")
	}
}

func TestGetExpireRecords(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	futureRecord := &model.Record{
		Expire: time.Now().Add(10 * time.Hour),
		Data:   map[string]interface{}{"future": "data"},
	}
	if _, err := db.CreateRecord(futureRecord); err != nil {
		t.Fatalf("Failed to create future record: %v", err)
	}

	expiredRecord := &model.Record{
		Expire: time.Now().Add(-10 * time.Hour),
		Data:   map[string]interface{}{"expired": "data"},
	}
	if _, err := db.CreateRecord(expiredRecord); err != nil {
		t.Fatalf("Failed to create expired record: %v", err)
	}

	records, err := db.GetExpireRecords()
	if err != nil {
		t.Fatalf("Error fetching expired records: %v", err)
	}

	if len(records) != 1 {
		t.Errorf("Expected 1 expired record, but got %d", len(records))
		return
	}

	if records[0].Data["expired"] != "data" {
		t.Error("Expected expired record data not retrieved")
	}
}

func TestDeleteRecords(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	record := &model.Record{
		Expire: time.Now().Add(-10 * time.Minute),
		Data:   map[string]interface{}{"delete": "me"},
	}
	createdRecord, _ := db.CreateRecord(record)

	rowsAffected, err := db.DeleteRecords([]int{createdRecord.ID})
	if err != nil {
		t.Fatalf("Error deleting record: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("Expected 1 row affected, but got %d", rowsAffected)
	}

	records, err := db.GetExpireRecords()
	if err != nil {
		t.Fatalf("Error fetching expired records: %v", err)
	}

	for _, rec := range records {
		if rec.ID == createdRecord.ID {
			t.Error("Deleted record still exists in the database")
		}
	}
}
