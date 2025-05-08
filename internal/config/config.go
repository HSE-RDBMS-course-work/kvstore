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
	ConnectionTimeout time.Duration `mapstructure:"connection_timeout"`
}

type Raft struct {
	Advertise       string        `mapstructure:"advertise"`
	NodeID          string        `mapstructure:"-"`
	Timeout         time.Duration `mapstructure:"timeout"`
	MaxPool         int           `mapstructure:"max_pool"`
	SnapshotsRetain int           `mapstructure:"snapshots_retain"`
}

func Read() (*Config, error) {
	vp := viper.New()

	vp.SetConfigFile(*configPath)
	if err := vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("cannot read config file (%s): %w", *configPath, err)
	}

	var c Config
	if err := vp.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from file: %w", err)
	}

	if err := c.expandEnv(); err != nil {
		return nil, fmt.Errorf("cannot expand env varibles in config: %w", err)
	}

	c.choose(&c.DataPath, dataPath)
	c.choose(&c.Host, host)
	c.choose(&c.PublicPort, pPort)
	c.choose(&c.InternalPort, iPort)
	c.choose(&c.Username, username)
	c.choose(&c.Password, password)
	c.choose(&c.RaftConfig.Advertise, advertise) //todo move this field to root of the config

	c.RaftConfig.NodeID = c.RaftConfig.Advertise

	if c.RaftConfig.Advertise == "" {
		c.RaftConfig.Advertise = c.address("localhost", c.InternalPort)
	}

	return &c, nil
}

func (c *Config) Store() core.Config {
	return core.Config{
		CleanInterval:    c.StoreConfig.CleanInterval,
		MaxCleanDuration: c.StoreConfig.MaxCleanDuration,
		InitialCapacity:  c.StoreConfig.InitialCapacity,
	}
}

func (c *Config) JoinTo() string {
	return *joinTo
}

func (c *Config) Raft() raft.Config {
	return raft.Config{
		RealAddress:       c.address(c.Host, c.InternalPort),
		AdvertisedAddress: c.RaftConfig.Advertise,
		NodeID:            c.RaftConfig.NodeID,
		DataLocation:      c.DataPath,
		SnapshotsRetain:   c.RaftConfig.SnapshotsRetain,
		MaxPool:           c.RaftConfig.MaxPool,
		Timeout:           c.RaftConfig.Timeout,
	}
}

func (c *Config) ClusterNode() raft.ClusterNodeConfig {
	return raft.ClusterNodeConfig{
		ID:                raft.ServerID(c.RaftConfig.NodeID),
		RealAddress:       raft.ServerAddress(c.address(c.Host, c.InternalPort)),
		AdvertisedAddress: raft.ServerAddress(c.RaftConfig.Advertise),
		BootstrapCluster:  *joinTo == "",
	}
}

func (c *Config) GRPCServer() servers.Config {
	return servers.Config{
		Address:           c.address(c.Host, c.PublicPort),
		ConnectionTimeout: c.GRPCServerConfig.ConnectionTimeout,
	}
}

func (c *Config) choose(target *string, flag *string) {
	if target == nil || flag == nil {
		return
	}

	if *flag != "" {
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

func (c *Config) address(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}
