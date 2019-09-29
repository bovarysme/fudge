package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	h, err := NewHandler()
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	server := &http.Server{
		Addr:         "localhost:8080",
		Handler:      h.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     logger,
	}

	logger.Println("Starting server on", server.Addr)
	log.Fatal(server.ListenAndServe())
}
