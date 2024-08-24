package domain

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	AuthorID  string    `json:"author_id"`
	ChannelID string    `json:"channel_id"`
}

func NewMessage(content string, authorID string, channelID string) Message {
	return Message{
		ID:        uuid.New().String(),
		Content:   content,
		CreatedAt: time.Now(),
		AuthorID:  authorID,
		ChannelID: channelID,
	}
}

type MessageBus interface {
	Read() (Message, error)
	Write(message Message) error
}

type MessageRepository interface {
	InsertMessage(m Message) error
	GetMessages(channelID string) ([]Message, error)
}
