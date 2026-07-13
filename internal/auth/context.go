package auth

import "context"

type contextKey string

const userContextKey contextKey = "user"

type Principal struct {
	ID    int64
	Email string
}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, userContextKey, p)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	value := ctx.Value(userContextKey)
	if value == nil {
		return Principal{}, false
	}

	principal, ok := value.(Principal)
	return principal, ok
}
