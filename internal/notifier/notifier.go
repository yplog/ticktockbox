package notifier

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/yplog/ticktockbox/internal/config"
	"github.com/yplog/ticktockbox/internal/model"
)

type Notifier struct {
	config         config.Config
	wsClients      map[*websocket.Conn]bool
	wsClientsMutex sync.RWMutex
}

func NewNotifier(config config.Config) *Notifier {
	return &Notifier{
		config:    config,
		wsClients: make(map[*websocket.Conn]bool),
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
}

func (n *Notifier) NotifyExpiredItem(item model.Item) {
	if n.config.Notifier.UseWebSocket {
		n.notifyWebSocket(item)
	}
}

func (n *Notifier) notifyWebSocket(item model.Item) {
	n.wsClientsMutex.RLock()
	defer n.wsClientsMutex.RUnlock()

	for client := range n.wsClients {
		err := client.WriteJSON(item)
		if err != nil {
			log.Printf("Error sending WebSocket message: %v", err)
			client.Close()
			delete(n.wsClients, client)
		}
	}
}
