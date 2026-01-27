package main

import (
	"encoding/json"
	"fmt"
	"grysha11/httpServersGo/internal/database"
	"grysha11/httpServersGo/internal/auth"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"database/sql"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID	`json:"id"`
	CreatedAt time.Time	`json:"created_at"`
	UpdatedAt time.Time	`json:"updated_at"`
	Body      string	`json:"body"`
	UserID    uuid.UUID	`json:"user_id"`
}

type User struct {
		ID				uuid.UUID	`json:"id"`
		CreatedAt		time.Time	`json:"created_at"`
		UpdatedAt		time.Time	`json:"updated_at"`
		Email			string		`json:"email"`
		Token			string		`json:"token"`
		RefreshToken	string		`json:"refresh_token"`
}

type apiConfig struct {
	FileserverHits	atomic.Int32
	DB				*database.Queries
	Platform		string
	JWTSecret		string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	bodyText := "OK"

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(bodyText))
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	numOfReq := int64(cfg.FileserverHits.Load())
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

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	bodyText := "RESET"
	bodyFail := "FAIL"
	if strings.Compare(cfg.Platform, "dev") != 0 {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(403)
		w.Write([]byte(bodyFail))
		return
	}
	
	cfg.FileserverHits.Store(0)
	err := cfg.DB.DeleteUsers(r.Context())
	if err != nil {
		log.Printf("Couldn't execute db query: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(bodyFail))
		return
	}
	err = cfg.DB.DeleteChirps(r.Context())
	if err != nil {
		log.Printf("Couldn't execute db query: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(bodyFail))
		return
	}
	err = cfg.DB.DeleteRefreshTokens(r.Context())
	if err != nil {
		log.Printf("Couldn't execute db query: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(bodyFail))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(bodyText))
}

func formatBody(bodyStr string) string {
	words := strings.Split(bodyStr, " ")

	for i, word := range words {
		word = strings.ToLower(word)
		if word == "kerfuffle" || word == "sharbert" || word == "fornax" {
			words[i] = "****"
		}
	}

	res := strings.Join(words, " ")

	return res
}

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email			string	`json:"email"`
		Password		string	`json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while decoding request: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while hashing password: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	user, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	respSuccess := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) handleCreateChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body	string		`json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while validating authentication: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.JWTSecret)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while validating authentication: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while decoding request: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	if len(params.Body) == 0 {
		errorStr := "Error body is null"
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(errorStr))
		return
	}

	chirp, err := cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: params.Body,
		UserID: userID,
	})
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	respSuccess := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(data)
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	respSuccess := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		respSuccess[i] = Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		}
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while parsing uuid: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	chirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		errorStr := fmt.Sprintf("Error: Couldn't find chirp with id: %v", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(404)
		w.Write([]byte(errorStr))
		return
	}

	respSuccess := Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body: chirp.Body,
		UserID: chirp.UserID,
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email				string	`json:"email"`
		Password			string	`json:"password"`
		ExpiresInSeconds	int		`json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while decoding request: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	if params.ExpiresInSeconds == 0 {
		params.ExpiresInSeconds = 60 * 60
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(404)
		w.Write([]byte(errorStr))
		return
	}

	isCorrect, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making dehashing password: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	if isCorrect == false {
		errorStr := "Password is incorrect"
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Second * time.Duration(params.ExpiresInSeconds))
	if err != nil {
		log.Printf("Error creating JWT token: %v\n", err)
		w.WriteHeader(500)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v\n", err)
		w.WriteHeader(500)
		return
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token: refreshToken,
		UserID: user.ID,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
	})
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	respSuccess := User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
		Token: token,
		RefreshToken: refreshToken,
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	type ResponseSuccess struct {
		Token	string	`json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while getting token: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Hour)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while creating access token: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	respSuccess := ResponseSuccess{
		Token: accessToken,
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v\n", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while getting token: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	w.WriteHeader(204)
}

func (cfg *apiConfig) handlePutUsers(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email		string	`json:"email"`
		Password	string	`json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while decoding request: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while getting token: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.JWTSecret)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while validating token: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	passwordHash, err := auth.HashPassword(params.Password)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while hashing password: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	user, err := cfg.DB.UpdateUserByID(r.Context(), database.UpdateUserByIDParams{
		ID: userID,
		Email: params.Email,
		HashedPassword: passwordHash,
	})
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(404)
		w.Write([]byte(errorStr))
		return
	}

	respSuccess := User{
		ID: userID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func (cfg *apiConfig) handleDeleteChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while parsing uuid: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while getting token: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.JWTSecret)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while validating token: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(401)
		w.Write([]byte(errorStr))
		return
	}

	chirp, err := cfg.DB.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(404)
		w.Write([]byte(errorStr))
		return
	}

	if chirp.UserID != userID {
		errorStr := "Error, you are not a creator of a chirp."
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(403)
		w.Write([]byte(errorStr))
		return
	}

	err = cfg.DB.DeleteChirpByID(r.Context(), chirpID)
	if err != nil {
		errorStr := fmt.Sprintf("Error occured while making db call: %v\n", err)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(errorStr))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(204)
}

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Printf("Error connecting to db: %v\n", err)
		return
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}

	platform := os.Getenv("PLATFORM")
	jwtSecret := os.Getenv("JWT_SECRET")
	cfg := &apiConfig{
		DB: dbQueries,
		Platform: platform,
		JWTSecret: jwtSecret,
	}

	apiRouter := http.NewServeMux()
	apiRouter.HandleFunc("GET /healthz", handleHealthz)
	apiRouter.HandleFunc("POST /users", cfg.handleUsers)
	apiRouter.HandleFunc("POST /chirps", cfg.handleCreateChirps)
	apiRouter.HandleFunc("GET /chirps", cfg.handleGetChirps)
	apiRouter.HandleFunc("GET /chirps/{chirpID}", cfg.handleGetChirpByID)
	apiRouter.HandleFunc("POST /login", cfg.handleLogin)
	apiRouter.HandleFunc("POST /refresh", cfg.handleRefresh)
	apiRouter.HandleFunc("POST /revoke", cfg.handleRevoke)
	apiRouter.HandleFunc("PUT /users", cfg.handlePutUsers)
	apiRouter.HandleFunc("DELETE /chirps/{chirpID}", cfg.handleDeleteChirpByID)

	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("GET /metrics", cfg.handleMetrics)
	adminRouter.HandleFunc("POST /reset", cfg.handleReset)

	mux.Handle("GET /app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mux.Handle("/admin/", http.StripPrefix("/admin", adminRouter))

	log.Printf("Listening on port: %v\n", server.Addr)
	err = server.ListenAndServe()
	log.Printf("Error during listen and serve: %v\n", err)
}
