package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/LeperGnome/auth/internal/shared/application"
	"github.com/LeperGnome/auth/internal/shared/infrastructure"
	"github.com/LeperGnome/auth/internal/shared/presentation"
)

var (
	port         = os.Getenv("PORT")
	clientID     = os.Getenv("AUTH_CLIENT_ID")
	clientSecret = os.Getenv("AUTH_CLIENT_SECRET")
	tokenSecret  = os.Getenv("JWT_SECRET")
	redirectURL  = fmt.Sprintf("http://localhost:%s/auth/google/callback", port)
	providerURL  = "https://accounts.google.com"
)

func main() {
	dbConfig := infrastructure.GetPGConfig()

	userRepo := infrastructure.NewPGUserRepository(
		dbConfig.DBUser,
		dbConfig.DBPassword,
		dbConfig.DBHost,
		dbConfig.DBName,
	)

	userService := application.NewUserService(userRepo, tokenSecret)
	googleAuthService := application.NewGoogleAuth(clientID, clientSecret, providerURL, redirectURL)

	mux := http.NewServeMux()
	mux.Handle("/", presentation.NewGoogleAuthInitHandler(userService, googleAuthService))
	mux.Handle("/auth/google/callback", presentation.NewGoogleAuthCallbackHandler(userService, googleAuthService))
	mux.HandleFunc("/health", presentation.Health)

	log.Printf("Listening on %s", port)
	err := http.ListenAndServe(":"+port, presentation.LoggerMiddleware(presentation.NewProxyMiddleware(mux, userService)))
	if err != nil {
		log.Fatal(err)
	}
}
