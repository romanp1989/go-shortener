package middlewares

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/compress"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := compress.NewCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		logger.Log.Info("Request info",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
		)

		responseData := &logger.ResponseData{
			Status: 0,
			Size:   0,
		}
		lw := logger.LoggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			ResponseData:   responseData,
		}

		//Запуск оригинального handler
		h.ServeHTTP(&lw, r) // внедряем реализацию http.ResponseWriter

		duration := time.Since(start)

		logger.Log.Info("Response info",
			zap.Int("status", responseData.Status), // получаем перехваченный код статуса ответа
			zap.Duration("duration", duration),
			zap.Int("size", responseData.Size), // получаем перехваченный размер ответа
		)
	})
}

func AuthMiddlewareSet(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid *uuid.UUID

		cookie, err := r.Cookie("auth")

		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		if cookie == nil {
			userID := auth.EnsureRandom()
			uid = &userID

			auth.NewCookie(w, uid)
		} else if cookie != nil {
			uid, _ = auth.GetUserID(cookie.Value)
		}

		if uid == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := auth.Context(r.Context(), *uid)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}

func AuthMiddlewareRead(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid *uuid.UUID

		cookie, err := r.Cookie("auth")

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if cookie == nil {
			userID := auth.EnsureRandom()
			uid = &userID

			auth.NewCookie(w, uid)
		} else if cookie != nil {
			uid, _ = auth.GetUserID(cookie.Value)
		}

		if uid == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := auth.Context(r.Context(), *uid)
		r = r.WithContext(ctx)

		h.ServeHTTP(w, r)
	})
}
