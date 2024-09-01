package presentation

import (
	"net/http"
	"time"

	"github.com/Doki-Doki-IT-Literature-Club/auth/internal/shared/application"
)

func NewGoogleAuthCallbackHandler(userService *application.UserService, googleAuthService *application.GoogleAuth) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		state, err := r.Cookie("state")
		if err != nil {
			http.Error(w, "state not found", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("state") != state.Value {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		idToken, err := googleAuthService.Verify(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
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

		user, err := userService.GetOrCreateUser(claims.Email)
		if err != nil {
			http.Error(w, "Failed to get or create user", http.StatusInternalServerError)
			return
		}

		token, err := userService.CreateUserJWTToken(user.ID, user.Email)
		if err != nil {
			http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
			return
		}

		setCallbackCookie(w, r, "token", token)

		http.Redirect(w, r, "/", http.StatusFound)
	})
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
