package storage

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/models"
	"log"
)

// Storage structure for storage
type Storage struct {
	Storage models.Storage
}

// Init Factory for create storage
func Init(dbPath string, path string) *Storage {
	if dbPath != "" {
		return &Storage{Storage: NewDB(dbPath)}
	}
	if path == "" {
		storage := NewCacheStorage()
		return &Storage{Storage: storage}
	}

	storage, err := NewFileStorage(path)
	if err != nil {
		log.Fatalf("Ошибка: %s", err)
	}

	return &Storage{Storage: storage}
}

// GetURL function for get URL from storage
func (s *Storage) GetURL(inputURL string) (string, error) {
	url, err := s.Storage.Get(inputURL)
	return url, err
}

// SaveURL function for save URL in storage
func (s *Storage) SaveURL(ctx context.Context, originalURL string, shortURL string, userID *uuid.UUID) (string, error) {
	return s.Storage.Save(ctx, originalURL, shortURL, userID)
}

// SaveBatchURL function for saving URL list
func (s *Storage) SaveBatchURL(ctx context.Context, urls []models.StorageURL, userID *uuid.UUID) ([]string, error) {
	return s.Storage.SaveBatch(ctx, urls, userID)
}

// GetAllUrlsByUser function for get all user's URLs
func (s *Storage) GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	return s.Storage.GetAllUrlsByUser(ctx, userID)
}

// DeleteUrlsBatch function for delete URLs list
func (s *Storage) DeleteUrlsBatch(ctx context.Context, userID *uuid.UUID, urls []string) error {
	return s.Storage.DeleteBatch(ctx, userID, urls)
}

// Ping function for ping storage connection
func (s *Storage) Ping(ctx context.Context) error {
	return s.Storage.Ping(ctx)
}

// Ping function for ping storage connection
func (s *Storage) GetStats(ctx context.Context) ([]models.StorageStats, error) {
	return s.Storage.GetStats(ctx)
}
