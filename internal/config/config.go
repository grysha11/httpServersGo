package config

import (
	"sync/atomic"
	"grysha11/httpServersGo/internal/database"
)

type AppConfig struct {
	FileserverHits	atomic.Int32
	DB				*database.Queries
	Platform		string
	JWTSecret		string
	APIKey			string
}
