package handler

import (
	"encoding/json"
	"net/http"
	"log"

	"grysha11/httpServersGo/internal/auth"
	"grysha11/httpServersGo/internal/database"
	"grysha11/httpServersGo/internal/middleware"
	"github.com/google/uuid"
)

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Decode error: %v", err)
		RespondWithString(w, 500, "Decode error")
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Hashing error: %v", err)
		RespondWithString(w, 500, "Hashing error")
		return
	}

	user, err := h.Cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		log.Printf("Database error Create User: %v", err)
		RespondWithString(w, 500, "Database error")
		return
	}

	RespondWithJSON(w, 201, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (h *Handler) PutUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)

	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Decode error: %v", err)
		RespondWithString(w, 500, "Decode error")
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		log.Printf("Hashing error: %v", err)
		RespondWithString(w, 500, "Hashing error")
		return
	}

	user, err := h.Cfg.DB.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		log.Printf("Database error Update User By ID: %v", err)
		RespondWithString(w, 404, "Database error")
		return
	}

	RespondWithJSON(w, 200, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}
