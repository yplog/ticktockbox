package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/yplog/ticktockbox/internal/model"
)

type Database struct {
	DB *sql.DB
}

func InitDatabase(path string) (*Database, error) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
    CREATE TABLE IF NOT EXISTS records (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        expire DATETIME,
        data TEXT
    );
    `
	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	log.Println("Database initialized and table created")
	return &Database{DB: db}, nil
}

func (db *Database) CreateRecord(record *model.Record) (*model.Record, error) {
	stmt, err := db.DB.Prepare("INSERT INTO records(expire, data) VALUES(?, ?)")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("Error closing statement:", err)
		}
	}()

	jsonData, err := record.SerializeData()
	if err != nil {
		return nil, err
	}

	result, err := stmt.Exec(record.Expire, jsonData)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	record.ID = int(id)

	return record, nil
}

func (db *Database) GetExpireRecords() ([]*model.Record, error) {
	rows, err := db.DB.Query("SELECT id, expire, data FROM records WHERE expire <= datetime('now')")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Println("Error closing rows:", err)
		}
	}()

	var records []*model.Record
	for rows.Next() {
		var id int
		var expire time.Time
		var jsonDataStr string

		if err := rows.Scan(&id, &expire, &jsonDataStr); err != nil {
			return nil, err
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(jsonDataStr), &jsonData); err != nil {
			return nil, err
		}

		record := model.NewRecordWithID(id, expire, jsonData)
		records = append(records, record)
	}

	return records, nil
}

func (db *Database) DeleteRecords(ids []int) (int64, error) {
	if len(ids) == 0 {
		return 0, fmt.Errorf("no ids provided")
	}

	placeholders := make([]string, len(ids))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("DELETE FROM records WHERE id IN (%s)", strings.Join(placeholders, ","))

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Println("Error closing statement:", err)
		}
	}()

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected == 0 {
		return 0, sql.ErrNoRows
	}

	return rowsAffected, nil
}
