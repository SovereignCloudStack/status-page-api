package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Database holds configuration regarding the database connection.
type Database struct {
	ConnectionString string
}

// Server holds configuration regarding the server.
type Server struct {
	ListenAddress string
	CorsOrigins   []string
}

// Config holds all application configuration.
type Config struct {
	ProvisioningFile string
	Database         Database
	Server           Server
	Verbose          int
}

const (
	envPrefix = "SCS_STATUS_PAGE"

	verbose = "verbose"

	databaseConnectionString        = "database-connection-string"
	databaseConnectionStringDefault = "host=127.0.0.1 user=postgres dbname=postgres port=5432 password=debug sslmode=disable" //nolint:lll

	serverListenAddress        = "server-listen-address"
	serverListenAddressDefault = ":3000"
	serverCorsOrigins          = "server-cors-origins"

	provisioningFile        = "provisioning-file"
	provisioningFileDefault = "./provisioning.yaml"
)

var serverCorsOriginsDefault = []string{"127.0.0.1", "localhost"} //nolint:gochecknoglobals

func setDefaults() {
	viper.SetDefault(verbose, 0)

	viper.SetDefault(databaseConnectionString, databaseConnectionStringDefault)

	viper.SetDefault(serverListenAddress, serverListenAddressDefault)
	viper.SetDefault(serverCorsOrigins, serverCorsOriginsDefault)

	viper.SetDefault(provisioningFile, provisioningFileDefault)
}

func setFlags() {
	pflag.CountP(verbose, "v", "Increase log level")

	pflag.String(databaseConnectionString, databaseConnectionStringDefault, "Database connection string")

	pflag.String(serverListenAddress, serverListenAddressDefault, "Server listen address")
	pflag.StringArray(serverCorsOrigins, serverCorsOriginsDefault, "Server CORS origins to accept")

	pflag.String(provisioningFile, provisioningFileDefault, "YAML file with startup provisioning")
}

func buildConfig() *Config {
	return &Config{
		Database: Database{
			ConnectionString: viper.GetString(databaseConnectionString),
		},
		Server: Server{
			ListenAddress: viper.GetString(serverListenAddress),
			CorsOrigins:   viper.GetStringSlice(serverCorsOrigins),
		},
		ProvisioningFile: viper.GetString(provisioningFile),
		Verbose:          viper.GetInt(verbose),
	}
}

// New creates a new configuration.
func New() (*Config, error) {
	// defaults
	setDefaults()

	// envs
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// flags
	setFlags()
	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("error binding flags: %w", err)
	}

	// new config
	return buildConfig(), nil
}
