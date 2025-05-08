package interceptors

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewAuth(username, password string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, in any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		incomingUsername := md["username"]
		if len(incomingUsername) == 0 {
			return nil, status.Error(codes.Unauthenticated, "username missing")
		}

		incomingPassword := md["password"]
		if len(incomingPassword) == 0 {
			return nil, status.Error(codes.Unauthenticated, "password missing")
		}

		if incomingUsername[0] != username || incomingPassword[0] != password {
			return nil, status.Error(codes.Unauthenticated, "username or password mismatch")
		}

		return handler(ctx, in)
	}
}
