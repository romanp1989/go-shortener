package storage

import "github.com/romanp1989/go-shortener/internal/models"

type Storage struct {
	storage models.Storage
}

func Init(path string) *Storage {
	if path == "" {
		return &Storage{storage: NewCacheStorage()}
	}
	return &Storage{storage: NewFileStorage(path)}
}

func (s *Storage) GetURL(inputURL string) string {
	return s.storage.Get(inputURL)
}

func (s *Storage) SaveURL(originalURL string, shortURL string) error {
	return s.storage.Save(originalURL, shortURL)
}
