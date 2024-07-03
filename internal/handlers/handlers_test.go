package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	type want struct {
		statusCode  int
		responseURL string
	}

	tests := []struct {
		name        string
		method      string
		requestBody string
		want        want
	}{
		{
			name:        "Valid URL",
			method:      http.MethodPost,
			requestBody: "https://ya.ru",
			want: want{
				statusCode:  http.StatusCreated,
				responseURL: "http://example.com/6YGS4ZUF",
			},
		},
		{
			name:        "Empty URL",
			method:      http.MethodPost,
			requestBody: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
		{
			name:        "Wrong request type",
			method:      http.MethodPut,
			requestBody: "",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "Некорректный тип запроса\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.requestBody)
			r := httptest.NewRequest(tt.method, "/", body)
			r.Header.Set("Content-Type", "text/plain")

			w := httptest.NewRecorder()

			Encode(w, r)

			result := w.Result()

			resBody, err := io.ReadAll(result.Body)
			defer result.Body.Close()

			require.NoError(t, err)

			assert.Equal(t, tt.want.responseURL, string(resBody))

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

		})
	}
}

func TestDecode(t *testing.T) {
	type want struct {
		statusCode  int
		responseURL string
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "Success redirect",
			want: want{
				statusCode:  http.StatusTemporaryRedirect,
				responseURL: "https://ya.ru",
			},
		},
		{
			name: "Fail redirect",
			want: want{
				statusCode:  http.StatusBadRequest,
				responseURL: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashID := shortURL(tt.want.responseURL)
			if tt.want.responseURL != "" {
				urlStore[hashID] = tt.want.responseURL
			}
			body := httptest.NewRequest(http.MethodGet, "/"+hashID, nil)
			w := httptest.NewRecorder()

			Decode(w, body)

			result := w.Result()

			defer result.Body.Close()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if tt.want.responseURL != "" {
				url := w.Header().Get("Location")
				assert.Equal(t, tt.want.responseURL, url)
			}

		})
	}
}
