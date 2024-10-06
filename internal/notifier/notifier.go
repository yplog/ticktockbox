package notifier

import (
	"github.com/yplog/ticktockbox/internal/model"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/yplog/ticktockbox/internal/config"
)

type Notifier struct {
	useRabbitMQ    bool
	rabbitMQURL    string
	useWebSocket   bool
	wsClients      map[*websocket.Conn]bool
	wsClientsMutex sync.RWMutex
}

func NewNotifier(cfg config.NotifierConfig) *Notifier {
	return &Notifier{
		useRabbitMQ:  cfg.UseRabbitMQ,
		rabbitMQURL:  cfg.RabbitMQURL,
		useWebSocket: cfg.UseWebSocket,
		wsClients:    make(map[*websocket.Conn]bool),
	}
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
