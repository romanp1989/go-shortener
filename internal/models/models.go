package models

import (
	"context"
	"github.com/gofrs/uuid"
)

// ShortenRequest structure for Shorten handler request
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse structure for Shorten handler response
type ShortenResponse struct {
	Result string `json:"result"`
}

// StorageURL structure for save URLs in DB
type StorageURL struct {
	UserID      *uuid.UUID `json:"user_id"`
	OriginalURL string     `json:"original_url"`
	ShortURL    string     `json:"short_url"`
}

// Storage interface for storage
type Storage interface {
	Save(ctx context.Context, OriginalURL string, ShortURL string, userID *uuid.UUID) (string, error)
	Get(inputURL string) (string, error)
	SaveBatch(ctx context.Context, urls []StorageURL, userID *uuid.UUID) ([]string, error)
	Ping(ctx context.Context) error
	GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]StorageURL, error)
	DeleteBatch(ctx context.Context, userID *uuid.UUID, urls []string) error
	GetStats(ctx context.Context) ([]StorageStats, error)
}

// BatchShortenRequest structure for batch save URLs handler request
type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchShortenResponse structure for batch save URLs handler response
type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type StorageStats struct {
	Users int64 `json:"users"`
	URLs  int64 `json:"urls"`
}
