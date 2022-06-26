package wsnotifier

import "net/http"

type NotifierPool struct {
	hubs map[string]*Hub
}

func NewNotifierPool() *NotifierPool {
	return &NotifierPool{hubs: make(map[string]*Hub)}
}

func (p *NotifierPool) AddBroadcast(id string) {
	p.hubs[id] = newHub()
	go p.hubs[id].run()
}

func (p *NotifierPool) BroadcastAt(id string, payload interface{}) {
	p.hubs[id].broadcast <- payload
}

func (p *NotifierPool) SubscribeClient(id string, w http.ResponseWriter, r *http.Request) {
	serveWs(p.hubs[id], w, r)
}

func (p *NotifierPool) RemoveBroadcast(id string) {
	p.hubs[id].stop <- true
	delete(p.hubs, id)
}
