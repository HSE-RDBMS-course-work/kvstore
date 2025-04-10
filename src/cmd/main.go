package main

import (
	"context"
	grpcapi "kvstore/internal/api/grpc"
	"kvstore/internal/app"
	"kvstore/internal/config"
	"kvstore/internal/core"
	"kvstore/internal/raft"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

//todo repositry https://github.com/otoolep/hraftd
//todo grpc https://github.com/Jille/raft-grpc-example/blob/master/main.go
//todo package for getters with no "Get"
//todo логирование во всех модулях проверить как работет заинджектить свое
//todo нормальный конфиг
//todo прокинуть volume и протестить

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("cannot read config file: %s", err)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	})

	logger := slog.New(handler)

	store := core.NewStore()
	fsm := raft.NewFSM(store, logger)

	raftNode := app.StartRaftNode(&cfg.RaftConfig, fsm)

	//todo rename it
	distStore := raft.NewStore(store, raftNode)

	storeAPI := grpcapi.NewKVStoreServer(distStore)
	raftAPI := grpcapi.NewRaftServer(distStore)

	server := app.StartGRPCServer(&cfg.GRPCServerConfig, storeAPI, raftAPI, logger)

	<-ctx.Done()

	server.GracefulStop()
	raftNode.Shutdown()
}
