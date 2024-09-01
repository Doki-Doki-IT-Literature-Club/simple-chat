package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/Doki-Doki-IT-Literature-Club/simple-chat/internal/shared/domain"
	"github.com/jackc/pgx/v5"
)

type PGMessageRepository struct {
	conn *pgx.Conn
}

func NewPGMessageRepository(DBUser, DBPassword, DBHost, DBName string) (*PGMessageRepository, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		DBUser,
		DBPassword,
		DBHost,
		DBName,
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

	repo := &PGMessageRepository{conn: conn}
	err = repo.createMessageTable()
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *PGMessageRepository) Close(ctx context.Context) error {
	return r.conn.Close(ctx)
}

func (r *PGMessageRepository) createMessageTable() error {
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

func (r *PGMessageRepository) InsertMessage(m domain.Message) error {
	_, err := r.conn.Exec(context.Background(), `
		INSERT INTO messages (id, content, created_at, author_id, channel_id)
		VALUES ($1, $2, $3, $4, $5)
	`, m.ID, m.Content, m.CreatedAt, m.AuthorID, m.ChannelID)

	return err
}

func (r *PGMessageRepository) GetMessages(channelID string) ([]domain.Message, error) {
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

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		err := rows.Scan(&m.ID, &m.Content, &m.CreatedAt, &m.AuthorID, &m.ChannelID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}

	return messages, nil
}
