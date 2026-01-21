package main

import (
	"encoding/json"
	"fmt"
	"grysha11/httpServersGo/internal/database"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"database/sql"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits	atomic.Int32
	db				*database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
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
	numOfReq := int64(cfg.fileserverHits.Load())
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
	cfg.fileserverHits.Store(0)
	bodyText := "RESET"

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

func handleValidateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body	string	`json:"body"`
	}

	type responseBody struct {
		ErrorStr	string	`json:"error"`
		Valid		bool	`json:"valid"`
		CleanedBody	string	`json:"cleaned_body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respErorr := responseBody{
			ErrorStr: err.Error(),
			Valid: false,
			CleanedBody: "",
		}
		data, _ := json.Marshal(respErorr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		w.Write(data)
		return
	}
	if len(params.Body) > 140 {
		respErorr := responseBody{
			ErrorStr: "chirp is too long",
			Valid: false,
			CleanedBody: "",
		}
		data, _ := json.Marshal(respErorr)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		w.Write(data)
		return
	}
	respSuccess := responseBody{
		ErrorStr: "",
		Valid: true,
		CleanedBody: formatBody(params.Body),
	}
	data, err := json.Marshal(respSuccess)
	if err != nil {
		log.Printf("Error marshaling data: %v", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
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
	cfg := &apiConfig{
		db: dbQueries,
	}

	apiRouter := http.NewServeMux()
	apiRouter.HandleFunc("/healthz", handleHealthz)
	apiRouter.HandleFunc("/validate_chirp", handleValidateChirp)

	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("/metrics", cfg.handleMetrics)
	adminRouter.HandleFunc("/reset", cfg.handleReset)

	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mux.Handle("/admin/", http.StripPrefix("/admin", adminRouter))

	log.Printf("Listening on port: %v\n", server.Addr)
	err = server.ListenAndServe()
	log.Printf("Error during listen and serve: %v\n", err)
}
