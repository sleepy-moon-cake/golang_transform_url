package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/compressor"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/config"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/logger"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
)

func ListenAndServe(cng *config.Config) error {
	handler := URLHandle{baseURL: cng.BaseURLAddress, service: service.NewService(cng.FileStoragePath)}

	router := createRouter(&handler)

	return http.ListenAndServe(cng.ServerAddress, logger.Logger(compressor.Compressor(router)))
}

func createRouter(handler *URLHandle) http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/{shortURL}", handler.getShortURL)
		r.Post("/", handler.createShortURL)
	})
	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", handler.shortenURL)
	})

	return r
}

type URLHandle struct {
	baseURL string
	service *service.Service
}

func (h URLHandle) createShortURL(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateShortURL(string(body))

	if err != nil {
		slog.Error("Failed to create short URL", slog.String("URL", string(body)), slog.String("Error", err.Error()))
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	slog.Info("CreateShortURL", slog.String("URL", string(body)), slog.String("URL-ID", id))

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, id)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (h URLHandle) getShortURL(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, err := h.service.GetURLByCode(shortURL)

	slog.Info("GetShortURL", slog.String("URL-SHORT", shortURL), slog.String("URL-ORIGIN", originalURL))

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h URLHandle) shortenURL(w http.ResponseWriter, r *http.Request) {
	var shortenURL model.ShortenURLRequest

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&shortenURL); err != nil {
		slog.Debug("Decoding is failed", slog.String("Method", r.Method), slog.String("path", r.URL.Path))

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	recordURl, err := h.service.CreateShortURL(shortenURL.URL)

	if err != nil {
		slog.Error("Failed to send short URL", slog.String("Error", err.Error()))
		http.Error(w, "Failed to send short URL", http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, recordURl)

	var response = model.ShortenURLResponse{Result: shortURL}

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil {
		slog.Debug("Encoding is failed", slog.String("Method", r.Method), slog.String("path", r.URL.Path))

		w.WriteHeader(http.StatusInternalServerError)
	}
	slog.Debug("HTTP 200")
}
