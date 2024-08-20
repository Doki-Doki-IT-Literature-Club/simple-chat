package main

import (
	"encoding/json"
	"log"
	"net/http"

	infrastructure "github.com/LeperGnome/simple-chat/internal/shared/infrastructure"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		next.ServeHTTP(w, r)
	})
}

func newMessageHandler(repo *infrastructure.Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		channelID := r.URL.Query().Get("channel_id")

		if channelID == "" {
			http.Error(w, "channel_id is required", http.StatusBadRequest)
			return
		}

		messages, err := repo.GetMessages(channelID)

		if err != nil {
			log.Printf("Error getting messages: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		jsonBytes, err := json.Marshal(messages)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})
}

func main() {
	repo, err := infrastructure.NewRepository(infrastructure.GetDBConfig())

	if err != nil {
		panic(err)
	}

	repo.CreateMessageTable()

	messageHandler := newMessageHandler(repo)

	http.Handle("/messages", corsMiddleware(messageHandler))
	http.ListenAndServe(":8080", nil)

}
