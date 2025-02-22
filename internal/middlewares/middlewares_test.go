package middlewares

import (
	"github.com/gofrs/uuid"
	"github.com/romanp1989/go-shortener/internal/auth"
	"github.com/romanp1989/go-shortener/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_newCookie(t *testing.T) {
	type fields struct {
		Cfg        *config.ConfigENV
		JwtService *auth.JWTService
	}
	type args struct {
		w         http.ResponseWriter
		userID    *uuid.UUID
		secretKey string
	}

	w := httptest.NewRecorder()

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "NewCookie_Success",
			fields: fields{
				Cfg:        &config.ConfigENV{},
				JwtService: &auth.JWTService{},
			},
			args: args{
				w:         w,
				userID:    &uuid.UUID{},
				secretKey: "verycomplexsecretkey",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Middleware{
				Cfg:        tt.fields.Cfg,
				JwtService: tt.fields.JwtService,
			}
			m.newCookie(tt.args.w, tt.args.userID, tt.args.secretKey)
		})
	}
}
