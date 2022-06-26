// Package wsnotifier implements a websocket utility for
// managing hubs of clients. A hub is a broadcast group.
// This package relies heavily on the example code from the
// gorilla websockets example library chat application, [https://github.com/gorilla/websocket/tree/master/examples/chat]
// extending it to multiple broadcast groups for JSON message
// passing.
package wsnotifier

import "net/http"

//A NotifierPool holds broadcast groups
type NotifierPool struct {
	hubs map[string]*BroadcastGroup
}

// NewNotifierPool returns a NotifierPool object
func NewNotifierPool() *NotifierPool {
	return &NotifierPool{hubs: make(map[string]*BroadcastGroup)}
}

// AddBroadcast adds a broadcast group to the NotifierPool
// tagged by an id. The id should be unique among all broadcast
// groups and must be maintained to send messages at the group
func (p *NotifierPool) AddBroadcast(id string) {
	p.hubs[id] = newBroadcastGroup()
	go p.hubs[id].run()
}

// BroadcastAt sends a JSON payload at each client subscribed to
// a BroadcastGroup tagged by an id.
func (p *NotifierPool) BroadcastAt(id string, payload interface{}) {
	p.hubs[id].broadcast <- payload
}

// SubscribeClient subscribes a websocket client bound to w,r to a
// BroadcastGroup tagged by an id.
func (p *NotifierPool) SubscribeClient(id string, w http.ResponseWriter, r *http.Request) {
	serveWs(p.hubs[id], w, r)
}

func (p *NotifierPool) ClientHandler(id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		serveWs(p.hubs[id], w, r)
	}
}

// RemoveBroadcast ends all client websocket connections from a
// BroadcastGroup and removes the BroadcastGroup tagged by an id
// from the NotifierPool.
func (p *NotifierPool) RemoveBroadcast(id string) {
	p.hubs[id].stop <- true
	delete(p.hubs, id)
}
