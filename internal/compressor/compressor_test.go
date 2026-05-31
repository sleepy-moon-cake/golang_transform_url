package compressor

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тест проверяет, что мидлвара успешно сжимает ответ,
// если клиент поддерживает gzip и тип контента совпадает (application/json)
func TestCompressor_ResponseCompression(t *testing.T) {
	// Создаем тестовый хэндлер, который возвращает JSON
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok","message":"success"}`))
	})

	// Оборачиваем его в нашу мидлвару Compressor
	handlerToTest := Compressor(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// Симулируем поддержку gzip со стороны клиента
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// 1. Проверяем, что сервер выставил правильный заголовок сжатия
	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Errorf("expected Content-Encoding 'gzip', got '%s'", resp.Header.Get("Content-Encoding"))
	}

	// 2. Проверяем, что данные действительно сжаты и успешно распаковываются
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader to read response: %v", err)
	}
	defer gzReader.Close()

	bodyBytes, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("failed to read decompressed body: %v", err)
	}

	expectedBody := `{"status":"ok","message":"success"}`
	if string(bodyBytes) != expectedBody {
		t.Errorf("expected body '%s', got '%s'", expectedBody, string(bodyBytes))
	}
}

// Тест проверяет, что мидлвара успешно распаковывает тело запроса,
// пришедшее от клиента в сжатом gzip формате
func TestCompressor_RequestDecompression(t *testing.T) {
	requestData := `{"url":"https://yandex.ru"}`

	// Сжимаем наши тестовые данные в gzip буфер
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	_, err := gzWriter.Write([]byte(requestData))
	if err != nil {
		t.Fatalf("failed to write data to gzip writer: %v", err)
	}
	_ = gzWriter.Close()

	// Тестовый хэндлер проверяет, что к нему приходят уже ЧИСТЫЕ (распакованные) данные
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body inside handler: %v", err)
		}

		if string(bodyBytes) != requestData {
			t.Errorf("handler received wrong data: expected '%s', got '%s'", requestData, string(bodyBytes))
		}
		w.WriteHeader(http.StatusOK)
	})

	handlerToTest := Compressor(testHandler)

	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	// Выставляем обязательные заголовки, чтобы сработал метод shouldDecompressRequest
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

// Тест проверяет, что если тип контента не подходит под правила сжатия (например, text/plain),
// то ответ сервера отдается обычным текстом БЕЗ gzip
func TestCompressor_NoCompressionForWrongContentType(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain") // text/plain нет в правилах WriteHeader вашего compressWriter
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("plain text response"))
	})

	handlerToTest := Compressor(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	w := httptest.NewRecorder()
	handlerToTest.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	// Заголовок Content-Encoding должен отсутствовать
	if resp.Header.Get("Content-Encoding") == "gzip" {
		t.Error("expected no Content-Encoding for text/plain content type")
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	if string(bodyBytes) != "plain text response" {
		t.Errorf("expected body 'plain text response', got '%s'", string(bodyBytes))
	}
}
