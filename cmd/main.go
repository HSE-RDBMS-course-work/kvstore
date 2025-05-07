package main

import (
	"context"
	"errors"
	"kvstore/internal/config"
	"kvstore/internal/core"
	"kvstore/internal/grpc/clients"
	"kvstore/internal/grpc/servers"
	"kvstore/internal/raft"
	"kvstore/internal/sl"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

//todo repositry https://github.com/otoolep/hraftd
//todo нормальный конфиг
//todo прокинуть volume и протестить

//todo авторизация, ttl, healthcheck, рефактор raft.New, логирование во всех модулях проверить как работет заинджектить свое (slog.NewLog)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf, err := config.Read()
	if err != nil {
		log.Fatalf("cannot read config file: %s", err)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: false,
	})
	logger := slog.New(handler)

	cl := logger.With(sl.Component("di"))

	store, err := core.NewStore(logger, conf.Store())
	if err != nil {
		cl.Error("cannot create store", sl.Error(err))
		return
	}

	existLeader, err := clients.NewRaftClient(conf.JoinTo())
	if errors.Is(err, clients.ErrAddressIsEmpty) {
		cl.Warn("no cluster node to join to cluster was provided. It is ok if you want to bootstrap cluster")
	}
	if err != nil && !errors.Is(err, clients.ErrAddressIsEmpty) {
		cl.Error("cannot create raft client", sl.Error(err))
		return
	}

	fsm, err := raft.NewFSM(logger, store)
	if err != nil {
		cl.Error("cannot create FSM", sl.Error(err))
		return
	}

	r, recovered, err := raft.New(logger, fsm, os.Stderr, conf.Raft()) //todo в конфиге какие то значение дефолтные в config. какие видмо в конструкторе (SnapshotsRetain сделать где то дефолтное значение) возможно изавитьяс от дефолтных значений во флагах
	if err != nil {
		cl.Error("cannot create raft instance", sl.Error(err))
		return
	}

	distributedStore, err := raft.NewStore(logger, r, store)
	if err != nil {
		cl.Error("cannot create distributed store", sl.Error(err))
		return
	}

	clusterNode, err := raft.NewClusterNode(logger, r, existLeader, conf.ClusterNode())
	if err != nil {
		cl.Error("cannot create cluster node", sl.Error(err))
		return
	}

	srv, err := servers.New(logger, conf.GRPCServer())
	if err != nil {
		cl.Error("cannot create server", sl.Error(err))
		return
	}

	raftServer, err := servers.NewRaftServer(clusterNode)
	if err != nil {
		cl.Error("cannot create raft grpc server", sl.Error(err))
		return
	}
	raftServer.RegisterTo(srv.Server)

	kvstoreServer, err := servers.NewKVStoreServer(distributedStore)
	if err != nil {
		cl.Error("cannot create kvstore grpc server", sl.Error(err))
		return
	}
	kvstoreServer.RegisterTo(srv.Server)

	go func() { //todo нормально сделать
		if recovered {
			return
		}
		if err := clusterNode.Run(ctx); err != nil {
			cl.Error("cannot start cluster node", sl.Error(err))
			stop()
		}
	}()

	go func() {
		if err := srv.Run(); err != nil {
			logger.Error("cannot start server", sl.Error(err))
			stop()
		}
	}()

	<-ctx.Done()

	srv.Stop()

	if err := clusterNode.Shutdown(); err != nil {
		logger.Error("cannot shutdown cluster", sl.Error(err))
	}
}
