package wsnotifier

type BroadcastGroup struct {
	clients    map[*Client]bool
	broadcast  chan interface{}
	register   chan *Client
	unregister chan *Client
	stop       chan bool
}

func newBroadcastGroup() *BroadcastGroup {
	return &BroadcastGroup{
		broadcast:  make(chan interface{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		stop:       make(chan bool),
	}
}

func (h *BroadcastGroup) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.sendJson)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.sendJson <- message:
				default:
					close(client.sendJson)
					delete(h.clients, client)
				}
			}
		case stop := <-h.stop:
			if stop {
				for client := range h.clients {
					close(client.sendJson)
					delete(h.clients, client)
				}
				return
			}
		}
	}
}
