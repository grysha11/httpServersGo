package handler

import (
	"fmt"
	"log"
	"net/http"
)

func (h *Handler) Metrics(w http.ResponseWriter, r *http.Request) {
	numOfReq := int64(h.Cfg.FileserverHits.Load())
	bodyHtml := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, numOfReq)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(bodyHtml))	
}

func (h *Handler) Reset(w http.ResponseWriter, r *http.Request) {
	if h.Cfg.Platform != "dev" {
		RespondWithString(w, 403, "FAIL")
		return
	}

	h.Cfg.FileserverHits.Store(0)

	if err := h.Cfg.DB.DeleteUsers(r.Context()); err != nil {
		log.Printf("Error deleting users from DB: %v", err)
	}
	if err := h.Cfg.DB.DeleteChirps(r.Context()); err != nil {
		log.Printf("Error deleting chirps from DB: %v", err)
	}
	if err := h.Cfg.DB.DeleteRefreshTokens(r.Context()); err != nil {
		log.Printf("Error deleting refresh tokens from DB: %v", err)
	}

	RespondWithString(w, 200, "OK")
}