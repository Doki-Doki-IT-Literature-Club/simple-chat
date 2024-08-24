package domain

type Message struct {
	From    string `json:"from"`
	Content string `json:"content"`
	Channel string `json:"channel"`
}

type MessageBus interface {
    Read() (Message, error)
    Write(message Message) error
}
