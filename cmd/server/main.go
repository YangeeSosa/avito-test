package main

import (
	"log"
	"net/http"
	"os"

	"github.com/test-avito/internal/app"
)

func main() {
	application := app.New()
	addr := getAddr()
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, application.Handler()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func getAddr() string {
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		return v
	}
	return ":8080"
}
