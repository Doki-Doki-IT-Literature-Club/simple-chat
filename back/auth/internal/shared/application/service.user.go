package application

import (
	"log"
	"time"

	"github.com/Doki-Doki-IT-Literature-Club/auth/internal/shared/domain"
)

type UserService struct {
	userRepo domain.UserRepository
	secret   string
}

func NewUserService(userRepo domain.UserRepository, secret string) *UserService {
	return &UserService{userRepo: userRepo, secret: secret}
}

func (s *UserService) GetOrCreateUser(email string) (domain.User, error) {
	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		user = domain.NewUser(email)
		err = s.userRepo.InsertUser(user)
	}
	return user, err
}

func (s *UserService) GetUserByEmail(email string) (domain.User, error) {
	return s.userRepo.GetUserByEmail(email)
}

func (s *UserService) InsertUser(user domain.User) error {
	return s.userRepo.InsertUser(user)
}

func (s *UserService) CreateUserJWTToken(id, email string) (string, error) {
	token, err := CreateUserJWTToken(id, email, time.Now().Add(time.Hour), s.secret)
	return token, err
}

func (s *UserService) ParseUserJWT(tokenString string) (*UserClaims, error) {
	claims, err := ParseUserJWT(tokenString, s.secret)
	if err != nil {
		log.Printf("Failed to parse token: %v", err)
		return nil, err
	}

	return claims, nil
}
