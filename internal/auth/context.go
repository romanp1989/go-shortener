package auth

import (
	"context"
	"github.com/gofrs/uuid"
)

var ctxAuthKey = "auth"

func Context(parent context.Context, uid uuid.UUID) context.Context {
	return context.WithValue(parent, ctxAuthKey, uid)
}

func UIDFromContext(ctx context.Context) *uuid.UUID {
	val := ctx.Value(ctxAuthKey)
	if val == nil {
		return nil
	}
	uid := val.(uuid.UUID)
	return &uid
}
