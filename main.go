package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
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

func main() {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	cfg := &apiConfig{}

	apiRouter := http.NewServeMux()
	apiRouter.HandleFunc("/healthz", handleHealthz)

	adminRouter := http.NewServeMux()
	adminRouter.HandleFunc("/metrics", cfg.handleMetrics)
	adminRouter.HandleFunc("/reset", cfg.handleReset)

	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.Handle("/api/", http.StripPrefix("/api", apiRouter))
	mux.Handle("/admin/", http.StripPrefix("/admin", adminRouter))

	fmt.Printf("Listening on port: %v\n", server.Addr)
	err := server.ListenAndServe()
	fmt.Printf("Error during listen and serve: %v\n", err)
}
