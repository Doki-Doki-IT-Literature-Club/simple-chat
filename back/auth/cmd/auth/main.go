package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
)

var (
	clientID     = os.Getenv("AUTH_CLIENT_ID")
	clientSecret = os.Getenv("AUTH_CLIENT_SECRET")
	redirectURL  = "http://localhost:5556/auth/google/callback"
	providerURL  = "https://accounts.google.com"
)

type Claims struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setCallbackCookie(w http.ResponseWriter, r *http.Request, name, value string) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   r.TLS != nil,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, c)
}

func main() {
	ctx := context.Background()

	con, err := pgx.Connect(ctx, dbConString())
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return
	}
	defer con.Close(ctx)

	err = createUserTableIfNotExists(con)
	if err != nil {
		log.Fatalf("Failed to create user table: %v", err)
		return
	}

	provider, err := oidc.NewProvider(ctx, providerURL)
	if err != nil {
		log.Fatal(err)
	}
	oidcConfig := &oidc.Config{
		ClientID: clientID,
	}
	verifier := provider.Verifier(oidcConfig)

	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("token")
		if err == nil {

			_, err := parseJWT(token.Value)
			if err != nil {
				http.Error(w, "Failed to parse JWT token", http.StatusInternalServerError)
				return
			}

			proxyHandler(w, r)

			return
		}

		state, err := randString(16)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		nonce, err := randString(16)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}
		setCallbackCookie(w, r, "state", state)
		setCallbackCookie(w, r, "nonce", nonce)

		http.Redirect(w, r, config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
	})

	http.HandleFunc("/auth/google/callback", func(w http.ResponseWriter, r *http.Request) {
		state, err := r.Cookie("state")
		if err != nil {
			http.Error(w, "state not found", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("state") != state.Value {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			http.Error(w, "No id_token field in oauth2 token.", http.StatusInternalServerError)
			return
		}
		idToken, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "Failed to verify ID Token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		nonce, err := r.Cookie("nonce")
		if err != nil {
			http.Error(w, "nonce not found", http.StatusBadRequest)
			return
		}
		if idToken.Nonce != nonce.Value {
			http.Error(w, "nonce did not match", http.StatusBadRequest)
			return
		}

		var claims struct {
			Email string `json:"email"`
		}

		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userId, err := getUserIdByEmail(con, claims.Email)
		if err != nil {
			uuid, err := uuid.NewRandom()
			if err != nil {
				http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
				return
			}

			err = insertUser(con, uuid.String(), claims.Email)
			if err != nil {
				http.Error(w, "Failed to insert user", http.StatusInternalServerError)
				return
			}

			userId = uuid.String()
		}

		token, err := createJWTToken(userId, claims.Email, time.Now().Add(time.Hour))
		if err != nil {
			http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
			return
		}

		setCallbackCookie(w, r, "token", token)

		w.WriteHeader(http.StatusOK)
	})

	port := ":5556"
	log.Printf("Listening on %s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func parseJWT(tokenString string) (*Claims, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func createJWTToken(id, email string, exp time.Time) (string, error) {
	claims := Claims{
		id,
		email,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	return tokenString, err
}

func dbConString() string {
	DBUser := os.Getenv("DB_USER")
	DBPassword := os.Getenv("DB_PASSWORD")
	DBHost := os.Getenv("DB_HOST")
	DBName := os.Getenv("DB_NAME")

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s",
		DBUser,
		DBPassword,
		DBHost,
		DBName,
	)

	return connString
}

func createUserTableIfNotExists(con *pgx.Conn) error {
	_, err := con.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT,
			created_at TIMESTAMP
		)
	`)

	return err
}

func getUserIdByEmail(con *pgx.Conn, email string) (string, error) {
	var id string
	err := con.QueryRow(context.Background(), `
		SELECT id FROM users WHERE email = $1
	`, email).Scan(&id)

	return id, err
}

func insertUser(con *pgx.Conn, id, email string) error {
	_, err := con.Exec(context.Background(), `
		INSERT INTO users (id, email, created_at)
		VALUES ($1, $2, $3)
	`, id, email, time.Now())

	return err
}

func proxyWebSocket(w http.ResponseWriter, req *http.Request) {
	targetURL := "ws://api-gateway" + req.URL.Path
	target, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}

	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")

	proxy.ServeHTTP(w, req)
}

func proxyHandler(w http.ResponseWriter, req *http.Request) {
	if req.Header.Get("Upgrade") == "websocket" {
		proxyWebSocket(w, req)
		return
	}

	targetURL := "http://api-gateway" + req.URL.Path + "?" + req.URL.RawQuery
	proxyReq, err := http.NewRequest(req.Method, targetURL, req.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	proxyReq.Header = req.Header

	transport := http.DefaultTransport
	resp, err := transport.RoundTrip(proxyReq)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
