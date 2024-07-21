package handlers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/romanp1989/go-shortener/internal/config"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var urlStore = make(map[string]string)

func ShortenerRouter() chi.Router {
	r := chi.NewRouter()

	r.Post("/", Encode)
	r.Get("/{id}", Decode)

	return r
}

func Encode(w http.ResponseWriter, r *http.Request) {
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

func shortURL(url string) string {
	sum := md5.Sum([]byte(url))
	encoded := base64.StdEncoding.EncodeToString(sum[:])
	encoded = strings.Replace(encoded, "/", "", -1)[:8]

	return encoded
}

func Decode(w http.ResponseWriter, r *http.Request) {
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
