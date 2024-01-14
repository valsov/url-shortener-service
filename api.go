package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"shortener/store"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type UrlRequest struct {
	Url string
}

type Api struct {
	store          *store.Store
	charset        []byte
	shortUrlLength int
}

func NewApi(store *store.Store, shortUrlCharset []byte, shortUrlLength int) *Api {
	return &Api{store, shortUrlCharset, shortUrlLength}
}

func (a *Api) create(w http.ResponseWriter, r *http.Request) {
	var baseUrl UrlRequest
	err := json.NewDecoder(r.Body).Decode(&baseUrl)
	if err != nil {
		log.Printf("failed to read and unmarshall request body, err: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	short, err := a.generateShortUrl()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortUrl := store.ShortUrl{
		Base:  baseUrl.Url,
		Short: short,
	}
	if err := a.store.Create(shortUrl); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *Api) delete(w http.ResponseWriter, r *http.Request) {
	url := chi.URLParam(r, "url")
	if err := a.store.Delete(url); err != nil {
		if err == store.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *Api) get(w http.ResponseWriter, r *http.Request) {
	url := chi.URLParam(r, "url")
	shortUrl, err := a.store.Get(url)
	if err != nil {
		if err == store.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	result := struct{ Url string }{shortUrl.Base}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, result)
}

// Generate a new short url that isn't already in use
func (a *Api) generateShortUrl() (string, error) {
	b := make([]byte, a.shortUrlLength)
	// Loop until a valid random sequence is generated
	for {
		for i := range b {
			b[i] = a.charset[rand.Intn(len(a.charset))]
		}
		url := string(b)

		// Verify if it exists
		_, err := a.store.Get(url)
		if err == nil {
			// Already exists
			continue
		}

		if err == store.ErrNotFound {
			// Not used: done
			return url, nil
		}

		// Unexpected error
		return "", err
	}
}
