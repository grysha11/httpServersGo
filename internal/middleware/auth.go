package middleware

import (
	"context"
	"fmt"
	"grysha11/httpServersGo/internal/auth"
	"grysha11/httpServersGo/internal/config"
	"net/http"
)

type contextKey string
const UserIDKey contextKey = "userID"

func RequireJWT(cfg *config.AppConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, fmt.Sprintf("Missing auth: %v", err), http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid auth: %v", err), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
