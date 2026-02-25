package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

type URLHandle struct {
	baseURL string
	service URLService
}

type URLService interface {
	CreateShortURL(ctx context.Context, str string) (string, error)
	GetURLByCode(ctx context.Context, code string) (string, error)
	Ping(ctx context.Context) error
}

func NewURLHandler(baseURL string, service URLService) *URLHandle {
	return &URLHandle{baseURL: baseURL, service: service}
}

func (h URLHandle) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateShortURL(r.Context(), string(body))

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

func (h URLHandle) GetShortURL(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, err := h.service.GetURLByCode(r.Context(), shortURL)

	slog.Info("GetShortURL", slog.String("URL-SHORT", shortURL), slog.String("URL-ORIGIN", originalURL))

	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h URLHandle) ShortenURL(w http.ResponseWriter, r *http.Request) {
	var shortenURL model.ShortenURLRequest

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&shortenURL); err != nil {
		slog.Debug("Decoding is failed", slog.String("Method", r.Method), slog.String("path", r.URL.Path))

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	recordURL, err := h.service.CreateShortURL(r.Context(), shortenURL.URL)

	if err != nil {
		slog.Error("Failed to send short URL", slog.String("Error", err.Error()))
		http.Error(w, "Failed to send short URL", http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, recordURL)

	var response = model.ShortenURLResponse{Result: shortURL}

	enc := json.NewEncoder(w)
	if err := enc.Encode(response); err != nil {
		slog.Debug("Encoding is failed", slog.String("Method", r.Method), slog.String("path", r.URL.Path))

		w.WriteHeader(http.StatusInternalServerError)
	}
	slog.Debug("HTTP 200")
}

func (h *URLHandle) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
