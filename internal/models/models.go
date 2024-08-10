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
	Ping(ctx context.Context) error
}
