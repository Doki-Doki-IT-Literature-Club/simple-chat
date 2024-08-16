package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Second * 10,
	WriteBufferSize:  1_000,
	ReadBufferSize:   1_000,
}

func main() {
	addr := "0.0.0.0:4444"

	h := newHandler()
	mux := http.NewServeMux()
	mux.HandleFunc("/channels/{channel_id}", h.handleConn)

	fmt.Println("Starting server at " + addr)
	http.ListenAndServe(addr, mux)
}

type handler struct {
	dispatcher *Dispatcher
}

func (h *handler) handleConn(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	channelID := r.PathValue("channel_id")
	if channelID == "" {
		return
	}
	fmt.Println("New connection in room " + channelID)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	h.dispatcher.RegisterClient(&Client{ChannelID: channelID, Name: name, Ws: ws})
}

func newHandler() handler {
	return handler{dispatcher: NewDispatcher()}
}
