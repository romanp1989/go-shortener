package handlers

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/models"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Handlers handlers
type Handlers struct {
	storage storage.Storage
}

// New Factory for create handlers
func New(storage storage.Storage) Handlers {
	return Handlers{
		storage: storage,
	}
}

// Encode handler for creating a shortened URL based on the original one
// @Accept string
// @Success 201 {string} short URL
// @Failure 400 bad request
// @Failure 401 error if user unauthorized
// @Failure 409 error if URL already exists in DB
func (h *Handlers) Encode() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		_, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		userID := auth.UIDFromContext(ctx)
		if userID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil || string(body) == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		stringURI := string(body)

		if _, err := url.ParseRequestURI(stringURI); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hashID := shortURL(stringURI)
		shortID, err := h.storage.SaveURL(r.Context(), stringURI, hashID, userID)
		if err != nil {
			logger.Log.Debug("Ошибка добавления данных", zap.Error(err))

			var errConflict *storage.URLConflictError
			if errors.As(err, &errConflict) {
				shortID = errConflict.URL
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(fmt.Sprintf("%s/%s", config.Options.FlagShortURL, shortID)))
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		resp := fmt.Sprintf("%s/%s", config.Options.FlagShortURL, shortID)

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(resp))
	}

	return http.HandlerFunc(fn)
}

func shortURL(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])
	encoded = strings.Replace(encoded, "/", "", -1)[:8]

	return encoded
}

// Decode handler for getting the original URL from short URL
// @Accept string
// @Success 307 {string} redirect to result URL
// @Failure 400 bad request
// @Failure 410 error if URL already deleted
// @Failure 404 error if URL not found
// @Failure 409 error if URL already exists in DB
func (h *Handlers) Decode() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Некорректный тип запроса", http.StatusBadRequest)
			return
		}

		id := chi.URLParam(r, "id")

		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fullURL, err := h.storage.GetURL(id)
		if err != nil {
			var errURLDeleted *storage.AlreadyDeleted
			if errors.As(err, &errURLDeleted) {
				w.WriteHeader(http.StatusGone)
				return
			}
			logger.Log.Debug("error get url response", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if fullURL != "" {
			http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, "Не найден url для указанного ID", http.StatusNotFound)
	}
	return http.HandlerFunc(fn)
}

// Shorten handler for creating a shortened URL based on the original one
// @Accept json
// @Success 201 {json} short URL json
// @Failure 500 internal error if can't decode request body
// @Failure 401 error if user unauthorized
// @Failure 409 error if URL already exists in DB
func (h *Handlers) Shorten() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		logger.Log.Debug("decoding request")

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		userID := auth.UIDFromContext(ctx)
		if userID == nil {
			//shortenResponse := models.ShortenResponse{
			//	Result: "Пользователь неавторизован",
			//}
			//enc := json.NewEncoder(w)
			//if err := enc.Encode(shortenResponse); err != nil {
			//	logger.Log.Debug("Ошибка создания ответа", zap.Error(err))
			//	w.WriteHeader(http.StatusBadRequest)
			//	return
			//}

			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		var req models.ShortenRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))

			//shortenResponse := models.ShortenResponse{
			//	Result: "cannot decode request JSON body",
			//}
			//enc := json.NewEncoder(w)
			//if err := enc.Encode(shortenResponse); err != nil {
			//	logger.Log.Debug("Ошибка создания ответа", zap.Error(err))
			//	w.WriteHeader(http.StatusBadRequest)
			//	return
			//}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hashID := shortURL(req.URL)
		shortID, err := h.storage.SaveURL(r.Context(), req.URL, hashID, userID)
		if err != nil {
			logger.Log.Debug("Ошибка добавления данных", zap.Error(err))

			var errConflict *storage.URLConflictError
			if errors.As(err, &errConflict) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)

				shortenResponse := models.ShortenResponse{
					Result: fmt.Sprintf("%s/%s", config.Options.FlagShortURL, shortID),
				}
				enc := json.NewEncoder(w)
				if err := enc.Encode(shortenResponse); err != nil {
					logger.Log.Debug("Ошибка создания ответа", zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				return
			} else {

				w.WriteHeader(http.StatusBadRequest)

				shortenResponse := models.ShortenResponse{
					Result: "",
				}
				enc := json.NewEncoder(w)
				if err := enc.Encode(shortenResponse); err != nil {
					logger.Log.Debug("Ошибка создания ответа", zap.Error(err))
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				return
			}
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
		}

		resp := fmt.Sprintf("%s/%s", config.Options.FlagShortURL, shortID)

		shortenResponse := models.ShortenResponse{
			Result: resp,
		}

		enc := json.NewEncoder(w)
		if err := enc.Encode(shortenResponse); err != nil {
			logger.Log.Debug("Ошибка создания ответа", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("sending HTTP 200 response")
	}

	return http.HandlerFunc(fn)
}

// SaveBatch handler for creating a shortened URL based on the original one
// @Accept json
// @Success 201 {json} list of short URLs json
// @Failure 400 bad request error if can't decode request body
// @Failure 401 error if user unauthorized
func (h *Handlers) SaveBatch() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var batchReq []models.BatchShortenRequest

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		userID := auth.UIDFromContext(ctx)
		if userID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err := json.NewDecoder(r.Body).Decode(&batchReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var shortURLs []models.StorageURL

		for _, value := range batchReq {
			if _, err := url.ParseRequestURI(value.OriginalURL); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hashID, err := h.storage.GetURL(value.OriginalURL)
			if err != nil {
				logger.Log.Debug("error get url response", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if hashID == "" {
				hashID = shortURL(value.OriginalURL)
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

		urls, err := h.storage.SaveBatchURL(r.Context(), shortURLs, userID)
		if err != nil {
			logger.Log.Debug("error urls save", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		res := make([]models.BatchShortenResponse, 0, len(urls))

		for i, shortURL := range urls {
			res = append(res, models.BatchShortenResponse{
				CorrelationID: batchReq[i].CorrelationID,
				ShortURL:      fmt.Sprintf("%s/%s", config.Options.FlagShortURL, shortURL),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		//w.Header().Set("Location", config.Options.FlagShortURL)
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err := enc.Encode(res); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("sending HTTP 200 response")
	}

	return http.HandlerFunc(fn)
}
