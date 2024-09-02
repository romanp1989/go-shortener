package storage

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/models"
)

type CacheStorage struct {
	storageURL map[string]string
}

func NewCacheStorage() *CacheStorage {
	return &CacheStorage{storageURL: make(map[string]string)}
}

func (s *CacheStorage) Get(inputURL string) (string, error) {
	if foundurl, ok := s.storageURL[inputURL]; ok {
		return foundurl, nil
	}
	return "", nil
}

func (s *CacheStorage) Save(ctx context.Context, originalURL string, shortURL string, userID *uuid.UUID) (string, error) {
	s.storageURL[shortURL] = originalURL
	s.storageURL[originalURL] = shortURL
	return shortURL, nil
}

func (s *CacheStorage) SaveBatch(ctx context.Context, urls []models.StorageURL, userID *uuid.UUID) ([]string, error) {
	return nil, nil
}

func (s *CacheStorage) DeleteBatch(ctx context.Context, userID *uuid.UUID, urls []string) error {
	return nil
}

func (s *CacheStorage) GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	return nil, nil
}

func (s *CacheStorage) Ping(ctx context.Context) error {
	return nil
}
