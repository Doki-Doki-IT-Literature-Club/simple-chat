package infrastructure

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/LeperGnome/simple-chat/internal/session-keeper/domain"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaMessageBus struct {
	reader *kafka.Reader
	writer *kafka.Writer
}

func (k *KafkaMessageBus) Read() (domain.Message, error) {
	var newMessage domain.Message
	msg, err := k.reader.ReadMessage(context.Background())
	if err != nil {
		slog.Error("Failed reading message from kafka: " + err.Error())
		return newMessage, err
	}
	json.Unmarshal(msg.Value, &newMessage)

	return newMessage, nil
}

func (k *KafkaMessageBus) Write(message domain.Message) error {
	msgb, err := json.Marshal(message)
	if err != nil {
		return err
	}
	k.writer.WriteMessages(context.TODO(), kafka.Message{
		Key:   []byte("const"),
		Value: msgb,
	})
	return nil
}

func NewKafkaMessageBus(
	topicName string,
	kafkaAddr string,
) (*KafkaMessageBus, error) {
	groupUUID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	groupName := groupUUID.String()

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
		GroupID: groupName,
		Topic:   topicName,
		Dialer:  dialer,
	})

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{kafkaAddr},
		Topic:    topicName,
		Balancer: &kafka.Hash{},
		Dialer:   dialer,
	})
	writer.AllowAutoTopicCreation = true
	return &KafkaMessageBus{
		reader: reader,
		writer: writer,
	}, nil
}
