package shortener_service

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/logger"
	"go.uber.org/zap"
)

type itemDelete struct {
	urls   []string
	userID *uuid.UUID
}

// DeleteURLs function for delete urls
func (d *ShortenerService) DeleteURLs(userID *uuid.UUID, urls []string) {
	go d.add(userID, urls)
}

// add function for add user to delete channel
func (d *ShortenerService) add(userID *uuid.UUID, urls []string) {
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

// process function starts the goroutine for delete user's urls
func (d *ShortenerService) process() {
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
