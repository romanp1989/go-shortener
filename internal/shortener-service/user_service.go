package shortenerservice

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/models"
)

// GetURLs function for creating a shortened URL based on the original one
func (s *ShortenerService) GetURLs(ctx context.Context, userID *uuid.UUID) ([]models.StorageURL, error) {
	urls, err := s.storage.GetAllUrlsByUser(ctx, userID)
	if err != nil {
		return []models.StorageURL{}, err
	}

	allUrls := make([]models.StorageURL, 0, len(urls))
	for _, v := range urls {
		var store models.StorageURL
		store.ShortURL = fmt.Sprintf("%s/%s", s.Cfg.BaseURL, v.ShortURL)
		store.OriginalURL = v.OriginalURL
		allUrls = append(allUrls, store)
	}

	return allUrls, nil
}
