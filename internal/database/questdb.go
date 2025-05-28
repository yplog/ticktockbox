package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	_ "github.com/lib/pq"
	"github.com/yplog/ticktockbox/internal/config"
)

type QuestDB struct {
	db      *sql.DB
	httpURL string
	client  *http.Client
}

type Message struct {
	MessageID int64     `json:"message_id"`
	Message   string    `json:"message"`
	ExpireAt  time.Time `json:"expire_at"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

func NewQuestDB(cfg *config.Config) (*QuestDB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
		cfg.QuestDBUser, cfg.QuestDBPass, cfg.QuestDBURL, cfg.QuestDBName, cfg.QuestDBSSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	qdb := &QuestDB{
		db:      db,
		httpURL: cfg.QuestDBHTTPURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	if err := qdb.createTables(); err != nil {
		return nil, err
	}

	return qdb, nil
}

func (q *QuestDB) createTables() error {
	messagesTable := `
		CREATE TABLE IF NOT EXISTS messages (
			ts TIMESTAMP,
			message_id LONG,
			message STRING,
			expire_at TIMESTAMP,
			event_type SYMBOL
		) timestamp(ts) PARTITION BY DAY;
	`

	if _, err := q.db.Exec(messagesTable); err != nil {
		return fmt.Errorf("failed to create messages table: %w", err)
	}

	return nil
}

func (q *QuestDB) InsertMessage(message string, expireAt time.Time) error {
	messageID := time.Now().UnixNano()

	query := `
		INSERT INTO messages (ts, message_id, message, expire_at, event_type) 
		VALUES ($1, $2, $3, $4, $5)
	`

	now := time.Now().UTC()
	_, err := q.db.Exec(query, now, messageID, message, expireAt.UTC(), "created")
	return err
}

func (q *QuestDB) GetExpiredMessages() ([]Message, error) {
	query := `
		WITH latest_messages AS (
			SELECT
				max(ts) AS ts,
				message_id
			FROM messages 
			GROUP BY message_id
		)
		SELECT
			m.message_id,
			m.message,
			m.expire_at,
			m.event_type,
			m.ts
		FROM latest_messages lm
		INNER JOIN messages m ON lm.ts = m.ts AND lm.message_id = m.message_id
		WHERE m.expire_at <= $1 AND m.event_type = 'created'
	`

	now := time.Now().UTC()
	rows, err := q.db.Query(query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.MessageID, &msg.Message, &msg.ExpireAt, &msg.EventType, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (q *QuestDB) GetAllMessages() ([]Message, error) {
	query := `
		WITH latest_messages AS (
			SELECT
				max(ts) AS ts,
				message_id
			FROM messages 
			GROUP BY message_id
		)
		SELECT
			m.message_id,
			m.message,
			m.expire_at,
			m.event_type,
			m.ts
		FROM latest_messages lm
		INNER JOIN messages m ON lm.ts = m.ts AND lm.message_id = m.message_id
		ORDER BY m.ts DESC
	`

	rows, err := q.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.MessageID, &msg.Message, &msg.ExpireAt, &msg.EventType, &msg.Timestamp); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

func (q *QuestDB) executeHTTPQuery(query string) error {
	encodedQuery := url.QueryEscape(query)
	requestURL := fmt.Sprintf("%s/exec?query=%s", q.httpURL, encodedQuery)

	resp, err := q.client.Get(requestURL)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err == nil {
		if errorMsg, exists := response["error"]; exists {
			return fmt.Errorf("QuestDB error: %v", errorMsg)
		}
	}

	return nil
}

func (q *QuestDB) CleanupOldPartitions() error {
	twoWeeksAgo := time.Now().UTC().AddDate(0, 0, -14)

	cutoffDate := twoWeeksAgo.Format("2006-01-02")

	dropQuery := fmt.Sprintf("ALTER TABLE messages DROP PARTITION WHERE timestamp < '%s'", cutoffDate)

	if err := q.executeHTTPQuery(dropQuery); err != nil {
		return fmt.Errorf("failed to drop old partitions: %w", err)
	}

	return nil
}

func (q *QuestDB) MarkAsProcessed(messages []Message) error {
	for _, msg := range messages {
		query := `
			INSERT INTO messages (ts, message_id, message, expire_at, event_type) 
			VALUES ($1, $2, $3, $4, $5)
		`

		now := time.Now().UTC()
		_, err := q.db.Exec(query, now, msg.MessageID, msg.Message, msg.ExpireAt.UTC(), "processed")
		if err != nil {
			return fmt.Errorf("failed to mark message %d as processed: %w", msg.MessageID, err)
		}
	}
	return nil
}

func (q *QuestDB) Close() error {
	return q.db.Close()
}
