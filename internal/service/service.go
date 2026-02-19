package service

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"

	"github.com/google/uuid"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
)

type Service struct {
	repository *repository.Repository
}

func NewService(repository *repository.Repository) *Service {
	return &Service{
		repository: repository,
	}
}

func (s *Service) CreateShortURL(str string) (string, error) {
	key, err := generateKey()

	if err != nil {
		slog.Error("Key generation")
		return "", err
	}

	record := model.ShortenURLRecord{
		ID:          uuid.NewString(),
		ShortURL:    key,
		OriginalURL: str,
	}

	if err := s.repository.Save(record); err != nil {
		slog.Error("Save record")
		return "", err
	}

	return key, nil
}

func (s *Service) GetURLByCode(code string) (string, error) {
	record, err := s.repository.FindByCode(code)

	if err != nil {
		return "", err
	}

	return record.OriginalURL, nil
}

func generateKey() (string, error) {
	const values = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 10

	b := make([]byte, length)

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(values))))
		if err != nil {
			return "", err
		}
		b[i] = values[n.Int64()]
	}

	return string(b), nil
}

func (s *Service) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}
