package main

import (
	"context"
	grpcapi "github.com/HSE-RDBMS-course-work/kvstore/internal/api/grpc"
	"github.com/HSE-RDBMS-course-work/kvstore/internal/app"
	"github.com/HSE-RDBMS-course-work/kvstore/internal/core"
	"github.com/HSE-RDBMS-course-work/kvstore/internal/fsm"
	"log/slog"
	"os"
	"os/signal"
)

//todo package for getters with no "Get"

//todo объеденить три ноды
//по идее они не отпралвяют друг другу никакие данные

type Config struct {
	app.RaftConfig       `yaml:"raft"`
	app.GRPCServerConfig `yaml:"grpc_server"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})

	logger := slog.New(handler)

	store := core.NewStore()
	fsmStore := fsm.NewFSM(store)

	api := grpcapi.NewServer(store)

	server := app.StartGRPCServer(nil, api, logger)
	raft := app.StartRaftNode(nil, fsmStore)

	<-ctx.Done()

	server.GracefulStop()
	raft.Shutdown()
}
