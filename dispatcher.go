package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

const (
	topicName = "messages"
	userName  = "user1"
	password  = "..."
)

var brokers = []string{
	"simple-chat-kafka-controller-0.simple-chat-kafka-controller-headless.default.svc.cluster.local:9092",
	"simple-chat-kafka-controller-1.simple-chat-kafka-controller-headless.default.svc.cluster.local:9092",
	"simple-chat-kafka-controller-2.simple-chat-kafka-controller-headless.default.svc.cluster.local:9092",
}

type Client struct {
	Name      string
	ChannelID string
	Ws        *websocket.Conn
}

type Message struct {
	From    string `json:"from"`
	Content string `json:"content"`
	Channel string `json:"channel"`
}

type Dispatcher struct {
	conns       map[string][]*Client
	mu          sync.RWMutex
	kafkaReader *kafka.Reader
	kafkaWriter *kafka.Writer
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
		msgb, err := json.Marshal(newMessage)
		if err != nil {
			continue // TODO?
		}
		d.kafkaWriter.WriteMessages(context.TODO(), kafka.Message{
			Key:   []byte("const"),
			Value: msgb,
		})
		slog.Info("Got new message from kafka", slog.Any("message", newMessage))
	}
}

func (d *Dispatcher) Dispatch() {
	for {
		msg, err := d.kafkaReader.ReadMessage(context.Background())
		if err != nil {
			slog.Error("Failed reading message from kafka: " + err.Error())
			break // TODO
		}
		var newMessage Message
		json.Unmarshal(msg.Value, &newMessage)
		slog.Info("Got new message from kafka", slog.Any("message", newMessage))

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

func NewDispatcher() *Dispatcher {
	mechanism, err := scram.Mechanism(scram.SHA512, userName, password)
	if err != nil {
		panic(err)
	}
	groupUUID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	groupName := groupUUID.String()

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupName,
		Topic:   topicName,
		Dialer:  dialer,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topicName,
		Balancer: &kafka.Hash{},
		Dialer:   dialer,
	})
	writer.AllowAutoTopicCreation = true

	d := &Dispatcher{
		mu:          sync.RWMutex{},
		conns:       map[string][]*Client{},
		kafkaReader: reader,
		kafkaWriter: writer,
	}
	go d.Dispatch()
	return d
}
