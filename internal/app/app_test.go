package app

import (
	"context"
	"errors"
	"github.com/romanp1989/go-shortener/internal/config"
	"github.com/romanp1989/go-shortener/internal/handlers"
	"github.com/romanp1989/go-shortener/internal/route"
	"github.com/romanp1989/go-shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestNewApp(t *testing.T) {
	_ = config.ParseFlags()
	s := storage.Init(config.Options.FlagDatabaseDsn, config.Options.FlagFileStorage)
	handler := handlers.New(*s)
	deleteHandler, _ := handlers.NewDelete(s)
	r := route.New(handler, deleteHandler)

	tests := []struct {
		name string
		want *App
	}{
		{
			name: "Success Run App",
			want: &App{
				flagRunPort: ":8080",
				chi:         r,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewApp(); !assert.Equal(t, got.Addr, tt.want.flagRunPort) {
				t.Errorf("NewApp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunServer(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Success Run Server",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewApp()

			serviceRunning := make(chan struct{})
			serviceDone := make(chan struct{})

			go func() {
				close(serviceRunning)

				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					t.Errorf("RunServer() error = %v, want %v", err.Error(), http.ErrServerClosed)
				}

				defer close(serviceDone)
			}()

			<-serviceRunning
			err := server.Shutdown(context.Background())
			if err != nil {
				return
			}
			<-serviceDone
		})
	}
}
