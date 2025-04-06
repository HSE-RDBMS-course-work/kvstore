package main

import (
	"context"
	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
	grpcapi "kvstore/internal/api/grpc"
	"kvstore/internal/app"
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

type Config struct {
	RaftConfig       app.RaftConfig       `yaml:"raft"`
	GRPCServerConfig app.GRPCServerConfig `yaml:"grpc_server"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfgFile, err := os.OpenFile("config.yaml", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("cannot open config file: %s", err)
	}

	cfg := Config{}
	if err := yaml.NewDecoder(cfgFile).Decode(&cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("cannot parse environment variables: %s", err)
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
