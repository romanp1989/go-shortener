package shortenerservice

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
func (s *ShortenerService) DeleteURLs(userID *uuid.UUID, urls []string) {
	go s.add(userID, urls)
}

// add function for add user to delete channel
func (s *ShortenerService) add(userID *uuid.UUID, urls []string) {
	for i := 0; i < len(urls); i += s.size {
		endSlice := i + s.size
		if endSlice > len(urls) {
			endSlice = len(urls)
		}

		s.inChan <- itemDelete{
			urls:   urls[i:endSlice],
			userID: userID,
		}
	}
}

// process function starts the goroutine for delete user's urls
func (s *ShortenerService) process() {
	for {
		select {
		case batch, ok := <-s.inChan:
			if !ok {
				return
			}

			if err := s.storage.DeleteUrlsBatch(context.Background(), batch.userID, batch.urls); err != nil {
				logger.Log.Error("Ошибка при удалении url: %v, %v", zap.String("userID", batch.userID.String()), zap.Error(err))
			}
		case <-s.closeChan:
			close(s.inChan)
			logger.Log.Debug("Закрыт канал для удаления url")
			return
		}
	}
}
