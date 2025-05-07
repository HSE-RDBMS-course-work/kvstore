package config

import (
	"flag"
)

var (
	advertise = flag.String("advertise", "",
		"Advertise address to other cluster nodes can interact with this node "+
			"(You must not provide it if you run it with no custom DNS like docker DNS. "+
			"And it must be either localhost or domain name)",
	)
	configPath = flag.String("config", "", "Path to configuration file")
	dataPath   = flag.String("data", "", "Path to directory with kvstore data")
	username   = flag.String("username", "", "Username to use for authentication")
	password   = flag.String("password", "", "Password to use for authentication")
	host       = flag.String("host", "0.0.0.0", "Host to use for authentication")
	pPort      = flag.String("public-port", "3000", "Port to use for authentication")
	iPort      = flag.String("internal-port", "8090", "Port to use for authentication")
	joinTo     = flag.String("join-to", "", "Address of the leader or some node of cluster which is running, provide it to join to this cluster")
)

func init() {
	flag.Parse()
}
