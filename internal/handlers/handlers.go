package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/models"
	shortener_service "github.com/romanp1989/go-shortener/internal/shortener-service"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Handlers handlers
type Handlers struct {
	appService *shortener_service.ShortenerService
}

// New Factory for create handlers
func New(appService *shortener_service.ShortenerService) Handlers {
	return Handlers{
		appService: appService,
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

		if _, err = url.ParseRequestURI(stringURI); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		shortURL, err := h.appService.Encode(ctx, stringURI)

		if err != nil {
			logger.Log.Debug("Ошибка добавления данных", zap.Error(err))

			var errConflict *storage.URLConflictError
			if errors.As(err, &errConflict) {
				//shortID = errConflict.URL
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(shortURL))
				return
			} else {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))
	}

	return http.HandlerFunc(fn)
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

		fullURL, err := h.appService.Decode(id)
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
			w.WriteHeader(http.StatusUnauthorized)

			return
		}

		var req models.ShortenRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		shortURL, err := h.appService.Shorten(r.Context(), req.URL, userID)
		if err != nil {
			logger.Log.Debug("Ошибка добавления данных", zap.Error(err))

			var errConflict *storage.URLConflictError
			if errors.As(err, &errConflict) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)

				shortenResponse := models.ShortenResponse{
					Result: shortURL,
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

		shortenResponse := models.ShortenResponse{
			Result: shortURL,
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
		var err error
		var resp []models.BatchShortenResponse

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()

		userID := auth.UIDFromContext(ctx)
		if userID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = json.NewDecoder(r.Body).Decode(&batchReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err = h.appService.SaveBatch(ctx, batchReq, userID)

		if err != nil {
			logger.Log.Debug("error urls save", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("sending HTTP 200 response")
	}

	return http.HandlerFunc(fn)
}
