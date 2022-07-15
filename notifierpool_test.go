package wsnotifier

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type TestJSON struct {
	Msg string `json:"msg"`
}

func setupServer(handler http.HandlerFunc) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler)
	srv := &http.Server{Addr: ":3000", Handler: mux}

	go func() {

		srv.ListenAndServe()

	}()

	return srv
}

func TestNotifierCreate(t *testing.T) {
	p := NewNotifierPool()
	p.AddBroadcast("hi", nil)

	p.RemoveBroadcast("hi")
}

func TestClientBroadcast(t *testing.T) {
	p := NewNotifierPool()
	p.AddBroadcast("hi", nil)
	srv := setupServer(p.ClientHandler("hi"))

	u := url.URL{Scheme: "ws", Host: "localhost:3000", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		rj := TestJSON{}
		for {
			err := c.ReadJSON(&rj)
			if err != nil {
				fmt.Println("Killing Client goroutine", err)
				return
			}
			fmt.Println("Received Message: ", rj)
			if rj.Msg != "Hello there!" {
				log.Fatal("Msg failed")
			}
		}
	}()

	p.BroadcastAt("hi", TestJSON{Msg: "Hello there!"})
	time.Sleep(20 * time.Millisecond)

	fmt.Println("Shutting down server")
	srv.Shutdown(context.TODO())
}
