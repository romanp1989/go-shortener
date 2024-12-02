package auth

import (
	"context"
	"github.com/gofrs/uuid"
)

type ctxAuthKey string

// AuthKey User authorization key
const AuthKey ctxAuthKey = "auth"

// Context Function add authorization key to context
func Context(parent context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(parent, AuthKey, uid)
}

// UIDFromContext Function get user authorization key from context
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
