package middlewares

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/compress"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
	"time"
)

// Middleware middleware struct
type Middleware struct {
	Cfg        *config.ConfigENV
	JwtService *auth.JWTService
}

// GzipMiddleware Middleware for archiving the hanlders response
func (m Middleware) GzipMiddleware(h http.Handler) http.Handler {
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

// WithLogging Middleware for logging request
func (m Middleware) WithLogging(h http.Handler) http.Handler {
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

// AuthMiddlewareSet Middleware for set authorization cookie
func (m Middleware) AuthMiddlewareSet(h http.Handler) http.Handler {
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
			userID := m.JwtService.EnsureRandom()
			uid = &userID

			m.newCookie(w, uid, m.Cfg.SecretKey)
		} else if cookie.Value != "" {
			uid, _ = m.JwtService.GetUserID(cookie.Value)
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

// AuthMiddlewareRead Middleware for authorization users
func (m Middleware) AuthMiddlewareRead(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var uid *uuid.UUID

		cookie, err := r.Cookie("auth")

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if cookie == nil {
			userID := m.JwtService.EnsureRandom()
			uid = &userID

			m.newCookie(w, uid, m.Cfg.SecretKey)
		} else if cookie.Value != "" {
			uid, _ = m.JwtService.GetUserID(cookie.Value)
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

// ValidateSubnet validate user ip for internal access
func (m Middleware) ValidateSubnet(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var trustedSubnets string
		var ipNet *net.IPNet
		var err error

		trustedSubnets = m.Cfg.TrustedSubnet
		if trustedSubnets != "" {
			_, ipNet, err = net.ParseCIDR(trustedSubnets)
			if err != nil {
				logger.Log.Error("Parse error trusted subnet config: %v", zap.String("error", err.Error()))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		clientIPHeader := r.Header.Get("X-Real-IP")
		ip := net.ParseIP(clientIPHeader)

		if ip == nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if !ipNet.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// NewCookie Function add new authorization cookie
func (m Middleware) newCookie(w http.ResponseWriter, userID *uuid.UUID, secretKey string) {

	token, err := m.JwtService.CreateToken(userID, secretKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:  "auth",
		Value: token,
		Path:  "/",
	}

	http.SetCookie(w, cookie)
}
