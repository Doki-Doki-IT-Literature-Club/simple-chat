package domain

import (
	"log/slog"

	sharedDom "github.com/LeperGnome/simple-chat/internal/shared/domain"
)

func SyncMessages(repo sharedDom.MessageRepository, bus sharedDom.MessageBus) error {
	for {
		receivedMessage, err := bus.Read()
		if err != nil {
			slog.Error("Failed reading message from kafka: " + err.Error())
			return err
		}

		err = repo.InsertMessage(receivedMessage)
		if err != nil {
			slog.Error("Failed to insert message into db: " + err.Error())
			return err
		}
	}
}
