package notifier

import (
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

/*func (n *Notifier) NotifyExpiredItem(item *model.Item) {
	if n.useWebSocket {
		n.notifyWebSocket(item)
	}
}

func (n *Notifier) notifyWebSocket(item *model.Item) {
	n.wsClientsMutex.RLock()
	defer n.wsClientsMutex.RUnlock()

	for client := range n.wsClients {
		err := client.WriteJSON(item)
		if err != nil {
			log.Printf("Error sending WebSocket message: %v", err)
			go n.RemoveWebSocketClient(client)
		}
	}
}*/
