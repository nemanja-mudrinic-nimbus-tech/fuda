package board

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

type tokenKey struct{}

func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}

func TokenFrom(ctx context.Context) string {
	token, _ := ctx.Value(tokenKey{}).(string)
	return token
}

func tokenDigest(ctx context.Context) string {
	sum := sha256.Sum256([]byte(TokenFrom(ctx)))
	return hex.EncodeToString(sum[:])
}
