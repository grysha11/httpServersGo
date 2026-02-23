package handler

import (
	"encoding/json"
	"grysha11/httpServersGo/internal/auth"
	"grysha11/httpServersGo/internal/database"
	"log"
	"net/http"
	"time"
)

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email				string	`json:"email"`
		Password			string	`json:"password"`
		ExpiresInSeconds	int		`json:"expires_in_seconds"`
	}

	params := parameters{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		log.Printf("Decode error: %v", err)
		RespondWithString(w, 500, "Decode error")
		return
	}
	
	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = 60 * 60
	}

	user, err := h.Cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		RespondWithString(w, 404, "User not found")
		return
	}

	isCorrect, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !isCorrect {
		RespondWithString(w, 401, "Password is incorrect")
		return
	}

	token, err := auth.MakeJWT(user.ID, h.Cfg.JWTSecret, time.Second * time.Duration(params.ExpiresInSeconds))
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		RespondWithString(w, 500, "Internal Server Error")
		return	
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		RespondWithJSON(w, 500, "Internal Server Error")
		return
	}

	_, err = h.Cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		log.Printf("Database error Create Refresh Token: %v", err)
		RespondWithString(w, 500, "Database error")
		return
	}

	RespondWithJSON(w, 200, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChirpyRed:  user.IsChirpyRed,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithString(w, 401, "No token provided")
		return
	}

	user, err := h.Cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		RespondWithString(w, 401, "Invalid or expired token")
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, h.Cfg.JWTSecret, time.Hour)
	if err != nil {
		log.Printf("Error creating JWT token: %v", err)
		RespondWithString(w, 500, "Error creating token")
		return
	}

	RespondWithJSON(w, 200, map[string]string{"token": accessToken})
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		RespondWithString(w, 401, "No token provided")
		return
	}

	if err = h.Cfg.DB.RevokeRefreshToken(r.Context(), refreshToken); err != nil {
		RespondWithString(w, 401, "Invalid token")
		return
	}

	w.WriteHeader(204)
}
