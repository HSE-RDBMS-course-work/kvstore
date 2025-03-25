package main

import (
	"context"
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
//todo package for getters with no "Get"

//todo объеденить три ноды
//по идее они не отпралвяют друг другу никакие данные

type Config struct {
	RaftConfig       app.RaftConfig       `yaml:"raft"`
	GRPCServerConfig app.GRPCServerConfig `yaml:"grpc_server"`
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
	fsm := raft.NewFSM(store)
	distStore := raft.NewStore(store)

	api := grpcapi.NewServer(distStore)

	cfgFile, err := os.OpenFile("config.yaml", os.O_RDONLY, 0666)
	if err != nil {
		log.Fatalf("cannot open config file: %s", err)
	}

	cfg := Config{}
	if err := yaml.NewDecoder(cfgFile).Decode(&cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	server := app.StartGRPCServer(&cfg.GRPCServerConfig, api, logger)
	raftNode := app.StartRaftNode(&cfg.RaftConfig, fsm)

	<-ctx.Done()

	server.GracefulStop()
	raftNode.Shutdown()
}
