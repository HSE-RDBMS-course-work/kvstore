package config

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"gopkg.in/yaml.v3"
	"kvstore/internal/core"
	"kvstore/internal/grpc/clients"
	"kvstore/internal/grpc/servers"
	"kvstore/internal/raft"
	"kvstore/internal/sl"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Host             string     `yaml:"host"`
	PublicPort       string     `yaml:"public_port"`
	InternalPort     string     `yaml:"internal_port"`
	Advertise        string     `yaml:"advertise"`
	Username         string     `yaml:"username"`
	Password         string     `yaml:"password"`
	DataPath         string     `yaml:"data_path"`
	LoggerConfig     Logger     `yaml:"logger"`
	StoreConfig      Store      `yaml:"storage"`
	GRPCServerConfig GRPCServer `yaml:"grpc_server"`
	RaftConfig       Raft       `yaml:"raft"`
}

type Logger struct {
	Level int `yaml:"level"`
}

type Store struct {
	CleanInterval    time.Duration `yaml:"clean_interval"`
	MaxCleanDuration time.Duration `yaml:"clean_duration"`
	InitialCapacity  int64         `yaml:"initial_capacity"`
}

type GRPCServer struct {
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

type Raft struct {
	NodeID          string        `yaml:"-"`
	TCPTimeout      time.Duration `yaml:"tcp_timeout"`
	MaxPool         int           `yaml:"max_pool"`
	SnapshotsRetain int           `yaml:"snapshots_retain"`
}

func Read() (*Config, error) {
	file, err := os.Open(*configPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open config file: %w", err)
	}

	var c Config
	if err := yaml.NewDecoder(file).Decode(&c); err != nil {
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
	c.choose(&c.Advertise, advertise)

	if c.Advertise == "" {
		c.Advertise = c.address("localhost", c.InternalPort)
	}

	c.RaftConfig.NodeID = c.Advertise //todo it is bad

	return &c, nil
}

func (c *Config) Logger() *slog.HandlerOptions {
	level := slog.Level(c.LoggerConfig.Level)
	if *verbose {
		level = slog.LevelDebug
	}

	return &slog.HandlerOptions{
		AddSource: false,
		Level:     level,
		//todo костыль чтобы slog был как hclog в идеалье реализовать hlog.Logger через slog
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				return slog.Attr{Key: "@timestamp", Value: a.Value}
			case slog.LevelKey:
				return slog.Attr{Key: "@level", Value: slog.StringValue(strings.ToLower(a.Value.String()))}
			case slog.MessageKey:
				return slog.Attr{Key: "@message", Value: a.Value}
			case sl.MessageComponent:
				return slog.Attr{Key: "@module", Value: a.Value}
			default:
				return a
			}
		},
	}
}

func (c *Config) HashicorpLogger() *hclog.LoggerOptions {
	return &hclog.LoggerOptions{
		Name:                     "hashicorp.Raft.(raft.internal)",
		Level:                    hclog.Level(c.LoggerConfig.Level),
		Output:                   os.Stderr,
		Mutex:                    nil,
		JSONFormat:               true,
		JSONEscapeDisabled:       true,
		IncludeLocation:          false,
		AdditionalLocationOffset: 0,
		TimeFormat:               "",
		TimeFn:                   time.Now,
		DisableTime:              false,
		Color:                    0,
		ColorHeaderOnly:          false,
		ColorHeaderAndFields:     false,
		Exclude:                  nil,
		IndependentLevels:        false,
		SyncParentLevel:          false,
		SubloggerHook:            nil,
	}
}

func (c *Config) Store() core.Config {
	return core.Config{
		CleanInterval:   c.StoreConfig.CleanInterval,
		CleanDuration:   c.StoreConfig.MaxCleanDuration,
		InitialCapacity: c.StoreConfig.InitialCapacity,
	}
}

func (c *Config) ExistingRaftClient() clients.RaftClientConfig {
	return clients.RaftClientConfig{
		Address:  *joinTo,
		Username: c.Username,
		Password: c.Password,
	}
}

func (c *Config) Raft() raft.Config {
	return raft.Config{
		RealAddress:       c.address(c.Host, c.InternalPort),
		AdvertisedAddress: c.Advertise,
		NodeID:            c.RaftConfig.NodeID,
		DataLocation:      c.DataPath,
		SnapshotsRetain:   c.RaftConfig.SnapshotsRetain,
		MaxPool:           c.RaftConfig.MaxPool,
		TCPTimeout:        c.RaftConfig.TCPTimeout,
	}
}

func (c *Config) ClusterNode() raft.ClusterNodeConfig {
	return raft.ClusterNodeConfig{
		ID:               raft.ServerID(c.RaftConfig.NodeID),
		RealAddress:      raft.ServerAddress(c.address(c.Host, c.InternalPort)),
		Advertise:        raft.ServerAddress(c.Advertise),
		BootstrapCluster: *joinTo == "",
	}
}

func (c *Config) GRPCServer() servers.Config {
	return servers.Config{
		Address:           c.address(c.Host, c.PublicPort),
		Username:          c.Username,
		Password:          c.Password,
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
