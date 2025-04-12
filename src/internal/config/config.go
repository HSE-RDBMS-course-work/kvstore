package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/prometheus/common/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	RaftConfig       RaftConfig       `mapstructure:"raft"`
	GRPCServerConfig GRPCServerConfig `mapstructure:"grpc_server"`
}

type GRPCServerConfig struct {
	Host    string        `mapstructure:"host" env:"GRPC_SERVER_HOST"`
	Port    string        `mapstructure:"port" env:"GRPC_SERVER_PORT"`
	Timeout time.Duration `mapstructure:"timeout" env:"GRPC_SERVER_TIMEOUT"`
}

type RaftConfig struct {
	Host             string        `mapstructure:"host" env:"RAFT_HOST"`
	Advertise        string        `mapstructure:"advertise" env:"RAFT_ADVERTISE"`
	Port             string        `mapstructure:"port" env:"RAFT_PORT"`
	LocalID          string        `mapstructure:"local_id" env:"RAFT_LOCAL_ID"`
	LogLocation      string        `mapstructure:"log_location" env:"RAFT_LOG_LOCATION"`
	StableLocation   string        `mapstructure:"stable_location" env:"RAFT_STABLE_LOCATION"`
	SnapshotLocation string        `mapstructure:"snapshot_location" env:"RAFT_SNAPSHOT_LOCATION"`
	LeaderAddr       string        `mapstructure:"leader_address" env:"RAFT_LEADER_ADDRESS"`
	Timeout          time.Duration `mapstructure:"timeout" env:"RAFT_TIMEOUT"`
	MaxPool          int           `mapstructure:"max_pool" env:"RAFT_MAX_POOL"`
}

const defaultConfigPath = "./config.yaml"

func New() (*Config, error) {
	// Parsing flags
	pflag.String("config-path", defaultConfigPath, "YAML Config Path")
	pflag.Parse()

	// Reading yaml
	viper.SetConfigFile(pflag.Lookup("config-path").Value.String())
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Info("config.yaml not found")
	}

	// Setup config
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
