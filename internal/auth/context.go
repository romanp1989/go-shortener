package auth

import (
	"context"
	"github.com/gofrs/uuid"
)

type ctxAuthKey string

const authKey ctxAuthKey = "auth"

func Context(parent context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(parent, authKey, uid)
}

func UIDFromContext(ctx context.Context) *uuid.UUID {
	val := ctx.Value(authKey)
	if val == nil {
		return nil
	}
	uid := val.(uuid.UUID)
	return &uid
}
