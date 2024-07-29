package handlers

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/compress"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var urlStore = make(map[string]string)

func ShortenerRouter() chi.Router {
	r := chi.NewRouter()

	r.Post("/", logger.WithLogging(compress.GzipMiddleware(Encode())))
	r.Get("/{id}", logger.WithLogging(Decode()))
	r.Post("/api/shorten", logger.WithLogging(compress.GzipMiddleware(Shorten())))

	return r
}

func Encode() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Некорректный тип запроса", http.StatusBadRequest)
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

		urlStore[hashID] = stringURI

		resp := fmt.Sprintf("%s/%s", config.Options.FlagShortURL, hashID)

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

func Decode() http.HandlerFunc {
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

		if fullURL, ok := urlStore[id]; ok {
			http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
			return
		}

		http.Error(w, "Не найден url для указанного ID", http.StatusNotFound)
	}
	return http.HandlerFunc(fn)
}

func Shorten() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Debug("decoding request")

		var req models.ShortenRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hashID := shortURL(req.URL)

		urlStore[hashID] = req.URL

		resp := fmt.Sprintf("%s/%s", config.Options.FlagShortURL, hashID)

		shortenResponse := models.ShortenResponse{
			Result: resp,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err := enc.Encode(shortenResponse); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		logger.Log.Debug("sending HTTP 200 response")
	}

	return http.HandlerFunc(fn)
}
