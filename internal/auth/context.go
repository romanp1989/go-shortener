package auth

import (
	"context"
	"github.com/gofrs/uuid"
)

type ctxAuthKey string

const AuthKey ctxAuthKey = "auth"

func Context(parent context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(parent, AuthKey, uid)
}

func UIDFromContext(ctx context.Context) *uuid.UUID {
	val, ok := ctx.Value(AuthKey).(uuid.UUID)
	if !ok {
		return nil
	}
	if val.IsNil() {
		return nil
	}
	uid := val
	return &uid
}
