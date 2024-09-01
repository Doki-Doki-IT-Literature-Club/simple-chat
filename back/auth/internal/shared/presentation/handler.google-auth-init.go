package presentation

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/LeperGnome/auth/internal/shared/application"
)

func NewGoogleAuthInitHandler(userService *application.UserService, googleAuthService *application.GoogleAuth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		http.Redirect(w, r, googleAuthService.AuthCodeURL(state, nonce), http.StatusFound)
	})
}

func randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
