package service

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"math/big"
	"time"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/repository"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/shared/contextkeys"
)

var ErrURLBeenDeleted = errors.New("record was deleted")

type Service struct {
	repository        repository.Repository
	deleteUrlsChannel chan []model.ShortenUrlDeleteRecord
}

func NewService(repository repository.Repository) *Service {
	instance := &Service{
		repository:        repository,
		deleteUrlsChannel: make(chan []model.ShortenUrlDeleteRecord, 1024),
	}

	go instance.newDeleteWorker()

	return instance
}

func (s *Service) CreateShortURL(ctx context.Context, str string) (string, error) {
	key, err := generateKey()

	if err != nil {
		slog.Error("Key generation")
		return "", err
	}

	value, ok := ctx.Value(contextkeys.UserId).(string)

	if !ok {
		return "", contextkeys.ErrContextKey
	}

	record := model.ShortenURLRecord{
		ShortURL:    key,
		OriginalURL: str,
		UserUUID:    value,
	}

	if record, err := s.repository.Save(ctx, record); err != nil {
		slog.Error("Save record")
		return record.ShortURL, err
	}

	return key, nil
}

func (s *Service) GetURLByCode(ctx context.Context, code string) (string, error) {
	record, err := s.repository.FindByCode(ctx, code)

	if err != nil {
		return "", err
	}

	if record.DeletedFlag {
		return "", ErrURLBeenDeleted
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

func (s *Service) Batch(ctx context.Context, shortURLBatch []model.ShortenURLBatchRequest) ([]model.ShortenURLBatchResponse, error) {
	shortenURLBatchRequestRecords := make([]model.ShortenURLBatchResponse, 0, len(shortURLBatch))
	records := make([]model.ShortenURLRecord, 0, len(shortURLBatch))

	value, ok := ctx.Value(contextkeys.UserId).(string)

	if !ok {
		return nil, contextkeys.ErrContextKey
	}

	for _, url := range shortURLBatch {
		key, err := generateKey()

		if err != nil {
			slog.Error("Key generation")
			return nil, err
		}

		records = append(records, model.ShortenURLRecord{
			ShortURL:    key,
			OriginalURL: url.OriginalURL,
			UserUUID:    value,
		})

		shortenURLBatchRequestRecords = append(shortenURLBatchRequestRecords, model.ShortenURLBatchResponse{
			CorrelationID: url.CorrelationID,
			ShortURL:      key,
		})
	}

	if err := s.repository.Batch(ctx, records); err != nil {
		return nil, err
	}

	return shortenURLBatchRequestRecords, nil
}

func (s *Service) GetUserShortUrls(ctx context.Context) ([]model.ShortenURLRecord, error) {
	value, ok := ctx.Value(contextkeys.UserId).(string)

	if !ok {
		return nil, contextkeys.ErrContextKey
	}

	return s.repository.GetURLsByUserID(ctx, value)
}

func (s *Service) DeleteBatchUrl(ctx context.Context, shotUrls []string) error {
	userUUID, ok := ctx.Value(contextkeys.UserId).(string)

	if !ok {
		return contextkeys.ErrContextKey
	}

	urls := make([]model.ShortenUrlDeleteRecord, 0, len(shotUrls))

	for _, url := range shotUrls {
		urls = append(urls, model.ShortenUrlDeleteRecord{ShortURL: url, UserUUID: userUUID})
	}

	s.deleteUrlsChannel <- urls

	return nil
}

func (s *Service) newDeleteWorker() {
	deleteTimer := time.NewTicker(300 * time.Millisecond)
	defer deleteTimer.Stop()

	deleteUrls := make([]model.ShortenUrlDeleteRecord, 0)

	ctx := context.Background()

	for {
		select {
		case urls, ok := <-s.deleteUrlsChannel:
			if !ok {
				if len(deleteUrls) > 0 {
					err := s.repository.DeleteBatch(ctx, deleteUrls)
					if err != nil {
						slog.Error("failed to delete batch in worker",
							slog.String("error", err.Error()),
							slog.Int("count", len(deleteUrls)),
						)
					}

				}
				return
			}

			deleteUrls = append(deleteUrls, urls...)

			if len(deleteUrls) > 5 {
				err := s.repository.DeleteBatch(ctx, deleteUrls)
				if err != nil {
					slog.Error("failed to delete batch in worker",
						slog.String("error", err.Error()),
						slog.Int("count", len(deleteUrls)),
					)
				}
				deleteUrls = deleteUrls[:0]
				deleteTimer.Reset(300 * time.Millisecond)
			}

		case <-deleteTimer.C:
			if len(deleteUrls) == 0 {
				continue
			}

			err := s.repository.DeleteBatch(ctx, deleteUrls)
			if err != nil {
				slog.Error("failed to delete batch in worker",
					slog.String("error", err.Error()),
					slog.Int("count", len(deleteUrls)),
				)
			}
			deleteUrls = deleteUrls[:0]
		}
	}
}
