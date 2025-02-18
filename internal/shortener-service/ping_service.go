package shortener_service

import (
	"context"
)

// PingDb function for ping server connection
func (s *ShortenerService) PingDB(ctx context.Context) error {
	if err := s.storage.Ping(ctx); err != nil {
		return err
	}

	return nil
}
