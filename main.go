package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"shortener/store"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func main() {
	// Get configuration
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("failed to convert SHORT_URL_LENGTH to int: %v", err)
	}
	dbUrl := os.Getenv("DB_URL")
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("COLLECTION_NAME")
	shortUrlLengthStr := os.Getenv("SHORT_URL_LENGTH")
	shortUrlLength, err := strconv.Atoi(shortUrlLengthStr)
	if err != nil {
		log.Fatalf("failed to convert SHORT_URL_LENGTH to int: %v", err)
	}

	// Build store
	store, err := store.NewStore(dbUrl, dbName, collectionName)
	if err != nil {
		log.Fatalf("error creating store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("error closing store: %v", err)
		}
	}()

	// Build and start api
	api := NewApi(store, []byte("abcdefghijklmnopqrstuvwxyz0123456789"), shortUrlLength)
	if err := start(port, api); err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

func start(port int, api *Api) error {
	r := chi.NewRouter()

	// Middleware
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Routes
	r.Post("/", api.create)
	r.Delete("/{url}", api.delete)
	r.Get("/{url}", api.get)
	r.Get("/{url}/stats", api.getStats)

	// Start server
	log.Printf("server started on :%d", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
