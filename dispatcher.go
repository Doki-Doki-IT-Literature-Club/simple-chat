package main

import (
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

type Dispatcher struct {
	conns map[string][]*websocket.Conn
	mu    sync.RWMutex
}

func (d *Dispatcher) RegisterClient(channelID string, ws *websocket.Conn) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.conns[channelID] = append(d.conns[channelID], ws)
	go d.dispatchCilent(channelID, ws)
}

func (d *Dispatcher) RemoveClient(channelID string, ws *websocket.Conn) {
	d.mu.Lock()
	defer d.mu.Unlock()

	conns, ok := d.conns[channelID]
	if !ok {
		return
	}
	conns = slices.DeleteFunc(conns, func(w *websocket.Conn) bool { return w == ws })
	ws.Close()
}

func (d *Dispatcher) dispatchCilent(channelID string, ws *websocket.Conn) error {
	defer d.RemoveClient(channelID, ws)
	for {
		t, msg, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		otherConns, ok := d.conns[channelID]
		if !ok {
			return nil // TODO: that's just odd
		}
		for _, otherWS := range otherConns {
			err = otherWS.WriteMessage(t, msg)
			if err != nil {
				return err
			}
		}
	}
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{mu: sync.RWMutex{}, conns: map[string][]*websocket.Conn{}}
}
