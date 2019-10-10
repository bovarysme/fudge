package main

//go:generate go run generate.go

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"fudge/config"
	"fudge/handler"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "config.yml", "path to the config file")
	flag.Parse()
}

func main() {
	cfg, err := config.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	h, err := handler.NewHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)

	server := &http.Server{
		Addr:         "localhost:8080",
		Handler:      h.Router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     logger,
	}

	logger.Println("Starting server on", server.Addr)
	log.Fatal(server.ListenAndServe())
}
