package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/mocks"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/service"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const domainURL = "http://localhost:8080"
const longURL = "https://yandex.ru"
const userIDKey = "userId"

// --- TESTS FOR CreateShortURL ---
func TestUrlHandle_CreateShortURL_Scenarios(t *testing.T) {
	t.Run("Positive - Success 201", func(t *testing.T) {
		// 1. Создаем контроллер gomock для управления жизненным циклом моков
		ctrl := gomock.NewController(t)
		defer ctrl.Finish() // Вызовется в конце теста и проверит, все ли ожидания выполнились

		// 2. Инициализируем сгенерированный мок (укажите ваш правильный пакет с моками)
		svc := mocks.NewMockURLService(ctrl)

		// 3. Задаем ожидание вызова и возвращаемое значение
		svc.EXPECT().
			CreateShortURL(gomock.Any(), longURL). // Ожидаем любой контекст и конкретную строку longURL
			Return("short123", nil).               // Возвращаем результат
			Times(1)                               // Проверяем, что метод был вызван ровно 1 раз

		// Дальше ваш оригинальный код теста без изменений
		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		h.CreateShortURL(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, domainURL+"/short123", w.Body.String())
	})

	t.Run("Negative - Wrong Content-Type 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Создаем мок. Никаких ожиданий (EXPECT) не пишем,
		// так как сервис вообще не должен вызываться при неверном Content-Type.
		svc := mocks.NewMockURLService(ctrl)

		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		req.Header.Set("Content-Type", "application/json") // Неверный тип
		w := httptest.NewRecorder()

		h.CreateShortURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Positive - URL Conflict 409", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocks.NewMockURLService(ctrl)

		// Настраиваем мок на возврат ошибки конфликта и уже существующего ID
		svc.EXPECT().
			CreateShortURL(gomock.Any(), longURL).
			Return("existingID", repository.ErrURLConflict).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(longURL))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		h.CreateShortURL(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		assert.Equal(t, domainURL+"/existingID", w.Body.String())
	})
}

// --- TESTS FOR GetShortURL ---

func TestUrlHandle_GetShortURL_Scenarios(t *testing.T) {
	t.Run("Negative - Empty Path 400", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocks.NewMockURLService(ctrl)

		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		h.GetShortURL(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Positive - Temporary Redirect 307", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocks.NewMockURLService(ctrl)

		svc.EXPECT().
			GetURLByCode(gomock.Any(), "short123").
			Return(longURL, nil).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodGet, "/short123", nil)
		w := httptest.NewRecorder()

		h.GetShortURL(w, req)

		assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
		assert.Equal(t, longURL, w.Header().Get("Location"))
	})

	t.Run("Negative - URL Been Deleted 410 Gone", func(t *testing.T) {
		// 1. Инициализируем контроллер gomock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок
		svc := mocks.NewMockURLService(ctrl)

		// 3. Ожидаем вызов GetURLByCode с кодом "deleted123"
		// и возвращаем ошибку удаления
		svc.EXPECT().
			GetURLByCode(gomock.Any(), "deleted123").
			Return("", service.ErrURLBeenDeleted).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodGet, "/deleted123", nil)
		w := httptest.NewRecorder()

		h.GetShortURL(w, req)

		assert.Equal(t, http.StatusGone, w.Code)
	})
}

// --- TESTS FOR ShortenURL (JSON API) ---

func TestUrlHandle_ShortenURL_Scenarios(t *testing.T) {
	t.Run("Positive - JSON Shorten Success 201", func(t *testing.T) {
		// 1. Инициализируем контроллер gomock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок
		svc := mocks.NewMockURLService(ctrl)

		// 3. Ожидаем вызов CreateShortURL с распакованным URL из JSON
		svc.EXPECT().
			CreateShortURL(gomock.Any(), "https://yandex.ru").
			Return("json123", nil).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		reqBody := `{"url":"https://yandex.ru"}`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		h.ShortenURL(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		var res model.ShortenURLResponse
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, domainURL+"/json123", res.Result)
	})

	t.Run("Positive - JSON Shorten Conflict 409", func(t *testing.T) {
		// 1. Инициализируем контроллер gomock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок
		svc := mocks.NewMockURLService(ctrl)

		// 3. Ожидаем вызов и имитируем ошибку конфликта из репозитория
		svc.EXPECT().
			CreateShortURL(gomock.Any(), "https://yandex.ru").
			Return("conflictedJSON", repository.ErrURLConflict).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		reqBody := `{"url":"https://yandex.ru"}`
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		h.ShortenURL(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		// Если ваш хендлер при 409 для JSON также должен возвращать
		// сформированный URL в теле ответа, вы можете раскомментировать строки ниже:
		// var res model.ShortenURLResponse
		// err := json.Unmarshal(w.Body.Bytes(), &res)
		// require.NoError(t, err)
		// assert.Equal(t, domainURL+"/conflictedJSON", res.Result)
	})
}

// --- TESTS FOR Ping ---

func TestUrlHandle_Ping(t *testing.T) {
	t.Run("Success 200", func(t *testing.T) {
		// 1. Инициализируем контроллер gomock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок
		svc := mocks.NewMockURLService(ctrl)

		// 3. Ожидаем вызов Ping с любым контекстом и возвращаем nil (база доступна)
		svc.EXPECT().
			Ping(gomock.Any()).
			Return(nil).
			Times(1)

		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		h.Ping(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Failure 500", func(t *testing.T) {
		// 1. Инициализируем контроллер gomock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок
		svc := mocks.NewMockURLService(ctrl)

		// 3. Имитируем ошибку подключения к базе данных
		svc.EXPECT().
			Ping(gomock.Any()).
			Return(errors.New("db disconnect")).
			Times(1)

		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		w := httptest.NewRecorder()

		h.Ping(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// --- TESTS FOR GetUserShortUrls ---

func TestUrlHandle_GetUserShortUrls_Scenarios(t *testing.T) {
	t.Run("Negative - Unauthorized 401", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок
		svc := mocks.NewMockURLService(ctrl)

		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil) // Нет UserId в контексте
		w := httptest.NewRecorder()

		h.GetUserShortUrls(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Positive - No Content 204", func(t *testing.T) {
		// 1. Инициализируем контроллер gomock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок
		svc := mocks.NewMockURLService(ctrl)

		svc.EXPECT().
			GetUserShortUrls(gomock.Any()).
			Return([]model.ShortenURLRecord{}, nil). // Возвращаем пустой слайс (ссылок нет)
			Times(1)

		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserId, userIDKey)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetUserShortUrls(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("Positive - Success 200 with Records", func(t *testing.T) {
		// 1. Инициализируем контроллер gomock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// 2. Создаем сгенерированный мок сервиса
		svc := mocks.NewMockURLService(ctrl)

		// 3. Задаем возвращаемые моком тестовые данные
		mockRecords := []model.ShortenURLRecord{
			{ShortURL: "id1", OriginalURL: "http://site1.com"},
		}

		// Настраиваем ожидание вызова
		svc.EXPECT().
			GetUserShortUrls(gomock.Any()).
			Return(mockRecords, nil).
			Times(1)

		h := NewURLHandler(domainURL, svc)
		req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		ctx := context.WithValue(req.Context(), contextkeys.UserId, userIDKey)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		h.GetUserShortUrls(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		var records []model.ShortenURLRecord
		err := json.Unmarshal(w.Body.Bytes(), &records)
		require.NoError(t, err)
		assert.Len(t, records, 1)

		// Проверяем, что хендлер корректно добавил префикс домена к короткому URL
		assert.Equal(t, domainURL+"/id1", records[0].ShortURL)
	})
}

// --- TESTS FOR DeleteBatch ---

func TestUrlHandle_DeleteBatch(t *testing.T) {
	t.Run("Success 202 Accepted", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocks.NewMockURLService(ctrl)

		expectedUrls := []string{"code1", "code2"}

		svc.EXPECT().
			DeleteBatchUrl(gomock.Any(), gomock.Eq(expectedUrls)).
			Return(nil).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		reqBody := `["code1", "code2"]`
		req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", strings.NewReader(reqBody))
		w := httptest.NewRecorder()

		h.DeleteBatch(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)
	})
}

func TestUrlHandle_GetStats(t *testing.T) {
	t.Run("Positive - Success 200", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocks.NewMockURLService(ctrl)

		// Данные, которые должен вернуть сервис
		expectedStats := model.URLStats{
			Urls:  42,
			Users: 10,
		}

		// Настраиваем мок: ожидаем вызов GetStats и возвращаем структуру
		svc.EXPECT().
			GetStats(gomock.Any()).
			Return(expectedStats, nil).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		w := httptest.NewRecorder()

		h.GetStats(w, req)

		// Проверяем статус-код и заголовок ответа
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		// Проверяем тело JSON-ответа
		var res model.URLStats
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.Equal(t, expectedStats.Urls, res.Urls)
		assert.Equal(t, expectedStats.Users, res.Users)
	})

	t.Run("Negative - Service Error 500", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		svc := mocks.NewMockURLService(ctrl)

		// Имитируем внутреннюю ошибку базы данных или сервиса
		svc.EXPECT().
			GetStats(gomock.Any()).
			Return(model.URLStats{}, errors.New("failed to fetch stats")).
			Times(1)

		h := NewURLHandler(domainURL, svc)

		req := httptest.NewRequest(http.MethodGet, "/api/internal/stats", nil)
		w := httptest.NewRecorder()

		h.GetStats(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
