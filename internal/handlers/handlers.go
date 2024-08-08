package handlers

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/models"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Handlers struct {
	storage *storage.Storage
}

var urlStore = make(map[string]string)

func New(storage *storage.Storage) Handlers {
	return Handlers{
		storage: storage,
	}
}

func (h *Handlers) Encode() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
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

		hashID, err := h.storage.GetURL(stringURI)
		if err != nil {
			logger.Log.Debug("error get url response", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if hashID == "" {
			hashID = shortURL(stringURI)
			err := h.storage.SaveURL(hashID, stringURI)
			if err != nil {
				log.Fatal(err)
			}
		}

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

func (h *Handlers) Shorten() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		logger.Log.Debug("decoding request")

		var req models.ShortenRequest
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		hashID, err := h.storage.GetURL(req.URL)
		if err != nil {
			logger.Log.Debug("error get url response", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if hashID == "" {
			hashID = shortURL(req.URL)
			err := h.storage.SaveURL(hashID, req.URL)
			if err != nil {
				log.Fatal(err)
			}
		}

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
