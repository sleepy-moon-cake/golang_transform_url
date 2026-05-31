package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type Storage interface {
	Send(context.Context, Event) error
}

type FileStorage struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		return nil, fmt.Errorf("failed to open file storage: %w", err)
	}

	encoder := json.NewEncoder(file)

	return &FileStorage{file: file, encoder: encoder}, nil
}

func (f *FileStorage) Send(_ context.Context, e Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := f.encoder.Encode(&e); err != nil {
		return fmt.Errorf("send: decode : %w", err)
	}
	return nil
}

func (f *FileStorage) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.file != nil {
		return f.file.Close()
	}
	return nil
}

type RemoteStorage struct {
	url    string
	client *http.Client
}

func NewRemoteStorage(url string) *RemoteStorage {
	return &RemoteStorage{url: url, client: &http.Client{Timeout: time.Second * 20}}
}

func (r *RemoteStorage) Send(ctx context.Context, e Event) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(&e); err != nil {
		return fmt.Errorf("send: encode : %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, &buf)
	if err != nil {
		return fmt.Errorf("send: request : %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := r.client.Do(req)

	if err != nil {
		return fmt.Errorf("send: do request : %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("remote storage: bad response status: %s", res.Status)
	}

	return nil
}
