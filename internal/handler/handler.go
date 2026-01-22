package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
)

func ListenAndServe(cng *config.Config) error {
	handler := URLHandle{baseURL: cng.BaseUrlAddress}

	router := createRouter(&handler)

	return http.ListenAndServe(cng.ServerAddress, router)
}

func createRouter(handler *URLHandle) http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", handler.getshortURL)
		r.Post("/", handler.createshortURL)
	})
	return r
}

type URLHandle struct {
	baseURL string
}

func (h URLHandle) createshortURL(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	id := service.CreateshortURL(string(body))
	shortURL := fmt.Sprintf("http://localhost:8080/%s", id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h URLHandle) getshortURL(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, err := service.GetURLByCode(shortURL)

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
