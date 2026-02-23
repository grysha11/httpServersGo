package handler

import (
	"encoding/json"
	"log"
	"grysha11/httpServersGo/internal/config"
	"net/http"
)

type Handler struct {
	Cfg *config.AppConfig
}

func RespondWithString(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(message))
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(data)
}

func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	RespondWithString(w, 200, "ok")
}
