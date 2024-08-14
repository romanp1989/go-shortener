package models

import "context"

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type StorageURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type Storage interface {
	Save(OriginalURL string, ShortURL string) error
	Get(inputURL string) (string, error)
	SaveBatch(ctx context.Context, urls []StorageURL) ([]string, error)
	Ping(ctx context.Context) error
}

type BatchShortenRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchShortenResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
