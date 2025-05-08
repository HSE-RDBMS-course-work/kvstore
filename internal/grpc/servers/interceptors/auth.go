package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"kvstore/internal/grpc/internal"
)

func NewAuth(username, password string, noAuthMethods []string) grpc.UnaryServerInterceptor {
	ignore := make(map[string]struct{}, len(noAuthMethods))
	for _, method := range noAuthMethods {
		ignore[method] = struct{}{}
	}

	return func(ctx context.Context, in any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := ignore[info.FullMethod]; ok {
			return handler(ctx, in)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		incomingUsername := md[internal.UsernameMetaDataKey]
		if len(incomingUsername) == 0 {
			return nil, status.Error(codes.Unauthenticated, "username missing")
		}

		incomingPassword := md[internal.PasswordMetaDataKey]
		if len(incomingPassword) == 0 {
			return nil, status.Error(codes.Unauthenticated, "password missing")
		}

		if incomingUsername[0] != username || incomingPassword[0] != password {
			return nil, status.Error(codes.Unauthenticated, "username or password mismatch")
		}

		return handler(ctx, in)
	}
}
