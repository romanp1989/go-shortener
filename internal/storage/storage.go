package storage

import (
	"context"
	"github.com/gofrs/uuid"
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

func (s *Storage) SaveURL(ctx context.Context, originalURL string, shortURL string, userID *uuid.UUID) (string, error) {
	return s.storage.Save(ctx, originalURL, shortURL, userID)
}

func (s *Storage) SaveBatchURL(ctx context.Context, urls []models.StorageURL, userID *uuid.UUID) ([]string, error) {
	return s.storage.SaveBatch(ctx, urls, userID)
}

func (s *Storage) GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	return s.storage.GetAllUrlsByUser(ctx, userID)
}

func (s *Storage) DeleteUrlsBatch(ctx context.Context, userID *uuid.UUID, urls []string) error {
	return s.storage.DeleteBatch(ctx, userID, urls)
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.storage.Ping(ctx)
}
