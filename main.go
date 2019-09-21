package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}

func main() {
	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/{repository}", homeHandler)
	router.HandleFunc("/{repository}/commits/{branch}", homeHandler)
	router.HandleFunc("/{repository}/tree/{commit}/{path:.*}", homeHandler)
	router.HandleFunc("/{repository}/blob/{commit}/{path:.*}", homeHandler)
	router.HandleFunc("/{repository}/raw/{commit}/{path:.*}", homeHandler)

	log.Fatal(http.ListenAndServe("localhost:8080", router))
}
