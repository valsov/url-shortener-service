package main

import (
	"log"
	"net/http"
	"os"
	"shortener/store"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func main() {
	dbUrl := os.Getenv("DB_URL")
	dbName := os.Getenv("DB_NAME")
	collectionName := os.Getenv("COLLECTION_NAME")
	shortUrlLengthStr := os.Getenv("SHORT_URL_LENGTH")
	shortUrlLength, err := strconv.Atoi(shortUrlLengthStr)
	if err != nil {
		log.Fatalf("failed to convert SHORT_URL_LENGTH to int: %v", err)
	}

	store, err := store.NewStore(dbUrl, dbName, collectionName)
	if err != nil {
		log.Fatalf("error creating store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("error closing store: %v", err)
		}
	}()
	api := NewApi(store, []byte("abcdefghijklmnopqrstuvwxyz0123456789"), shortUrlLength)

	err = start(1234, api)
	if err != nil {
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

	// Start server
	log.Print("server started on :3000")
	return http.ListenAndServe(":3000", r)
}
