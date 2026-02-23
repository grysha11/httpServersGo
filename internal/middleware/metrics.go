package middleware

import (
	"net/http"
	"grysha11/httpServersGo/internal/config"
)

func MetricsInc(cfg *config.AppConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}
