package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"grysha11/httpServersGo/internal/config"
	"grysha11/httpServersGo/internal/database"
	"grysha11/httpServersGo/internal/handler"
	"grysha11/httpServersGo/internal/router"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error connecting to db: %v\n", err)
	}

	appCfg := &config.AppConfig{
		DB:        database.New(db),
		Platform:  os.Getenv("PLATFORM"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		APIKey:    os.Getenv("POLKA_KEY"),
	}

	h := &handler.Handler{Cfg: appCfg}

	mux := router.NewRouter(appCfg, h)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Printf("Listening on port: %v\n", server.Addr)
	log.Fatal(server.ListenAndServe())
}
