package handler

import (
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"log"

	"grysha11/httpServersGo/internal/database"
	"grysha11/httpServersGo/internal/middleware"
	"github.com/google/uuid"
)

func formatBody(bodyStr string) string {
	words := strings.Split(bodyStr, " ")
	for i, word := range words {
		lower := strings.ToLower(word)
		if lower == "kerfuffle" || lower == "sharbert" || lower == "fornax" {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func (h *Handler) CreateChirp(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	type parameters struct {
		Body string `json:"body"`
	}
	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Decode error: %v", err)
		RespondWithString(w, 500, "Decode error")
		return
	}
	if len(params.Body) == 0 {
		RespondWithString(w, 400, "Body is null")
		return
	}

	chirp, err := h.Cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   formatBody(params.Body),
		UserID: userID,
	})
	if err != nil {
		log.Printf("Database error Create Chirp: %v", err)
		RespondWithString(w, 500, "Database error")
		return
	}

	RespondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (h *Handler) GetChirps(w http.ResponseWriter, r *http.Request) {
	authorIDString := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error

	if authorIDString != "" {
		authorID, parseErr := uuid.Parse(authorIDString)
		if parseErr != nil {
			RespondWithString(w, 400, "Invalid author ID")
			return
		}
		chirps, err = h.Cfg.DB.GetChirpsByAuthor(r.Context(), authorID)
	} else {
		chirps, err = h.Cfg.DB.GetAllChirps(r.Context())
	}

	if err != nil {
		log.Printf("Database error: %v", err)
		RespondWithString(w, 500, "Database error")
		return
	}

	resp := make([]Chirp, len(chirps))
	for i, c := range chirps {
		resp[i] = Chirp{ID: c.ID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, Body: c.Body, UserID: c.UserID}
	}

	if sortParam == "desc" {
		slices.SortFunc(resp, func(a, b Chirp) int {
			return b.CreatedAt.Compare(a.CreatedAt)
		})
	}

	RespondWithJSON(w, 200, resp)
}

func (h *Handler) GetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("UUID error: %v", err)
		RespondWithString(w, 500, "Invalid UUID")
		return
	}

	chirp, err := h.Cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		RespondWithString(w, 404, "Chirp not found")
		return
	}

	RespondWithJSON(w, 200, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (h *Handler) DeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("UUID error: %v", err)
		RespondWithString(w, 500, "Invalid UUID")
		return
	}

	chirp, err := h.Cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		RespondWithString(w, 404, "Chirp not found")
		return
	}

	if chirp.UserID != userID {
		RespondWithString(w, 403, "Not authorized to delete this chirp")
		return
	}

	if err = h.Cfg.DB.DeleteChirpByID(r.Context(), chirpID); err != nil {
		log.Printf("Database error Delete Chirp By ID: %v", err)
		RespondWithString(w, 500, "Database error")
		return
	}
	w.WriteHeader(204)
}
