package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Database holds configuration regarding the database connection.
type Database struct {
	Host     string
	User     string
	Name     string
	Password string
	SSLMode  string
	Port     int
}

// Server holds configuration regarding the server.
type Server struct {
	ListenAddress string
	CorsOrigins   []string
}

// Config holds all application configuration.
type Config struct {
	ProvisioningFile string
	Server           Server
	Database         Database
	Verbose          int
}

const (
	envPrefix = "SCS_STATUS_PAGE"

	verbose = "verbose"

	databaseHost            = "database-host"
	databaseHostDefault     = "127.0.0.1"
	databaseUser            = "database-user"
	databaseUserDefault     = "postgres"
	databaseName            = "database-name"
	databaseNameDefault     = "postgres"
	databasePort            = "database-port"
	databasePortDefault     = 5432
	databasePassword        = "database-password"
	databasePasswordDefault = "debug"
	databaseSSLMode         = "database-ssl-mode"
	databaseSSLModeDefault  = "disable"

	serverListenAddress        = "server-listen-address"
	serverListenAddressDefault = ":3000"
	serverCorsOrigins          = "server-cors-origins"

	provisioningFile        = "provisioning-file"
	provisioningFileDefault = "./provisioning.yaml"
)

var serverCorsOriginsDefault = []string{"127.0.0.1", "localhost"} //nolint:gochecknoglobals

func setDefaults() {
	viper.SetDefault(verbose, 0)

	viper.SetDefault(databaseHost, databaseHostDefault)
	viper.SetDefault(databaseUser, databaseUserDefault)
	viper.SetDefault(databaseName, databaseNameDefault)
	viper.SetDefault(databasePort, databasePortDefault)
	viper.SetDefault(databasePassword, databasePasswordDefault)
	viper.SetDefault(databaseSSLMode, databaseSSLModeDefault)

	viper.SetDefault(serverListenAddress, serverListenAddressDefault)
	viper.SetDefault(serverCorsOrigins, serverCorsOriginsDefault)

	viper.SetDefault(provisioningFile, provisioningFileDefault)
}

func setFlags() {
	pflag.CountP(verbose, "v", "Increase log level")

	pflag.String(databaseHost, databaseHostDefault, "Database host")
	pflag.String(databaseUser, databaseUserDefault, "Database user")
	pflag.String(databaseName, databaseNameDefault, "Database name")
	pflag.Int(databasePort, databasePortDefault, "Database port")
	pflag.String(databasePassword, databasePasswordDefault, "Database password")
	pflag.String(databaseSSLMode, databaseSSLModeDefault, "Database SSL mode")

	pflag.String(serverListenAddress, serverListenAddressDefault, "Server listen address")
	pflag.StringArray(serverCorsOrigins, serverCorsOriginsDefault, "Server CORS origins to accept")

	pflag.String(provisioningFile, provisioningFileDefault, "YAML file with startup provisioning")
}

func buildConfig() *Config {
	return &Config{
		Database: Database{
			Host:     viper.GetString(databaseHost),
			User:     viper.GetString(databaseUser),
			Name:     viper.GetString(databaseName),
			Port:     viper.GetInt(databasePort),
			Password: viper.GetString(databasePassword),
			SSLMode:  viper.GetString(databaseSSLMode),
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

// GetDSN returns the connection string to the database.
func (db *Database) GetDSN() string {
	return fmt.Sprintf(
		"host=%s user=%s dbname=%s port=%d password=%s sslmode=%s",
		db.Host,
		db.User,
		db.Name,
		db.Port,
		db.Password,
		db.SSLMode,
	)
}
