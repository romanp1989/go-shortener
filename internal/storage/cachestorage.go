package storage

import (
	"context"
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

func (c *CacheStorage) Save(originalURL string, shortURL string) error {
	c.storageURL[shortURL] = originalURL
	c.storageURL[originalURL] = shortURL
	return nil
}

func (c *CacheStorage) SaveBatch(ctx context.Context, urls []models.StorageURL) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CacheStorage) Ping(ctx context.Context) error {
	return nil
}
