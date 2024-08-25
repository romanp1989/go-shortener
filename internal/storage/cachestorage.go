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

func (c *CacheStorage) Get(inputURL string) (string, error) {
	if foundurl, ok := c.storageURL[inputURL]; ok {
		return foundurl, nil
	}
	return "", nil
}

func (c *CacheStorage) Save(ctx context.Context, originalURL string, shortURL string, userID *uuid.UUID) (string, error) {
	c.storageURL[shortURL] = originalURL
	c.storageURL[originalURL] = shortURL
	return shortURL, nil
}

func (c *CacheStorage) SaveBatch(ctx context.Context, urls []models.StorageURL, userID *uuid.UUID) ([]string, error) {
	return nil, nil
}

func (s *CacheStorage) GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	return nil, nil
}

func (c *CacheStorage) Ping(ctx context.Context) error {
	return nil
}
