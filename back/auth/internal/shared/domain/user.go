package domain

import (
	"github.com/google/uuid"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewUser(email string) User {
	return User{
		ID:    uuid.New().String(),
		Email: email,
	}
}

type UserRepository interface {
	InsertUser(u User) error
	GetUserByEmail(email string) (User, error)
}
