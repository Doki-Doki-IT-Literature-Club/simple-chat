package presentation

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/LeperGnome/auth/internal/shared/application"
)

func NewProxyMiddleware(next http.Handler, userService *application.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("token")
		if err == nil {
			_, err := userService.ParseUserJWT(token.Value)
			if err == nil {
				proxyHandler(w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
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
