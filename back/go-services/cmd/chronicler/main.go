package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	infrastructure "github.com/LeperGnome/simple-chat/internal/shared/infrastructure"
	"github.com/segmentio/kafka-go"
)

const (
	topicName = "messages"
	kafkaAddr = "kafka:9092"
)

type Message struct {
	From    string `json:"from"`
	Content string `json:"content"`
	Channel string `json:"channel"`
}

func main() {
	repo, err := infrastructure.NewRepository(infrastructure.GetDBConfig())

	if err != nil {
		panic(err)
	}

	repo.CreateMessageTable()

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	hasConnected := false
	for !hasConnected {
		conn, err := dialer.DialLeader(context.Background(), "tcp", kafkaAddr, topicName, 0)
		if err != nil {
			slog.Error("Failed to connect to kafka: " + err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		conn.Close()
		hasConnected = true
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{kafkaAddr},
		GroupID: "chronicler",
		Topic:   topicName,
		Dialer:  dialer,
	})

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			slog.Error("Failed reading message from kafka: " + err.Error())
			break // TODO
		}
		var receivedMessage Message
		json.Unmarshal(msg.Value, &receivedMessage)
		slog.Info("Got new message from kafka", slog.Any("message", receivedMessage))

		newMessage := infrastructure.NewMessage(receivedMessage.Content, receivedMessage.From, receivedMessage.Channel)
		err = repo.InsertMessage(newMessage)

		if err != nil {
			slog.Error("Failed to insert message into db: " + err.Error())
		}
	}

}
