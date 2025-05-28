package scheduler

import (
	"log"
	"time"

	"github.com/yplog/ticktockbox/internal/database"
	"github.com/yplog/ticktockbox/internal/rabbitmq"
	"github.com/yplog/ticktockbox/internal/websocket"
)

type Scheduler struct {
	db              *database.QuestDB
	rmq             *rabbitmq.RabbitMQ
	wsHub           *websocket.Hub
	ticker          *time.Ticker
	cleanupTicker   *time.Ticker
	done            chan bool
	lastCleanupTime time.Time
}

func New(db *database.QuestDB, rmq *rabbitmq.RabbitMQ, wsHub *websocket.Hub) *Scheduler {
	return &Scheduler{
		db:    db,
		rmq:   rmq,
		wsHub: wsHub,
		done:  make(chan bool),
	}
}

func (s *Scheduler) Start(interval time.Duration) {
	s.ticker = time.NewTicker(interval)

	s.cleanupTicker = time.NewTicker(24 * time.Hour)
	s.lastCleanupTime = time.Now()

	log.Printf("Scheduler started with interval: %v", interval)
	log.Printf("Partition cleanup will run every 24 hours")

	for {
		select {
		case <-s.ticker.C:
			s.processExpiredMessages()
		case <-s.cleanupTicker.C:
			s.cleanupOldPartitions()
		case <-s.done:
			s.ticker.Stop()
			if s.cleanupTicker != nil {
				s.cleanupTicker.Stop()
			}
			return
		}
	}
}

func (s *Scheduler) Stop() {
	s.done <- true
}

func (s *Scheduler) processExpiredMessages() {
	log.Printf("Scheduler tick: checking for expired messages...")

	expiredMessages, err := s.db.GetExpiredMessages()
	if err != nil {
		log.Printf("Failed to get expired messages: %v", err)
		return
	}

	log.Printf("Found %d expired messages", len(expiredMessages))

	if len(expiredMessages) == 0 {
		return
	}

	log.Printf("Processing %d expired messages", len(expiredMessages))

	if err := s.rmq.PublishExpiredMessages(expiredMessages); err != nil {
		log.Printf("Failed to publish expired messages to RabbitMQ: %v", err)
		return
	}

	s.wsHub.BroadcastExpiredMessages(expiredMessages)

	if err := s.db.MarkAsProcessed(expiredMessages); err != nil {
		log.Printf("Failed to mark messages as processed: %v", err)
		return
	}

	log.Printf("Successfully processed %d expired messages", len(expiredMessages))
}

func (s *Scheduler) cleanupOldPartitions() {
	log.Printf("Starting partition cleanup...")

	if err := s.db.CleanupOldPartitions(); err != nil {
		log.Printf("Failed to cleanup old partitions: %v", err)
		return
	}

	s.lastCleanupTime = time.Now()
	log.Printf("Partition cleanup completed successfully at %v", s.lastCleanupTime)
}
