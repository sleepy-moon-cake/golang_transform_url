package service

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/sleepy-moon-cake/golang_transform_url/internal/model"
)

func NewService(path string) *Service {
	return &Service{
		fileStoragePath: path,
	}
}

type Service struct {
	fileStoragePath string
	mutex           sync.Mutex
}

func (s *Service) CreateShortURL(str string) string {
	records, err := s.getAllRecords()

	if err != nil {
		panic(err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	key, err := generateKey()

	if err != nil {
		panic(err)
	}

	record := model.ShortenURLRecord{
		ID:          uuid.NewString(),
		ShortURL:    key,
		OriginalURL: str,
	}

	records = append(records, record)

	file, err := os.OpenFile(s.fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(&records); err != nil {
		panic(err)
	}
	return key
}

func (s *Service) GetURLByCode(code string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	records, err := s.getAllRecords()

	if err != nil {
		return "", errors.New("cant read records")
	}

	for _, v := range records {
		if v.ShortURL == code {
			return v.OriginalURL, nil
		}
	}

	return "", errors.New("no data")
}

func (s *Service) getAllRecords() ([]model.ShortenURLRecord, error) {
	records := []model.ShortenURLRecord{}

	data, err := os.ReadFile(s.fileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			return records, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return records, nil
	}

	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}

	return records, nil
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
