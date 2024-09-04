package models

import (
	"context"
	"github.com/gofrs/uuid"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type StorageURL struct {
	UserID      *uuid.UUID `json:"user_id"`
	OriginalURL string     `json:"original_url"`
	ShortURL    string     `json:"short_url"`
}

type Storage interface {
	Save(ctx context.Context, OriginalURL string, ShortURL string, userID *uuid.UUID) (string, error)
	Get(inputURL string) (string, error)
	SaveBatch(ctx context.Context, urls []StorageURL, userID *uuid.UUID) ([]string, error)
	Ping(ctx context.Context) error
	GetAllUrlsByUser(ctx context.Context, userID *uuid.UUID) ([]StorageURL, error)
	DeleteBatch(ctx context.Context, userID *uuid.UUID, urls []string) error
}

type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
