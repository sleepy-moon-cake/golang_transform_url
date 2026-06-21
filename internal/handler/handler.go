// Package handler предоставляет HTTP-хендлеры для маршрутизации,
// валидации и обработки входящих запросов сервиса сокращения URL.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
)

// URLHandle управляет HTTP-запросами, хранит базовый адрес сервера
// для генерации ссылок и интерфейс бизнес-логики.
type URLHandle struct {
	baseURL string
	service URLService
}

//go:generate mockgen -source=handler.go -destination=../mocks/service_mock.go -package=mocks

// URLService описывает контракт для работы с бизнес-логикой сокращения,
// хранения, пакетной обработки, удаления и извлечения URL-адресов.
type URLService interface {
	// CreateShortURL генерирует уникальный короткий идентификатор для оригинальной ссылки.
	CreateShortURL(ctx context.Context, str string) (string, error)
	// GetURLByCode извлекает оригинальный длинный URL по его сокращенному коду.
	GetURLByCode(ctx context.Context, code string) (string, error)
	// Ping выполняет проверку связи с используемым хранилищем данных.
	Ping(ctx context.Context) error
	// Batch выполняет массовое создание сокращенных кодов для пакета ссылок.
	Batch(ctx context.Context, shorURLBatch []model.ShortenURLBatchRequest) ([]model.ShortenURLBatchResponse, error)
	// GetUserShortUrls возвращает список всех созданных ссылок конкретного пользователя.
	GetUserShortUrls(ctx context.Context) ([]model.ShortenURLRecord, error)
	// DeleteBatchUrl ставит в очередь асинхронного воркера пакет кодов на удаление.
	DeleteBatchUrl(ctx context.Context, shotUrls []string) error
	// GetStats возращает статистику
	GetStats(context.Context) (model.URLStats, error)
}

// NewURLHandler выполняет инициализацию и возвращает новый указатель на структуру URLHandle.
func NewURLHandler(baseURL string, service URLService) *URLHandle {
	return &URLHandle{baseURL: baseURL, service: service}
}

// CreateShortURL обрабатывает POST-запросы со строкой в формате "text/plain".
// Возвращает статус 201 Created и сокращенную ссылку в теле ответа.
// Если URL уже существует в базе, возвращает статус 409 Conflict.
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

	id, createErr := h.service.CreateShortURL(r.Context(), string(body))

	if createErr != nil && !errors.Is(createErr, repository.ErrURLConflict) {
		slog.Error("CreateShortURL", slog.String("URL", string(body)), slog.String("Error", createErr.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	slog.Info("CreateShortURL", slog.String("URL", string(body)), slog.String("URL-ID", id))

	w.Header().Set("Content-Type", "text/plain")

	status := http.StatusCreated

	if errors.Is(createErr, repository.ErrURLConflict) {
		status = http.StatusConflict
	}

	shortURL, joinErr := url.JoinPath(h.baseURL, id)

	if joinErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write([]byte(shortURL))
}

// GetShortURL обрабатывает GET-запросы с коротким кодом в пути.
// Выполняет перенаправление 307 Temporary Redirect на оригинальный URL-адрес.
// Если ссылка была ранее удалена воркером, возвращает HTTP-статус 410 Gone.
func (h URLHandle) GetShortURL(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	shortURL := strings.TrimPrefix(r.URL.Path, "/")
	originalURL, err := h.service.GetURLByCode(r.Context(), shortURL)

	slog.Info("GetShortURL", slog.String("URL-SHORT", shortURL), slog.String("URL-ORIGIN", originalURL))

	if err != nil {
		if errors.Is(err, service.ErrURLBeenDeleted) {
			http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
			return
		}

		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Add("Location", originalURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

// ShortenURL обрабатывает POST-запросы к API, принимая JSON вида {"url": "..."}.
// Возвращает ответ со статусом 201 Created и JSON-телом {"result": "..."}.
// При конфликте уникальности возвращает статус 409 Conflict.
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
		slog.Error("ShortenURL", slog.String("error", err.Error()))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortURL, joinErr := url.JoinPath(h.baseURL, recordURL)

	if joinErr != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

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

// Ping проверяет работоспособность и доступность текущего хранилища данных.
// Возвращает статус 200 OK при успехе и 500 Internal Server Error при недоступности.
func (h *URLHandle) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Batch принимает JSON-массив с набором ссылок для их одновременного пакетного сокращения.
// Возвращает массив с correlation_id и сгенерированными короткими ссылками со статусом 201.
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

// GetUserShortUrls возвращает JSON-массив всех сокращенных URL, принадлежащих
// авторизованному пользователю (UUID извлекается из контекста).
// Если у пользователя нет сохраненных ссылок, возвращает статус 204 No Content.
func (h *URLHandle) GetUserShortUrls(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(contextkeys.UserId).(string)

	if !ok || userID == "" {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}

	records, err := h.service.GetUserShortUrls(r.Context())

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if len(records) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for i := range records {
		records[i].ShortURL = fmt.Sprintf("%s/%s", h.baseURL, records[i].ShortURL)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(records); err != nil {
		slog.Error("GetUserShortUrls", slog.String("error", err.Error()))
	}
}

// DeleteBatch принимает JSON-массив коротких кодов для их каскадного удаления.
// Не выполняет блокирующее удаление сразу, а возвращает статус 202 Accepted
// и отправляет данные в конкурентный канал неблокирующего фонового воркера.
func (h *URLHandle) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	var URLs = make([]string, 0)

	if err := json.NewDecoder(r.Body).Decode(&URLs); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteBatchUrl(r.Context(), URLs); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *URLHandle) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		slog.Error("GetStats", slog.String("error", err.Error()))
	}
}
