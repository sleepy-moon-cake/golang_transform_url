// Package repository предоставляет абстракции и конкретные реализации
// для хранения, извлечения и пакетного удаления сокращенных URL-адресов.
package repository

import (
	"context"
	"errors"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/config/db"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

// Repository описывает единый интерфейс (контракт) для взаимодействия
// с хранилищами данных любого типа (база данных или файловая система).
type Repository interface {
	// Ping проверяет доступность и статус подключения к хранилищу.
	Ping(ctx context.Context) error
	// FindByCode выполняет поиск записи с оригинальным URL по короткому коду.
	FindByCode(ctx context.Context, code string) (model.ShortenURLRecord, error)
	// Save сохраняет новую запись с сокращенным URL в хранилище.
	Save(ctx context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error)
	// Batch выполняет массовую транзакционную вставку пакета записей.
	Batch(ctx context.Context, shortenURLRecords []model.ShortenURLRecord) error
	// GetURLsByUserID возвращает все ссылки, созданные конкретным пользователем.
	GetURLsByUserID(ctx context.Context, userID string) ([]model.ShortenURLRecord, error)
	// DeleteBatch выполняет массовое каскадное обновление флага удаления для списка ссылок.
	DeleteBatch(ctx context.Context, urls []model.ShortenUrlDeleteRecord) error
}

// NewRepository является фабричной функцией, которая инициализирует репозиторий.
// Если передан объект db, возвращает SQLRepository, иначе возвращает FileRepository.
func NewRepository(filePath string, db *db.DBSQL) (Repository, error) {
	if db != nil {
		return NewSQLRepository(db.DB), nil
	}

	return NewFileRepository(filePath)
}

// ErrURLConflict возвращается, если оригинальный URL-адрес уже существует в базе данных
// и нарушает ограничение уникальности.
var ErrURLConflict = errors.New("DB already has that URL")

// ErrNotFound возвращается, если запрашиваемый короткий код отсутствует в текущем хранилище.
var ErrNotFound = errors.New("not found")
