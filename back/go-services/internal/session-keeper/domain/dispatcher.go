package domain

import (
	"github.com/gorilla/websocket"
	"log/slog"
	"slices"
	"sync"
)

type Client struct {
	Name      string
	ChannelID string
	Ws        *websocket.Conn
}

type Dispatcher struct {
	conns map[string][]*Client
	mu    sync.RWMutex
	bus   MessageBus
}

func (d *Dispatcher) RegisterClient(client *Client) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.conns[client.ChannelID] = append(d.conns[client.ChannelID], client)
	go d.readFromClient(client)
}

func (d *Dispatcher) RemoveClient(client *Client) {
	d.mu.Lock()
	defer d.mu.Unlock()

	conns, ok := d.conns[client.ChannelID]
	if !ok {
		return
	}
	conns = slices.DeleteFunc(conns, func(c *Client) bool { return c == client })

	d.conns[client.ChannelID] = conns

	client.Ws.Close()
}

func (d *Dispatcher) readFromClient(client *Client) error {
	defer d.RemoveClient(client)
	for {
		_, msg, err := client.Ws.ReadMessage()
		if err != nil {
			return err
		}
		newMessage := Message{
			Content: string(msg),
			From:    client.Name,
			Channel: client.ChannelID,
		}
		slog.Info("Got new message from ws", slog.Any("message", newMessage))
		d.bus.Write(newMessage)
	}
}

func (d *Dispatcher) Dispatch() {
	for {
		newMessage, err := d.bus.Read()
		if err != nil {
			break
		}
		slog.Info("Got new message from bus", slog.Any("message", newMessage))

		channelConns, ok := d.conns[newMessage.Channel]
		if !ok {
			continue // TODO?
		}
		for _, otherClient := range channelConns {
			err = otherClient.Ws.WriteMessage(websocket.TextMessage, []byte(newMessage.From+": "+newMessage.Content))
			if err != nil {
				continue // TODO?
			}
		}
	}
}

func NewDispatcher(bus MessageBus) *Dispatcher {
	d := &Dispatcher{
		mu:    sync.RWMutex{},
		conns: map[string][]*Client{},
		bus:   bus,
	}

	go d.Dispatch()
	return d
}
