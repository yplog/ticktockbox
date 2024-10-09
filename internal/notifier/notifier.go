package notifier

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/model"
)

type Notifier struct {
	useRabbitMQ       bool
	rabbitMQURL       string
	rabbitMQQueueName string
	rabbitMQConn      *amqp.Connection
	rabbitMQCh        *amqp.Channel

	useWebSocket   bool
	wsClients      map[*websocket.Conn]bool
	wsClientsMutex sync.RWMutex
}

func NewNotifier(cfg config.NotifierConfig) *Notifier {
	notifier := &Notifier{
		useRabbitMQ:       cfg.UseRabbitMQ,
		rabbitMQURL:       cfg.RabbitMQURL,
		rabbitMQQueueName: cfg.RabbitMQQueueName,
		useWebSocket:      cfg.UseWebSocket,
		wsClients:         make(map[*websocket.Conn]bool),
	}

	if notifier.useRabbitMQ {
		conn, err := amqp.Dial(notifier.rabbitMQURL)
		if err != nil {
			log.Fatalf("Failed to connect to RabbitMQ: %v", err)
		}
		ch, err := conn.Channel()
		if err != nil {
			log.Fatalf("Failed to open a channel: %v", err)
		}
		notifier.rabbitMQConn = conn
		notifier.rabbitMQCh = ch
	}

	return notifier
}

func (n *Notifier) AddWebSocketClient(conn *websocket.Conn) {
	n.wsClientsMutex.Lock()
	defer n.wsClientsMutex.Unlock()
	n.wsClients[conn] = true
}

func (n *Notifier) RemoveWebSocketClient(conn *websocket.Conn) {
	n.wsClientsMutex.Lock()
	defer n.wsClientsMutex.Unlock()
	delete(n.wsClients, conn)
	conn.Close()
}

func (n *Notifier) NotifyExpiredItem(record *model.Record) {
	if n.useWebSocket {
		n.notifyWebSocket(record)
	}

	if n.useRabbitMQ {
		n.notifyRabbitMQ(record)
	}
}

func (n *Notifier) notifyWebSocket(record *model.Record) {
	n.wsClientsMutex.RLock()
	defer n.wsClientsMutex.RUnlock()

	for client := range n.wsClients {
		err := client.WriteJSON(record)
		if err != nil {
			log.Printf("Error sending WebSocket message: %v", err)
			go n.RemoveWebSocketClient(client)
		}
	}
}

func (n *Notifier) notifyRabbitMQ(record *model.Record) {
	body, err := json.Marshal(record.ToJSON())
	if err != nil {
		log.Printf("Failed to marshal record to JSON: %v", err)
		return
	}

	err = n.rabbitMQCh.Publish(
		"",
		n.rabbitMQQueueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Printf("Failed to publish message to RabbitMQ: %v", err)
	}
}
