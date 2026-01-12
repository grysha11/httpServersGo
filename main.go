package main

import (
	"fmt"
	"net/http"
	"strconv"
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
	numOfReq := "Hits: " + strconv.FormatInt(int64(cfg.fileserverHits.Load()), 10)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(numOfReq))
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
	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", handleHealthz)
	mux.HandleFunc("/metrics", cfg.handleMetrics)
	mux.HandleFunc("/reset", cfg.handleReset)
	err := server.ListenAndServe()
	fmt.Printf("Error during listen and serve: %v\n", err)
}
