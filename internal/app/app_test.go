package app

import (
	"github.com/stretchr/testify/require"
	"syscall"
	"testing"
	"time"
)

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
			go func() {
				time.Sleep(1 * time.Second)
				_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}()

			err := RunServer()
			require.NoError(t, err)
		})
	}
}
