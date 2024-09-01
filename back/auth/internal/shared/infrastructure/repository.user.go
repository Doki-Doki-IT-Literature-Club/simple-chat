package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/LeperGnome/auth/internal/shared/domain"
	"github.com/jackc/pgx/v5"
)

type PGUserRepository struct {
	conn *pgx.Conn
}

func NewPGUserRepository(DBUser, DBPassword, DBHost, DBName string) *PGUserRepository {
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

	repo := &PGUserRepository{conn: conn}
	err = repo.createUserTable()
	if err != nil {
		log.Fatal(err)
	}
	return repo
}

func (r *PGUserRepository) createUserTable() error {
	_, err := r.conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT,
			created_at TIMESTAMP
		)
	`)

	return err
}

func (r *PGUserRepository) Close(ctx context.Context) error {
	return r.conn.Close(ctx)
}

func (r *PGUserRepository) GetUserByEmail(email string) (domain.User, error) {
	var user domain.User
	err := r.conn.QueryRow(context.Background(), `
		SELECT id, email
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email)

	return user, err
}

func (r *PGUserRepository) InsertUser(user domain.User) error {
	_, err := r.conn.Exec(context.Background(), `
		INSERT INTO users (id, email, created_at)
		VALUES ($1, $2, $3)
	`, user.ID, user.Email, time.Now())

	return err
}
