package authctx

import "context"

type ctxKey string

const (
	UserIDKey ctxKey = "user_id"
)

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func UserID(ctx context.Context) (string, bool) {
	v := ctx.Value(UserIDKey)
	s, ok := v.(string)
	return s, ok
}