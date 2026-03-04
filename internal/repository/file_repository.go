package repository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

type FileRepository struct {
	mutex           sync.RWMutex
	fileStoragePath string
	store           map[string]model.ShortenURLRecord
}

func NewFileRepository(filePath string) *FileRepository {
	fr := FileRepository{fileStoragePath: filePath, store: make(map[string]model.ShortenURLRecord)}
	fr.fillCasheStorage()
	return &fr
}

func (r *FileRepository) Save(_ context.Context, record model.ShortenURLRecord) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	file, err := os.OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return err
	}

	defer file.Close()

	if err := json.NewEncoder(file).Encode(record); err != nil {
		return err
	}

	r.store[record.ShortURL] = record

	return nil
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
	file, err := os.OpenFile(r.fileStoragePath, os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)

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

	file, err := os.OpenFile(r.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return err
	}

	defer file.Close()
	defer r.mutex.Unlock()

	encoder := json.NewEncoder(file)

	for _, url := range shortenURLRecords {
		if err := encoder.Encode(url); err != nil {
			return err
		}
	}

	for _, url := range shortenURLRecords {
		r.store[url.ShortURL] = url
	}

	return nil
}
