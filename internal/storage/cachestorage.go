package storage

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/models"
)

// CacheStorage Cache storage
type CacheStorage struct {
	storageURL map[string]string
}

// NewCacheStorage factory for create cache storage
func NewCacheStorage() *CacheStorage {
	return &CacheStorage{storageURL: make(map[string]string)}
}

// Get function for get URL from DB
func (s *CacheStorage) Get(inputURL string) (string, error) {
	if foundurl, ok := s.storageURL[inputURL]; ok {
		return foundurl, nil
	}
	return "", nil
}

// Save function for save URL in DB
func (s *CacheStorage) Save(ctx context.Context, originalURL string, shortURL string, userID *uuid.UUID) (string, error) {
	s.storageURL[shortURL] = originalURL
	s.storageURL[originalURL] = shortURL
	return shortURL, nil
}

// SaveBatch function for saving URL list
func (s *CacheStorage) SaveBatch(ctx context.Context, urls []models.StorageURL, userID *uuid.UUID) ([]string, error) {
	return nil, nil
}

// DeleteBatch function for delete URLs list
func (s *CacheStorage) DeleteBatch(ctx context.Context, userID *uuid.UUID, urls []string) error {
	return nil
}

// GetAllUrlsByUser function for get all user's URLs
func (s *CacheStorage) GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	return nil, nil
}

// Ping function for ping DB connection
func (s *CacheStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *CacheStorage) GetStats(ctx context.Context) (models.StorageStats, error) {
	return models.StorageStats{}, nil
}
