package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"fudge/config"
)

func main() {
	cfg, err := config.NewConfig("config.yml")
	if err != nil {
		log.Fatal(err)
	}

	handler, err := NewHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	server := &http.Server{
		Addr:         "localhost:8080",
		Handler:      handler.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     logger,
	}

	logger.Println("Starting server on", server.Addr)
	log.Fatal(server.ListenAndServe())
}
