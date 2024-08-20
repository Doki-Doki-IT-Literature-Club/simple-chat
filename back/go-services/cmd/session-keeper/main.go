package main

import (
	"log/slog"
	"os"

	"github.com/LeperGnome/simple-chat/internal/session-keeper/application"
	"github.com/kelseyhightower/envconfig"
)

func main() {
	var config application.Config

	envconfig.MustProcess("", &config)

	server := application.NewServer(config)
	err := server.Run()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
