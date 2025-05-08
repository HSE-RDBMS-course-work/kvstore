package config

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/spf13/viper"
	"kvstore/internal/core"
	"kvstore/internal/grpc/servers"
	"kvstore/internal/raft"
	"kvstore/internal/sl"
	"log/slog"
	"os"
	"strings"
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

type Config struct {
	Host             string     `mapstructure:"host"`
	PublicPort       string     `mapstructure:"public_port"`
	InternalPort     string     `mapstructure:"internal_port"`
	Username         string     `mapstructure:"username"`
	Password         string     `mapstructure:"password"`
	DataPath         string     `mapstructure:"data_path"`
	LoggerConfig     Logger     `mapstructure:"logger"`
	StoreConfig      Store      `mapstructure:"storage"`
	GRPCServerConfig GRPCServer `mapstructure:"grpc_server"`
	RaftConfig       Raft       `mapstructure:"raft"`
}

type Logger struct {
	Level int `mapstructure:"level"`
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

	if c.RaftConfig.Advertise == "" {
		c.RaftConfig.Advertise = c.address("localhost", c.InternalPort)
	}

	c.RaftConfig.NodeID = c.RaftConfig.Advertise

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
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr { //todo костыль чтобы slog был как hlog в идеалье реализовать hlog через slog
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
