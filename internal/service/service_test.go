package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
)

// --- MOCK REPOSITORY ---
type mockRepository struct {
	SaveFunc            func(ctx context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error)
	FindByCodeFunc      func(ctx context.Context, code string) (model.ShortenURLRecord, error)
	PingFunc            func(ctx context.Context) error
	BatchFunc           func(ctx context.Context, records []model.ShortenURLRecord) error
	GetURLsByUserIDFunc func(ctx context.Context, userID string) ([]model.ShortenURLRecord, error)
	DeleteBatchFunc     func(ctx context.Context, urls []model.ShortenUrlDeleteRecord) error
}

func (m *mockRepository) Save(ctx context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error) {
	return m.SaveFunc(ctx, record)
}
func (m *mockRepository) FindByCode(ctx context.Context, code string) (model.ShortenURLRecord, error) {
	return m.FindByCodeFunc(ctx, code)
}
func (m *mockRepository) Ping(ctx context.Context) error {
	return m.PingFunc(ctx)
}
func (m *mockRepository) Batch(ctx context.Context, records []model.ShortenURLRecord) error {
	return m.BatchFunc(ctx, records)
}
func (m *mockRepository) GetURLsByUserID(ctx context.Context, userID string) ([]model.ShortenURLRecord, error) {
	return m.GetURLsByUserIDFunc(ctx, userID)
}
func (m *mockRepository) DeleteBatch(ctx context.Context, urls []model.ShortenUrlDeleteRecord) error {
	if m.DeleteBatchFunc != nil {
		return m.DeleteBatchFunc(ctx, urls)
	}
	return nil
}

// --- TESTS ---

// Тест успешного создания короткого URL с проверкой UserId в контексте
func TestService_CreateShortURL_Success(t *testing.T) {
	mockRepo := &mockRepository{
		SaveFunc: func(ctx context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error) {
			if record.UserUUID != "user-123" {
				t.Errorf("expected UserUUID 'user-123', got '%s'", record.UserUUID)
			}
			if len(record.ShortURL) != 10 {
				t.Errorf("expected ShortURL length 10, got %d", len(record.ShortURL))
			}
			return record, nil
		},
	}

	s := NewService(mockRepo)
	defer close(s.deleteUrlsChannel)

	ctx := context.WithValue(context.Background(), contextkeys.UserId, "user-123")
	key, err := s.CreateShortURL(ctx, "https://yandex.ru")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key) != 10 {
		t.Errorf("expected generated key length 10, got %d", len(key))
	}
}

// Тест проверяет, что метод возвращает ошибку, если в контексте нет UserId
func TestService_CreateShortURL_NoUserInContext(t *testing.T) {
	s := NewService(&mockRepository{})
	defer close(s.deleteUrlsChannel)

	ctx := context.Background() // пустой контекст без UserId
	_, err := s.CreateShortURL(ctx, "https://yandex.ru")

	if !errors.Is(err, contextkeys.ErrContextKey) {
		t.Errorf("expected error '%v', got '%v'", contextkeys.ErrContextKey, err)
	}
}

// Тест получения оригинального URL по коду
func TestService_GetURLByCode_Success(t *testing.T) {
	mockRepo := &mockRepository{
		FindByCodeFunc: func(ctx context.Context, code string) (model.ShortenURLRecord, error) {
			return model.ShortenURLRecord{
				ShortURL:    "abcde12345",
				OriginalURL: "https://yandex.ru",
				DeletedFlag: false,
			}, nil
		},
	}

	s := NewService(mockRepo)
	defer close(s.deleteUrlsChannel)

	url, err := s.GetURLByCode(context.Background(), "abcde12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://yandex.ru" {
		t.Errorf("expected URL 'https://yandex.ru', got '%s'", url)
	}
}

// Тест проверяет, что если ссылка помечена как удаленная, сервис возвращает ErrURLBeenDeleted
func TestService_GetURLByCode_Deleted(t *testing.T) {
	mockRepo := &mockRepository{
		FindByCodeFunc: func(ctx context.Context, code string) (model.ShortenURLRecord, error) {
			return model.ShortenURLRecord{
				ShortURL:    "abcde12345",
				OriginalURL: "https://yandex.ru",
				DeletedFlag: true, // Ссылка удалена
			}, nil
		},
	}

	s := NewService(mockRepo)
	defer close(s.deleteUrlsChannel)

	_, err := s.GetURLByCode(context.Background(), "abcde12345")
	if !errors.Is(err, ErrURLBeenDeleted) {
		t.Errorf("expected error '%v', got '%v'", ErrURLBeenDeleted, err)
	}
}

// Тест проверяет асинхронную работу воркера удаления (DeleteWorker) по таймеру
func TestService_DeleteBatchWorker_Execution(t *testing.T) {
	deleteCalled := make(chan bool, 1)

	mockRepo := &mockRepository{
		DeleteBatchFunc: func(ctx context.Context, urls []model.ShortenUrlDeleteRecord) error {
			if len(urls) != 2 {
				t.Errorf("expected 2 urls to delete, got %d", len(urls))
			}
			if urls[0].UserUUID != "user-123" || urls[0].ShortURL != "code1" {
				t.Errorf("unexpected content in delete record: %+v", urls[0])
			}
			deleteCalled <- true
			return nil
		},
	}

	s := NewService(mockRepo)

	ctx := context.WithValue(context.Background(), contextkeys.UserId, "user-123")
	err := s.DeleteBatchUrl(ctx, []string{"code1", "code2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Ждем срабатывания тикера воркера (в коде прописано 300ms, закладываем 500ms на прохождение канала)
	select {
	case <-deleteCalled:
		// Успешно, воркер вычитал канал и вызвал метод репозитория
	case <-time.After(500 * time.Millisecond):
		t.Error("delete worker did not flush records within time limit")
	}

	close(s.deleteUrlsChannel)
}
