package handlers

import (
	"context"
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/logger"
	"github.com/romanp1989/go-shortener/internal/storage"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type itemDelete struct {
	urls   []string
	userID *uuid.UUID
}

type DeleteBatch struct {
	storage   *storage.Storage
	inChan    chan itemDelete
	closeChan chan struct{}
	size      int
}

func NewDelete(store *storage.Storage) (*DeleteBatch, error) {
	d := &DeleteBatch{
		storage:   store,
		inChan:    make(chan itemDelete, 1024),
		closeChan: make(chan struct{}),
		size:      500,
	}

	for i := 0; i < 2; i++ {
		go d.process()
	}

	return d, nil
}

func (d *DeleteBatch) DeleteURLs() http.HandlerFunc {
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

		go d.add(userID, urls)

		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	}

	return http.HandlerFunc(fn)
}

func (d *DeleteBatch) add(userID *uuid.UUID, urls []string) {

	for i := 0; i < len(urls); i += d.size {
		endSlice := i + d.size
		if endSlice > len(urls) {
			endSlice = len(urls)
		}

		d.inChan <- itemDelete{
			urls:   urls[i:endSlice],
			userID: userID,
		}
	}
}

func (d *DeleteBatch) close() {
	close(d.closeChan)
}

func (d *DeleteBatch) process() {
	for {
		select {
		case batch, ok := <-d.inChan:
			if !ok {
				return
			}

			if err := d.storage.DeleteUrlsBatch(context.Background(), batch.userID, batch.urls); err != nil {
				logger.Log.Error("Ошибка при удалении url: %v, %v", zap.String("userID", batch.userID.String()), zap.Error(err))
			}

		case <-d.closeChan:
			close(d.inChan)
			logger.Log.Debug("Закрыт канал для удаления url")
			return
		}
	}
}
