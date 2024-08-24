package application

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	dom "github.com/LeperGnome/simple-chat/internal/session-keeper/domain"
	"github.com/gorilla/websocket"
)

type Config struct {
	Port             int           `required:"true"`
	HandshakeTimeout time.Duration `split_words:"true" default:"10s"`
	WriteBufferSize  int           `split_words:"true" default:"1000"`
	ReadBufferSize   int           `split_words:"true" default:"1000"`
}

type Server struct {
	config     Config
	upgrader   websocket.Upgrader
	dispatcher *dom.Dispatcher
}

func (s *Server) Run() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.config.Port)

	mux := http.NewServeMux()
	mux.HandleFunc("/channels/{channel_id}", s.handleConn)

	slog.Info("Starting server at " + addr)
	return http.ListenAndServe(addr, mux)
}

func NewServer(config Config, bus dom.MessageBus) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			HandshakeTimeout: config.HandshakeTimeout,
			WriteBufferSize:  config.WriteBufferSize,
			ReadBufferSize:   config.ReadBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		dispatcher: dom.NewDispatcher(bus),
		config:     config,
	}
}

func (s *Server) handleConn(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	channelID := r.PathValue("channel_id")
	if channelID == "" {
		return
	}
	slog.Info("New connection in room " + channelID)
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	s.dispatcher.RegisterClient(&dom.Client{ChannelID: channelID, Name: name, Ws: ws})
}
