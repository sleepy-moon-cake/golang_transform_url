package repository

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

type Repository struct {
	mutex           sync.RWMutex
	fileStoragePath string
	store           map[string]model.ShortenURLRecord
}

var NotFoundRecord = errors.New("Not found")

func NewRepository(fileStoragePath string) *Repository {
	rep := Repository{fileStoragePath: fileStoragePath, store: make(map[string]model.ShortenURLRecord)}
	if err := rep.fillCasheStorage(); err != nil {
		slog.Error("Filling cashe error", slog.String("Error", err.Error()))
	}

	return &rep
}

func (r *Repository) Save(record model.ShortenURLRecord) error {
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

func (r *Repository) FindByCode(code string) (model.ShortenURLRecord, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if record, ok := r.store[code]; ok {
		return record, nil
	}
	return model.ShortenURLRecord{}, NotFoundRecord
}

func (r *Repository) fillCasheStorage() error {
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
