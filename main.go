package main

import (
	"net/http"
	"fmt"
)

func main() {
	mux := http.NewServeMux()
	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = mux
	err := server.ListenAndServe()
	fmt.Printf("Error during listen and serve: %v\n", err)
}