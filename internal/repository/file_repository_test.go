package repository

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Вспомогательная функция для создания репозитория во временном файле
func newTestFileRepository(t *testing.T) (*FileRepository, func()) {
	tmpFile, err := os.CreateTemp("", "repo_test_*.json")
	require.NoError(t, err)

	repo, err := NewFileRepository(tmpFile.Name())
	require.NoError(t, err)

	// Возвращаем репозиторий и функцию очистки (закрытие дескриптора и удаление файла)
	cleanup := func() {
		if repo.file != nil {
			_ = repo.file.Close()
		}
		_ = os.Remove(tmpFile.Name())
	}

	return repo, cleanup
}

func TestFileRepository_SaveAndFind(t *testing.T) {
	repo, cleanup := newTestFileRepository(t)
	defer cleanup()

	ctx := context.Background()
	record := model.ShortenURLRecord{
		ShortURL:    "short123",
		OriginalURL: "https://yandex.ru",
		UserUUID:    "user-1",
	}

	// 1. Проверяем сохранение
	saved, err := repo.Save(ctx, record)
	require.NoError(t, err)
	assert.Equal(t, record.ShortURL, saved.ShortURL)

	// 2. Проверяем успешный поиск
	found, err := repo.FindByCode(ctx, "short123")
	require.NoError(t, err)
	assert.Equal(t, "https://yandex.ru", found.OriginalURL)
	assert.Equal(t, "user-1", found.UserUUID)

	// 3. Проверяем поиск несуществующего кода
	_, err = repo.FindByCode(ctx, "not-exist")
	assert.Error(t, err)
	// Замените на вашу переменную ошибки (например, ErrNotFound или аналогичную), если она отличается
	assert.True(t, errors.Is(err, ErrNotFound) || strings.Contains(err.Error(), "not found"))
}

func TestFileRepository_Batch(t *testing.T) {
	repo, cleanup := newTestFileRepository(t)
	defer cleanup()

	ctx := context.Background()
	records := []model.ShortenURLRecord{
		{ShortURL: "batch1", OriginalURL: "https://site1.com", UserUUID: "user-2"},
		{ShortURL: "batch2", OriginalURL: "https://site2.com", UserUUID: "user-2"},
	}

	// Проверяем массовое сохранение
	err := repo.Batch(ctx, records)
	require.NoError(t, err)

	// Проверяем, что все записи доступны в кэше
	found1, err := repo.FindByCode(ctx, "batch1")
	require.NoError(t, err)
	assert.Equal(t, "https://site1.com", found1.OriginalURL)

	found2, err := repo.FindByCode(ctx, "batch2")
	require.NoError(t, err)
	assert.Equal(t, "https://site2.com", found2.OriginalURL)
}

func TestFileRepository_GetURLsByUserID(t *testing.T) {
	repo, cleanup := newTestFileRepository(t)
	defer cleanup()

	ctx := context.Background()
	_, _ = repo.Save(ctx, model.ShortenURLRecord{ShortURL: "id1", OriginalURL: "https://a.com", UserUUID: "user-target"})
	_, _ = repo.Save(ctx, model.ShortenURLRecord{ShortURL: "id2", OriginalURL: "https://b.com", UserUUID: "user-other"})
	_, _ = repo.Save(ctx, model.ShortenURLRecord{ShortURL: "id3", OriginalURL: "https://c.com", UserUUID: "user-target"})

	// Проверяем фильтрацию по ID пользователя
	userRecords, err := repo.GetURLsByUserID(ctx, "user-target")
	require.NoError(t, err)
	assert.Len(t, userRecords, 2)

	// Проверяем, что чужая ссылка не попала в выборку
	for _, rec := range userRecords {
		assert.Equal(t, "user-target", rec.UserUUID)
		assert.NotEqual(t, "id2", rec.ShortURL)
	}
}

func TestFileRepository_CacheRecovery(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "repo_cache_test_*.json")
	require.NoError(t, err)
	filePath := tmpFile.Name()
	_ = tmpFile.Close() // Закрываем базовый дескриптор, чтобы NewFileRepository мог с ним работать
	defer os.Remove(filePath)

	ctx := context.Background()

	// Этап 1: Создаем первый инстанс репозитория и сохраняем данные
	repo1, err := NewFileRepository(filePath)
	require.NoError(t, err)

	_, err = repo1.Save(ctx, model.ShortenURLRecord{ShortURL: "recX", OriginalURL: "https://google.com", UserUUID: "user-x"})
	require.NoError(t, err)
	_ = repo1.file.Close() // Симулируем завершение процесса (закрываем файл)

	// Этап 2: Создаем второй инстанс репозитория, указывая на ТОТ ЖЕ файл
	repo2, err := NewFileRepository(filePath)
	require.NoError(t, err)
	defer repo2.file.Close()

	// Проверяем, что метод fillCasheStorage автоматически восстановил данные из файла в RAM при старте
	found, err := repo2.FindByCode(ctx, "recX")
	require.NoError(t, err, "cache should populate store map from physical file on startup")
	assert.Equal(t, "https://google.com", found.OriginalURL)
}

func TestFileRepository_UnsupportedMethods(t *testing.T) {
	repo, cleanup := newTestFileRepository(t)
	defer cleanup()

	// Метод Ping должен возвращать фиксированную ошибку
	err := repo.Ping(context.Background())
	assert.Error(t, err)

	// Метод DeleteBatch для файлового хранилища не поддерживается
	err = repo.DeleteBatch(context.Background(), []model.ShortenUrlDeleteRecord{})
	assert.Error(t, err)
}
