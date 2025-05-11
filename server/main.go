package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Serve static files from frontend/dist
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("../client/vite-project/dist/")))

	// Optional API routes
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server running..."))
	}).Methods("GET")

	log.Println("Server running on http://localhost:8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}
