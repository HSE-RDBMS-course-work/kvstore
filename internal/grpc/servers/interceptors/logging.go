package interceptors

import (
	"context"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"log/slog"
)

type slogWrapper struct {
	logger *slog.Logger
}

func (sw *slogWrapper) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	sw.logger.Log(ctx, slog.Level(level), msg, fields)
}

func NewLogging(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(&slogWrapper{
		logger: logger,
	})
}
