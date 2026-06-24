package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func authContext(token string) context.Context {
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))
}
