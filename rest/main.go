package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/LeperGnome/simple-chat/pkg/chronicler"
)

func main() {
	repo, err := chronicler.NewRepository(getDBConfig())

	if err != nil {
		panic(err)
	}

	repo.CreateMessageTable()

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
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

		fmt.Println(messages)

		jsonBytes, err := json.Marshal(messages)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})
	http.ListenAndServe(":8080", nil)

}
