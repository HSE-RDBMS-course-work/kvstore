package config

import (
	"flag"
	"path/filepath"
)

var (
	configPath = flag.String(
		"config",
		filepath.Join("/", "var", "lib", "kvstore", "config.yaml"),
		"Path to configuration file",
	)
	dataPath = flag.String(
		"data",
		filepath.Join("/", "var", "lib", "kvstore"),
		"Path to directory with kvstore data",
	)

	username  = flag.String("username", "", "Username to use for authentication")
	password  = flag.String("password", "", "Password to use for authentication")
	host      = flag.String("host", "0.0.0.0", "Host to use for authentication")
	pPort     = flag.String("public-port", "8090", "Port to use for authentication")
	iPort     = flag.String("internal-port", "3000", "Port to use for authentication")
	advertise = flag.String("advertise", "", "Advertise address to other cluster nodes can interact with this node (default: {host}:{iport}")
	joinTo    = flag.String("join-to", "", "Address of the leader or some node of cluster which is running, provide it to join to this cluster")
)

func init() {
	flag.Parse()
}
