package grpc

import (
	"context"
	"strings"

	"google.golang.org/grpc/metadata"
)

func extractAccessToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		values = md.Get("Authorization")
	}
	if len(values) == 0 {
		return ""
	}

	token := strings.TrimSpace(values[0])
	lower := strings.ToLower(token)
	if strings.HasPrefix(lower, "bearer ") {
		return strings.TrimSpace(token[7:])
	}
	return token
}
