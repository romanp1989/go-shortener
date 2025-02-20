package auth

import (
	"github.com/gofrs/uuid"
	"reflect"
	"testing"
	"time"
)

func TestJWTService_EnsureRandom(t *testing.T) {
	type fields struct {
		secretKey string
		expire    time.Duration
		TokenName string
	}

	tests := []struct {
		name      string
		jwtFields fields
		wantRes   uuid.UUID
	}{
		{
			name: "EnsureRandom_Success",
			jwtFields: fields{
				secretKey: "verycomplexsecretkey",
				expire:    time.Hour * 3,
				TokenName: "auth",
			},
			wantRes: uuid.UUID{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JWTService{
				secretKey: tt.jwtFields.secretKey,
				expire:    tt.jwtFields.expire,
				TokenName: tt.jwtFields.TokenName,
			}

			j.EnsureRandom()
		})
	}
}

func TestJWTService_GetUserID(t *testing.T) {
	type fields struct {
		secretKey string
		expire    time.Duration
		TokenName string
	}

	type args struct {
		userID    *uuid.UUID
		jwtFields fields
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
				userID: &userID,
				jwtFields: fields{
					secretKey: "verycomplexsecretkey",
					expire:    time.Hour * 3,
					TokenName: "auth",
				},
			},
			want:    &userID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JWTService{
				secretKey: tt.args.jwtFields.secretKey,
				expire:    tt.args.jwtFields.expire,
				TokenName: tt.args.jwtFields.TokenName,
			}

			token, err := j.CreateToken(tt.args.userID, tt.args.jwtFields.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got, err := j.GetUserID(token)
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
