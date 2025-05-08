package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"kvstore/internal/grpc/internal"
)

func NewAuth(username, password string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.Pairs(
			internal.UsernameMetaDataKey, username,
			internal.PasswordMetaDataKey, password,
		)

		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
