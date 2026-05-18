package logger

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Тест проверяет, что кастомный ResponseWriter правильно перехватывает
// статус ответа и накапливает размер записанного тела (размер в байтах)
func TestLoggerResponse_CaptureData(t *testing.T) {
	rec := httptest.NewRecorder()
	data := &loggerResponseData{status: 200} // дефолтное значение как в middleware

	lr := &loggerResponse{
		ResponseWriter: rec,
		loggerData:     data,
	}

	// Симулируем отправку специфичного заголовка ответа
	expectedStatus := http.StatusCreated // 201
	lr.WriteHeader(expectedStatus)

	// Симулируем поэтапную запись тела ответа
	bytesFirst := []byte("hello")
	bytesSecond := []byte(" world")

	n1, err := lr.Write(bytesFirst)
	if err != nil {
		t.Fatalf("unexpected error on first write: %v", err)
	}

	n2, err := lr.Write(bytesSecond)
	if err != nil {
		t.Fatalf("unexpected error on second write: %v", err)
	}

	// 1. Проверяем корректность перехваченного статуса кода
	if data.status != expectedStatus {
		t.Errorf("expected captured status %d, got %d", expectedStatus, data.status)
	}

	// 2. Проверяем, что размер суммируется корректно
	expectedSize := len(bytesFirst) + len(bytesSecond) // 5 + 6 = 11
	if data.size != expectedSize {
		t.Errorf("expected captured size %d, got %d", expectedSize, data.size)
	}

	// 3. Убеждаемся, что нижележащий ResponseRecorder получил правильное количество байт
	if rec.Body.Len() != expectedSize {
		t.Errorf("expected recorder body length %d, got %d", expectedSize, rec.Body.Len())
	}

	if n1+n2 != expectedSize {
		t.Errorf("expected write return values sum %d, got %d", expectedSize, n1+n2)
	}
}

// Тест проверяет сквозное прохождение запроса через middleware Logger
func TestLoggerMiddleware_Execution(t *testing.T) {
	// Создаем простейший хэндлер, который возвращает 200 OK и пишет 4 байта
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	handlerToTest := Logger(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	// Запускаем выполнение (тест проверяет, что middleware не падает
	// и корректно передает контекст выполнения дальше по цепочке)
	handlerToTest.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "pong" {
		t.Errorf("expected body 'pong', got '%s'", w.Body.String())
	}
}

// Тест проверяет корректность работы маппинга строк в уровни логирования slog
func TestParseStringToLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{
			name:  "debug level",
			input: "debug",
			want:  slog.LevelDebug,
		},
		{
			name:  "warn level",
			input: "warn",
			want:  slog.LevelWarn,
		},
		{
			name:  "info level explicit",
			input: "info",
			want:  slog.LevelInfo,
		},
		{
			name:  "fallback to info on unknown string",
			input: "invalid_level_name",
			want:  slog.LevelInfo,
		},
		{
			name:  "fallback to info on empty string",
			input: "",
			want:  slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStringToLogLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseStringToLogLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
