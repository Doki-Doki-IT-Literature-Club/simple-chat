package chronicler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	conn *pgx.Conn
}

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBName     string
}

func NewRepository(config Config) (*Repository, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		config.DBUser,
		config.DBPassword,
		config.DBHost,
		config.DBName,
	)

	var conn *pgx.Conn
	var err error

	for conn == nil || err != nil {
		conn, err = pgx.Connect(context.Background(), connString)
		if err != nil {
			fmt.Println("Failed to connect to the database, retrying in 5 seconds")
			time.Sleep(5 * time.Second)
		}
	}

	return &Repository{conn: conn}, nil
}

func (r *Repository) Close() error {
	return r.conn.Close(context.Background())
}

func (r *Repository) CreateMessageTable() error {
	_, err := r.conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY,
			content TEXT,
			created_at TIMESTAMPTZ,
			author_id TEXT,
			channel_id TEXT
		)
	`)

	return err
}

type Message struct {
	ID        string
	Content   string
	CreatedAt time.Time
	AuthorID  string
	ChannelID string
}

// it shouldn't be here
func NewMessage(content string, authorID string, channelID string) Message {
	return Message{
		ID:        uuid.New().String(),
		Content:   content,
		CreatedAt: time.Now(),
		AuthorID:  authorID,
		ChannelID: channelID,
	}
}

func (r *Repository) InsertMessage(m Message) error {
	_, err := r.conn.Exec(context.Background(), `
		INSERT INTO messages (id, content, created_at, author_id, channel_id)
		VALUES ($1, $2, $3, $4, $5)
	`, m.ID, m.Content, m.CreatedAt, m.AuthorID, m.ChannelID)

	return err
}

func (r *Repository) GetMessages(channelID string) ([]Message, error) {
	rows, err := r.conn.Query(context.Background(), `
		SELECT id, content, created_at, author_id, channel_id
		FROM messages
		WHERE channel_id = $1
		ORDER BY created_at DESC
	`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &m.AuthorID, &m.ChannelID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}
