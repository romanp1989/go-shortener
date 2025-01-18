package auth

import (
	"github.com/gofrs/uuid"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestNewCookie(t *testing.T) {
	type args struct {
		w         http.ResponseWriter
		userID    *uuid.UUID
		secretKey string
	}

	w := httptest.NewRecorder()

	tests := []struct {
		name string
		args args
	}{
		{
			name: "NewCookie_Success",
			args: args{
				w:         w,
				userID:    &uuid.UUID{},
				secretKey: "secret_key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			NewCookie(tt.args.w, tt.args.userID, tt.args.secretKey)
		})
	}
}

func TestEnsureRandom(t *testing.T) {
	tests := []struct {
		name    string
		wantRes uuid.UUID
	}{
		{
			name:    "EnsureRandom_Success",
			wantRes: uuid.UUID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			EnsureRandom()
		})
	}
}

func TestGetUserID(t *testing.T) {
	type args struct {
		userID    *uuid.UUID
		secretKey string
	}
	userID := uuid.UUID{}

	tests := []struct {
		name    string
		args    args
		want    *uuid.UUID
		wantErr bool
	}{
		{
			name: "GetUserID_Success",
			args: args{
				userID:    &userID,
				secretKey: "verycomplexsecretkey",
			},
			want:    &userID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := CreateToken(tt.args.userID, tt.args.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := GetUserID(token, tt.args.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUserID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
