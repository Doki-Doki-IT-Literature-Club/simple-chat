package main

import (
	"context"
	"github.com/Doki-Doki-IT-Literature-Club/simple-chat/internal/chat-api/application"
	"github.com/Doki-Doki-IT-Literature-Club/simple-chat/internal/shared/infrastructure"
)

func main() {
	pgconf := infrastructure.GetPGConfig()
	repo, err := infrastructure.NewPGMessageRepository(pgconf.DBUser, pgconf.DBPassword, pgconf.DBHost, pgconf.DBName)
	if err != nil {
		panic(err)
	}
	defer repo.Close(context.TODO())

	addr := "0.0.0.0:8080"
	server := application.NewServer(addr, repo)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
