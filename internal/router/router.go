package router

import (
	"net/http"

	"grysha11/httpServersGo/internal/config"
	"grysha11/httpServersGo/internal/handler"
	"grysha11/httpServersGo/internal/middleware"
)

func NewRouter(cfg *config.AppConfig, h *handler.Handler) *http.ServeMux {
	mux := http.NewServeMux()

	apiRouter := http.NewServeMux()
	apiRouter.HandleFunc("GET /healthz", h.Healthz)
	
	apiRouter.HandleFunc("POST /users", h.CreateUser)
	apiRouter.HandleFunc("PUT /users", middleware.RequireJWT(cfg, h.PutUser))
	
	apiRouter.HandleFunc("POST /chirps", middleware.RequireJWT(cfg, h.CreateChirp))
	apiRouter.HandleFunc("GET /chirps", h.GetChirps)
	apiRouter.HandleFunc("GET /chirps/{chirpID}", h.GetChirpByID)
	apiRouter.HandleFunc("DELETE /chirps/{chirpID}", middleware.RequireJWT(cfg, h.DeleteChirpByID))
	
	apiRouter.HandleFunc("POST /login", h.Login)
	apiRouter.HandleFunc("POST /refresh", h.Refresh)
	apiRouter.HandleFunc("POST /revoke", h.Revoke)
	
	apiRouter.HandleFunc("POST /polka/webhooks", h.PolkaWebhook)

	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("GET /metrics", h.Metrics)
	adminRouter.HandleFunc("POST /reset", h.Reset)

	mux.Handle("GET /app/", http.StripPrefix("/app", middleware.MetricsInc(cfg, http.FileServer(http.Dir(".")))))
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mux.Handle("/admin/", http.StripPrefix("/admin", adminRouter))

	return mux
}
