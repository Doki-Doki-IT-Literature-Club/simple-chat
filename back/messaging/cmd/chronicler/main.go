package main

import (
	"context"

	"github.com/LeperGnome/simple-chat/internal/chronicler/domain"
	"github.com/LeperGnome/simple-chat/internal/shared/infrastructure"
)

const (
	groupID   = "chronicler"
	topicName = "messages"
	kafkaAddr = "kafka:9092"
)

func main() {
	pgconf := infrastructure.GetPGConfig()
	repo, err := infrastructure.NewPGMessageRepository(pgconf.DBUser, pgconf.DBPassword, pgconf.DBHost, pgconf.DBName)
	if err != nil {
		panic(err)
	}
	defer repo.Close(context.TODO())

	bus, err := infrastructure.NewKafkaMessageBus(groupID, topicName, kafkaAddr)
	if err != nil {
		panic(err)
	}

	err = domain.SyncMessages(repo, bus)
	if err != nil {
		panic(err)
	}
}
