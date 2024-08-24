package main

import (
	"log/slog"
	"os"

	"github.com/LeperGnome/simple-chat/internal/session-keeper/application"
	"github.com/LeperGnome/simple-chat/internal/shared/infrastructure"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var config application.Config

	envconfig.MustProcess("", &config)

	bus, err := infrastructure.NewKafkaMessageBus("messages", "kafka:9092")
	if err != nil {
		panic(err)
	}

	server := application.NewServer(config, bus)
	err = server.Run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
