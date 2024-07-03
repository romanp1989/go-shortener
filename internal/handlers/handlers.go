package handlers

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var urlStore = make(map[string]string)

func Encode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Некорректный тип запроса", http.StatusBadRequest)
	}

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

	resp := fmt.Sprintf("http://%s/%s", r.Host, hashID)

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

	idRaw := r.URL.Path[1:]
	id := strings.Split(idRaw, "/")[0]

	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if fullURL, ok := urlStore[id]; ok {
		http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
		return
	}

	http.Error(w, "Не найден url для указанного ID", http.StatusBadRequest)
}
