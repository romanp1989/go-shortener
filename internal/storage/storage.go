package storage

import (
	"context"
	"github.com/romanp1989/go-shortener/internal/models"
	"log"
)

type Storage struct {
	storage models.Storage
}

func Init(dbPath string, path string) *Storage {
	if dbPath != "" {
		return &Storage{storage: NewDB(dbPath)}
	}
	if path == "" {
		storage := NewCacheStorage()
		return &Storage{storage: storage}
	}

	storage, err := NewFileStorage(path)
	if err != nil {
		log.Fatalf("Ошибка: %s", err)
	}

	return &Storage{storage: storage}
}

func (s *Storage) GetURL(inputURL string) (string, error) {
	url, err := s.storage.Get(inputURL)
	return url, err
}

func (s *Storage) SaveURL(originalURL string, shortURL string) error {
	return s.storage.Save(originalURL, shortURL)
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
