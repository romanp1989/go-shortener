package shortenerservice

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/models"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"net/url"
	"strings"
)

type ShortenerService struct {
	storage *storage.Storage
	Cfg     *config.ConfigENV

	inChan    chan itemDelete
	closeChan chan struct{}
	size      int
}

func NewShortenerService(storage *storage.Storage, cfg *config.ConfigENV) *ShortenerService {
	service := &ShortenerService{
		storage:   storage,
		Cfg:       cfg,
		inChan:    make(chan itemDelete, 100),
		closeChan: make(chan struct{}),
		size:      100,
	}

	for i := 0; i < 2; i++ {
		go service.process()
	}

	return service
}

// ShortURL function for generate short name for URL
func (s *ShortenerService) ShortURL(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])
	encoded = strings.Replace(encoded, "/", "", -1)[:8]

	return encoded
}

// Encode function for creating a shortened URL based on the original one
func (s *ShortenerService) Encode(ctx context.Context, originalURL string) (string, error) {
	userID := auth.UIDFromContext(ctx)
	if userID == nil {

		return "", errors.Errorf("User unauthorized")
	}

	hashID := s.ShortURL(originalURL)
	shortID, err := s.storage.SaveURL(ctx, originalURL, hashID, userID)
	if err != nil {
		logger.Log.Debug("Ошибка добавления данных", zap.Error(err))

		var errConflict *storage.URLConflictError
		if errors.As(err, &errConflict) {
			shortID = errConflict.URL
			resp := fmt.Sprintf("%s/%s", s.Cfg.BaseURL, shortID)
			return resp, err
		} else {
			return "", err
		}
	}

	resp := fmt.Sprintf("%s/%s", s.Cfg.BaseURL, shortID)

	return resp, nil
}

// Decode service for getting the original URL from short URL
func (s *ShortenerService) Decode(id string) (string, error) {
	fullURL, err := s.storage.GetURL(id)
	if err != nil {
		return "", err
	}

	return fullURL, nil
}

// Shorten handler for creating a shortened URL based on the original one
func (s *ShortenerService) Shorten(ctx context.Context, originalURL string, userID *uuid.UUID) (string, error) {
	hashID := s.ShortURL(originalURL)
	shortID, err := s.storage.SaveURL(ctx, originalURL, hashID, userID)
	if err != nil {
		var errConflict *storage.URLConflictError
		if errors.As(err, &errConflict) {
			return fmt.Sprintf("%s/%s", s.Cfg.BaseURL, shortID), err
		} else {
			return "", err
		}
	}

	shortURL := fmt.Sprintf("%s/%s", s.Cfg.BaseURL, shortID)

	return shortURL, nil
}

// SaveBatch handler for creating a shortened URL based on the original one
func (s *ShortenerService) SaveBatch(ctx context.Context, batchReq []models.BatchShortenRequest, userID *uuid.UUID) ([]models.BatchShortenResponse, error) {
	var err error

	var shortURLs []models.StorageURL
	var hashID string

	for _, value := range batchReq {
		if _, err = url.ParseRequestURI(value.OriginalURL); err != nil {
			//w.WriteHeader(http.StatusBadRequest)
			return []models.BatchShortenResponse{}, err
		}

		hashID, err = s.storage.GetURL(value.OriginalURL)
		if err != nil {
			logger.Log.Debug("error get url response", zap.Error(err))
			//w.WriteHeader(http.StatusBadRequest)
			return []models.BatchShortenResponse{}, err
		}

		if hashID == "" {
			hashID = s.ShortURL(value.OriginalURL)
			shortURLs = append(shortURLs, models.StorageURL{
				OriginalURL: value.OriginalURL,
				ShortURL:    hashID,
			})
		} else {
			shortURLs = append(shortURLs, models.StorageURL{
				OriginalURL: value.OriginalURL,
				ShortURL:    hashID,
			})
		}
	}

	urls, err := s.storage.SaveBatchURL(ctx, shortURLs, userID)
	if err != nil {
		logger.Log.Debug("error urls save", zap.Error(err))
		return []models.BatchShortenResponse{}, err
	}

	res := make([]models.BatchShortenResponse, 0, len(urls))

	for i, shortURL := range urls {
		res = append(res, models.BatchShortenResponse{
			CorrelationID: batchReq[i].CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", s.Cfg.BaseURL, shortURL),
		})
	}

	return res, nil
}
