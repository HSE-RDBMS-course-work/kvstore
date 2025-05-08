package interceptors

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"kvstore/internal/sl"
	"log/slog"
	"runtime/debug"
)

func NewRecovery(logger *slog.Logger) grpc.UnaryServerInterceptor {
	recoveryLogger := logger.With(sl.Component("grpc.Recovery"))

	recoveryFunc := func(ctx context.Context, p any) (err error) {
		recoveryLogger.ErrorContext(ctx, "panic while handling grpc request",
			sl.Panic(p),
			slog.String("trace", string(debug.Stack())),
		)
		return status.Error(codes.Internal, "internal servers error")
	}

	return recovery.UnaryServerInterceptor(
		recovery.WithRecoveryHandlerContext(recoveryFunc),
	)
}
