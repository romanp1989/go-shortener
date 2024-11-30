package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInitialize(t *testing.T) {
	type args struct {
		level string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "success_create_logger",
			args:    args{level: "debug"},
			wantErr: false,
		},
		{
			name:    "error_create_logger",
			args:    args{level: "rgdfgdfgfd"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Initialize(tt.args.level); (err != nil) != tt.wantErr {
				t.Errorf("Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoggingResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	responseData := ResponseData{
		Status: 0,
		Size:   0,
	}
	lw := LoggingResponseWriter{
		ResponseWriter: w,
		ResponseData:   &responseData,
	}

	logStr := []byte("some string")
	_, err := lw.Write(logStr)

	if err != nil {
		t.Errorf("Write() error = %v", err)
		return
	}
}

func TestLoggingResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	responseData := ResponseData{
		Status: 0,
		Size:   0,
	}
	lw := LoggingResponseWriter{
		ResponseWriter: w,
		ResponseData:   &responseData,
	}

	lw.WriteHeader(http.StatusCreated)
}
