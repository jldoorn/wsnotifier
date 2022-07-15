package wsnotifier

type BroadcastGroup struct {
	clients         map[*Client]bool
	broadcast       chan interface{}
	register        chan *Client
	unregister      chan *Client
	stop            chan bool
	notifyPoolEmpty chan<- bool
}

func newBroadcastGroup(notifyEmpty chan<- bool) *BroadcastGroup {
	return &BroadcastGroup{
		broadcast:       make(chan interface{}),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		clients:         make(map[*Client]bool),
		stop:            make(chan bool),
		notifyPoolEmpty: notifyEmpty,
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
			if len(h.clients) == 0 && h.notifyPoolEmpty != nil {
				h.notifyPoolEmpty <- true
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
