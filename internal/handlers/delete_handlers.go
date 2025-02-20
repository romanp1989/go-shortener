package handlers

import (
	"context"
	"encoding/json"
	"github.com/romanp1989/go-shortener/internal/auth"
	"io"
	"net/http"
	"time"
)

// DeleteURLs function for delete urls
func (h *Handlers) DeleteURLs() http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var urls []string

		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		body, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, "Ошибка при парсинге body запроса", http.StatusBadRequest)
			return
		}

		userID := auth.UIDFromContext(ctx)
		if userID == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := json.Unmarshal(body, &urls); err != nil {
			http.Error(w, "Ошибка при парсинге спика url для удаления", http.StatusBadRequest)
			return
		}

		go h.appService.DeleteURLs(userID, urls)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	}

	return http.HandlerFunc(fn)
}
