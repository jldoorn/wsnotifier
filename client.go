package wsnotifier

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Client struct {
	hub *BroadcastGroup

	conn     *websocket.Conn
	sendJson chan interface{}
}

type ClientMessage interface{}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		fmt.Println("Closing Connection from readPump")
		c.conn.Close()
	}()

	var msg *ClientMessage

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(appData string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		msg = new(ClientMessage)
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			fmt.Println("readPump error")
			break
		}
		c.hub.broadcast <- msg
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod) // Ensures that the client is still active
	defer func() {
		ticker.Stop()
		fmt.Println("Closing Connection from writePump")

		c.conn.Close()
	}()
	for {
		select {
		case jsonMessage, ok := <-c.sendJson:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteJSON(jsonMessage); err != nil {
				return
			}
			n := len(c.sendJson)
			for i := 0; i < n; i++ {
				if err := c.conn.WriteJSON(<-c.sendJson); err != nil {
					return
				}
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *BroadcastGroup, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, sendJson: make(chan interface{}, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
