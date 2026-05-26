package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/mocks"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
	"go.uber.org/mock/gomock"
)

// Тест успешного создания короткого URL с проверкой UserId в контексте
func TestService_CreateShortURL_Success(t *testing.T) {
	ctrl := gomock.NewController(t)

	// 2. Создаем экземпляр сгенерированного мока
	mockRepo := mocks.NewMockRepository(ctrl)

	// 3. Настраиваем ожидания (EXPECT) для метода Save
	mockRepo.EXPECT().
		Save(
			gomock.Any(), // Разрешаем любой контекст
			gomock.Cond(func(record model.ShortenURLRecord) bool {
				// Переносим ваши проверки полей структуры прямо сюда
				if record.UserUUID != "user-123" {
					t.Errorf("expected UserUUID 'user-123', got '%s'", record.UserUUID)
					return false
				}
				if len(record.ShortURL) != 10 {
					t.Errorf("expected ShortURL length 10, got %d", len(record.ShortURL))
					return false
				}
				return true
			}),
		).
		// Говорим моку вернуть ту же запись, что пришла на вход, и отсутствие ошибки (nil)
		DoAndReturn(func(ctx context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error) {
			return record, nil
		})

	// 4. Передаем сгенерированный мок в ваш сервис
	s := NewService(mockRepo)
	defer close(s.deleteUrlsChannel)

	ctx := context.WithValue(context.Background(), contextkeys.UserId, "user-123")
	key, err := s.CreateShortURL(ctx, "https://yandex.ru")

	// 5. Проверяем результаты выполнения сервиса
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(key) != 10 {
		t.Errorf("expected generated key length 10, got %d", len(key))
	}
}

// Тест проверяет, что метод возвращает ошибку, если в контексте нет UserId
func TestService_CreateShortURL_NoUserInContext(t *testing.T) {
	ctrl := gomock.NewController(t)

	mockRepo := mocks.NewMockRepository(ctrl)

	s := NewService(mockRepo)
	defer close(s.deleteUrlsChannel)

	ctx := context.Background() // пустой контекст без UserId
	_, err := s.CreateShortURL(ctx, "https://yandex.ru")

	// 4. Проверяем ожидаемую ошибку
	if !errors.Is(err, contextkeys.ErrContextKey) {
		t.Errorf("expected error '%v', got '%v'", contextkeys.ErrContextKey, err)
	}
}

// Тест получения оригинального URL по коду
func TestService_GetURLByCode_Success(t *testing.T) {
	// 1. Инициализируем контроллер GoMock
	ctrl := gomock.NewController(t)

	// 2. Создаем экземпляр мока
	mockRepo := mocks.NewMockRepository(ctrl)

	expectedRecord := model.ShortenURLRecord{
		ShortURL:    "abcde12345",
		OriginalURL: "https://yandex.ru",
		DeletedFlag: false,
	}

	// 3. Настраиваем ожидание: метод FindByCode должен вызваться
	// с любым контекстом и строго с кодом "abcde12345"
	mockRepo.EXPECT().
		FindByCode(gomock.Any(), "abcde12345").
		Return(expectedRecord, nil)

	// 4. Передаем мок в сервис
	s := NewService(mockRepo)
	defer close(s.deleteUrlsChannel)

	// 5. Вызываем тестируемый метод
	url, err := s.GetURLByCode(context.Background(), "abcde12345")

	// 6. Проверяем результаты
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://yandex.ru" {
		t.Errorf("expected URL 'https://yandex.ru', got '%s'", url)
	}
}

// Тест проверяет, что если ссылка помечена как удаленная, сервис возвращает ErrURLBeenDeleted
func TestService_GetURLByCode_Deleted(t *testing.T) {
	// 1. Инициализируем контроллер GoMock
	ctrl := gomock.NewController(t)

	// 2. Создаем экземпляр мока
	mockRepo := mocks.NewMockRepository(ctrl)

	// Подготавливаем запись с флагом удаления
	deletedRecord := model.ShortenURLRecord{
		ShortURL:    "abcde12345",
		OriginalURL: "https://yandex.ru",
		DeletedFlag: true, // Ссылка удалена
	}

	// 3. Ожидаем вызов метода FindByCode с конкретным кодом
	mockRepo.EXPECT().
		FindByCode(gomock.Any(), "abcde12345").
		Return(deletedRecord, nil)

	// 4. Передаем мок в сервис
	s := NewService(mockRepo)
	defer close(s.deleteUrlsChannel)

	// 5. Вызываем метод и проверяем ожидаемую ошибку удаления
	_, err := s.GetURLByCode(context.Background(), "abcde12345")
	if !errors.Is(err, ErrURLBeenDeleted) {
		t.Errorf("expected error '%v', got '%v'", ErrURLBeenDeleted, err)
	}
}

// Тест проверяет асинхронную работу воркера удаления (DeleteWorker) по таймеру
func TestService_DeleteBatchWorker_Execution(t *testing.T) {
	// 1. Инициализируем контроллер GoMock
	ctrl := gomock.NewController(t)

	// 2. Создаем экземпляр мока
	mockRepo := mocks.NewMockRepository(ctrl)

	deleteCalled := make(chan bool, 1)

	// 3. Настраиваем ожидание вызова метода DeleteBatch
	mockRepo.EXPECT().
		// Ожидаем вызов с любым контекстом и слайсом из ровно 2 элементов
		DeleteBatch(gomock.Any(), gomock.Len(2)).
		// Внутри Do мы выполняем ваши кастомные проверки на значения
		Do(func(ctx context.Context, urls []model.ShortenUrlDeleteRecord) {
			if urls[0].UserUUID != "user-123" || urls[0].ShortURL != "code1" {
				t.Errorf("unexpected content in delete record: %+v", urls[0])
			}
			// Сигнализируем тестовому потоку, что асинхронный воркер успешно вызвал метод
			deleteCalled <- true
		}).
		// Возвращаем отсутствие ошибки
		Return(nil)

	// 4. Инициализируем сервис
	s := NewService(mockRepo)

	ctx := context.WithValue(context.Background(), contextkeys.UserId, "user-123")
	err := s.DeleteBatchUrl(ctx, []string{"code1", "code2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5. Ждем срабатывания тикера воркера
	select {
	case <-deleteCalled:
		// Успешно, воркер вычитал канал и вызвал метод репозитория
	case <-time.After(500 * time.Millisecond):
		t.Error("delete worker did not flush records within time limit")
	}

	close(s.deleteUrlsChannel)
}
