package main

import (
	"context"
	"errors"
	"github.com/hashicorp/go-hclog"
	"kvstore/internal/config"
	core "kvstore/internal/core"
	"kvstore/internal/grpc/clients"
	"kvstore/internal/grpc/servers"
	"kvstore/internal/raft"
	"kvstore/internal/sl"
	"log"
	"log/slog"
	"os"
	"os/signal"
)

// рефактор raft.New todo передавать таймаут для выборов итд

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	conf, err := config.Read()
	if err != nil {
		log.Fatalf("cannot read config file: %s", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, conf.Logger()))

	cl := logger.With(sl.Component("di"))

	store, err := core.NewStore(logger, conf.Store())
	if err != nil {
		cl.Error("cannot create store", sl.Error(err))
		return
	}

	existLeader, err := clients.NewRaftClient(conf.ExistingRaftClient())
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

	hcLogger := hclog.New(conf.HashicorpLogger())

	r, recovered, err := raft.New(logger, hcLogger, fsm, conf.Raft())
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

	healthServer := servers.NewHealthServer()
	healthServer.RegisterTo(srv.Server)

	go func() {
		if err := clusterNode.Run(ctx, recovered); err != nil {
			cl.Error("cannot start cluster node", sl.Error(err))
			stop()
		}
	}()

	go func() {
		if err := srv.Run(); err != nil {
			cl.Error("cannot start server", sl.Error(err))
			stop()
		}
	}()

	go func() {
		distributedStore.RunCleaning(ctx)
	}()

	<-ctx.Done()

	if err := srv.Shutdown(); err != nil {
		cl.Error("cannot shutdown server", sl.Error(err))
	}

	if err := clusterNode.Shutdown(); err != nil {
		logger.Error("cannot shutdown cluster", sl.Error(err))
	}
}
