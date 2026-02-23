package handler

import (
	"encoding/json"
	"net/http"
	"log"

	"grysha11/httpServersGo/internal/auth"
	"grysha11/httpServersGo/internal/database"
	"github.com/google/uuid"
)

func (h *Handler) PolkaWebhook(w http.ResponseWriter, r *http.Request) {
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil || apiKey != h.Cfg.APIKey {
		RespondWithString(w, 401, "Invalid or missing API key")
		return
	}

	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Decode error: %v", err)
		RespondWithString(w, 500, "Decode error")
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}

	_, err = h.Cfg.DB.UpgradeUserChirpyRedByID(r.Context(), database.UpgradeUserChirpyRedByIDParams{
		ID:          params.Data.UserID,
		IsChirpyRed: true,
	})
	if err != nil {
		RespondWithString(w, 404, "User not found")
		return
	}
	w.WriteHeader(204)
}
