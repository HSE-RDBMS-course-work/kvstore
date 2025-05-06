package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"kvstore/internal/core"
	"kvstore/internal/grpc/servers"
	"kvstore/internal/raft"
	"os"
	"time"
)

type Config struct {
	Host             string     `mapstructure:"host"`
	PublicPort       string     `mapstructure:"public_port"`
	InternalPort     string     `mapstructure:"internal_port"`
	Username         string     `mapstructure:"username"`
	Password         string     `mapstructure:"password"`
	JoinTo           string     `mapstructure:"join_to"`
	DataPath         string     `mapstructure:"data_path"`
	StoreConfig      Store      `mapstructure:"storage"`
	GRPCServerConfig GRPCServer `mapstructure:"grpc_server"`
	RaftConfig       Raft       `mapstructure:"raft"`
}

type Store struct {
	CleanInterval    time.Duration `mapstructure:"clean_interval"`
	MaxCleanDuration time.Duration `mapstructure:"max_clean_duration"`
	InitialCapacity  int64         `mapstructure:"initial_capacity"`
}

type GRPCServer struct {
	Timeout time.Duration `mapstructure:"timeout"`
}

type Raft struct {
	Advertise       string        `mapstructure:"advertise"`
	NodeID          string        `mapstructure:"node_id"`
	DataLocation    string        `mapstructure:"data_location"`
	Timeout         time.Duration `mapstructure:"timeout"`
	MaxPool         int           `mapstructure:"max_pool"`
	SnapshotsRetain int           `mapstructure:"snapshots_retain"`
}

func Read() (*Config, error) {
	vp := viper.New()

	vp.SetConfigFile(*configPath)
	if err := vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file (%s): %w", configPath, err)
	}

	var conf Config
	if err := vp.Unmarshal(&conf); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from file: %w", err)
	}

	if err := conf.expandEnv(); err != nil {
		return nil, fmt.Errorf("cannot expand env varibles in config: %w", err)
	}

	conf.choose(&conf.DataPath, dataPath)
	conf.choose(&conf.Host, host)
	conf.choose(&conf.PublicPort, pPort)
	conf.choose(&conf.InternalPort, iPort)
	conf.choose(&conf.Username, username)
	conf.choose(&conf.Password, password)

	return &conf, nil
}

func (c *Config) Store() core.Config {
	return core.Config{
		CleanInterval:    c.StoreConfig.CleanInterval,
		MaxCleanDuration: c.StoreConfig.MaxCleanDuration,
		InitialCapacity:  c.StoreConfig.InitialCapacity,
	}
}

func (c *Config) Raft() raft.Config {
	return raft.Config{
		RealAddress:       fmt.Sprintf("%s:%s", c.Host, c.InternalPort),
		AdvertisedAddress: c.RaftConfig.Advertise,
		NodeID:            c.RaftConfig.NodeID,
		DataLocation:      c.RaftConfig.DataLocation,
		SnapshotsRetain:   c.RaftConfig.SnapshotsRetain,
		MaxPool:           c.RaftConfig.MaxPool,
		Timeout:           c.RaftConfig.Timeout,
	}
}

func (c *Config) ClusterNode() raft.ClusterNodeConfig {
	return raft.ClusterNodeConfig{
		ID:                raft.ServerID(c.RaftConfig.NodeID),
		RealAddress:       raft.ServerAddress(c.RaftConfig.Advertise),
		AdvertisedAddress: raft.ServerAddress(c.RaftConfig.Advertise),
	}
}

func (c *Config) GRPCServer() servers.Config {
	return servers.Config{
		Address: fmt.Sprintf("%s:%s", c.Host, c.PublicPort),
		Timeout: c.GRPCServerConfig.Timeout,
	}
}

func (c *Config) choose(target *string, flag *string) {
	if target == nil || flag == nil {
		return
	}

	if *target == "" {
		*target = *flag
	}
}

func (c *Config) expandEnv() error {
	bytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("cannot marshal config to json: %w", err)
	}

	expanded := os.ExpandEnv(string(bytes))

	if err := json.Unmarshal([]byte(expanded), c); err != nil {
		return fmt.Errorf("cannot unmarshal config from json: %w", err)
	}

	return nil
}
