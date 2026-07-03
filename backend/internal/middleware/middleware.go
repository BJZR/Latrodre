package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"latrode-fusion/internal/models"
	"latrode-fusion/internal/repository"
)

type contextKey string

const userKey contextKey = "user"
const sessionTokenKey contextKey = "session_token"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-Id")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func JSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func Auth(userRepo *repository.UserRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			c, err := r.Cookie("session_id")
			if err == nil {
				token = c.Value
			}
			if token == "" {
				token = r.Header.Get("X-Session-Id")
			}

			if token == "" {
				http.Error(w, `{"error":"no autorizado"}`, http.StatusUnauthorized)
				return
			}

			session, err := userRepo.FindSession(token)
			if err != nil {
				http.Error(w, `{"error":"sesión inválida"}`, http.StatusUnauthorized)
				return
			}

			user, err := userRepo.FindByID(session.UserID)
			if err != nil {
				http.Error(w, `{"error":"usuario no encontrado"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userKey, user)
			ctx = context.WithValue(ctx, sessionTokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userKey).(*models.User)
		if !ok || user.Role != "admin" {
			http.Error(w, `{"error":"acceso denegado"}`, http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func OptionalAuth(userRepo *repository.UserRepo) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			c, err := r.Cookie("session_id")
			if err == nil {
				token = c.Value
			}
			if token == "" {
				token = r.Header.Get("X-Session-Id")
			}

			if token != "" {
				session, err := userRepo.FindSession(token)
				if err == nil {
					user, err := userRepo.FindByID(session.UserID)
					if err == nil {
						ctx := context.WithValue(r.Context(), userKey, user)
						ctx = context.WithValue(ctx, sessionTokenKey, token)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUser(r *http.Request) *models.User {
	user, ok := r.Context().Value(userKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

func GetSessionToken(r *http.Request) string {
	token, _ := r.Context().Value(sessionTokenKey).(string)
	return token
}

func GenerateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type APIError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func GetClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	if parts := strings.Split(r.RemoteAddr, ":"); len(parts) > 0 {
		return parts[0]
	}
	return r.RemoteAddr
}
