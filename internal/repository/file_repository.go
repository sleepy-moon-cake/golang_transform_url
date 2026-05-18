package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

type FileRepository struct {
	mutex   sync.RWMutex
	file    *os.File
	encoder *json.Encoder
	store   map[string]model.ShortenURLRecord
}

func NewFileRepository(filePath string) (*FileRepository, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, fmt.Errorf("NewFileRepository: %w", err)
	}

	encoder := json.NewEncoder(file)

	fr := FileRepository{
		file:    file,
		encoder: encoder,
		store:   make(map[string]model.ShortenURLRecord, 1000),
	}

	fr.fillCasheStorage()

	return &fr, nil
}

func (r *FileRepository) Save(_ context.Context, record model.ShortenURLRecord) (model.ShortenURLRecord, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if err := r.encoder.Encode(record); err != nil {
		return model.ShortenURLRecord{}, err
	}

	r.store[record.ShortURL] = record

	return record, nil
}

func (r *FileRepository) FindByCode(_ context.Context, code string) (model.ShortenURLRecord, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if record, ok := r.store[code]; ok {
		return record, nil
	}
	return model.ShortenURLRecord{}, ErrNotFound
}

func (r *FileRepository) fillCasheStorage() error {
	decoder := json.NewDecoder(r.file)

	for {
		var record model.ShortenURLRecord
		if err := decoder.Decode(&record); err != nil {
			if errors.Is(err, io.EOF) {
				break

			}
			return err
		}
		r.store[record.ShortURL] = record
	}
	return nil
}

func (r *FileRepository) Ping(_ context.Context) error {
	return errors.New("database not initialized")
}

func (r *FileRepository) Batch(ctx context.Context, shortenURLRecords []model.ShortenURLRecord) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for _, url := range shortenURLRecords {
		if err := r.encoder.Encode(url); err != nil {
			return err
		}
	}

	for _, url := range shortenURLRecords {
		r.store[url.ShortURL] = url
	}

	return nil
}

func (r *FileRepository) GetURLsByUserID(ctx context.Context, userID string) ([]model.ShortenURLRecord, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var userRecords []model.ShortenURLRecord

	for _, record := range r.store {
		if record.UserUUID == userID {
			userRecords = append(userRecords, record)
		}
	}

	return userRecords, nil
}

func (r *FileRepository) DeleteBatch(ctx context.Context, url []model.ShortenUrlDeleteRecord) error {
	return errors.New("not support for filestorage, pls use datastorage")
}
