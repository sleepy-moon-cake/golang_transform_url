package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
)

type URLHandle struct {
	baseURL string
	service URLService
}

type URLService interface {
	CreateShortURL(ctx context.Context, str string) (string, error)
	GetURLByCode(ctx context.Context, code string) (string, error)
	Ping(ctx context.Context) error
	Batch(ctx context.Context, shorURLBatch []model.ShortenURLBatchRequest) ([]model.ShortenURLBatchResponse, error)
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

	if err != nil && !errors.Is(err, repository.ErrURLConflict) {
		slog.Error("Failed to create short URL", slog.String("URL", string(body)), slog.String("Error", err.Error()))
		http.Error(w, "Failed to create short URL", http.StatusInternalServerError)
		return
	}

	slog.Info("CreateShortURL", slog.String("URL", string(body)), slog.String("URL-ID", id))

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, id)
	w.Header().Set("Content-Type", "text/plain")

	status := http.StatusCreated

	if errors.Is(err, repository.ErrURLConflict) {
		status = http.StatusConflict
	}

	w.WriteHeader(status)
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
	defer r.Body.Close()

	var req model.ShortenURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Debug("decoding failed",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	recordURL, err := h.service.CreateShortURL(r.Context(), req.URL)

	if err != nil && !errors.Is(err, repository.ErrURLConflict) {
		slog.Error("failed to create short url", slog.String("error", err.Error()))
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	shortURL := fmt.Sprintf("%s/%s", h.baseURL, recordURL)

	response := model.ShortenURLResponse{
		Result: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")

	status := http.StatusCreated

	if errors.Is(err, repository.ErrURLConflict) {
		status = http.StatusConflict
	}

	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Debug("encoding failed", slog.String("error", err.Error()))
	}
}

func (h *URLHandle) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *URLHandle) Batch(w http.ResponseWriter, r *http.Request) {
	var requestData []model.ShortenURLBatchRequest

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Batch, decoding", slog.String("Error", err.Error()))
		return
	}

	responseData, err := h.service.Batch(r.Context(), requestData)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Batch, saving", slog.String("Error", err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	for i := range responseData {
		responseData[i].ShortURL = fmt.Sprintf("%s/%s", h.baseURL, responseData[i].ShortURL)
	}

	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		slog.Error("Batch, sending", slog.String("Error", err.Error()))
		return
	}
}
