package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Database holds configuration regarding the database connection.
type Database struct {
	ConnectionString string `json:"-"` // do not leak databse password when logging.
}

// Config holds all application configuration.
type Config struct {
	ProvisioningFile string
	ListenAddress    string
	Database         Database
	CorsOrigins      []string
	Verbose          int
	SwaggerEnabled   bool
}

const (
	envPrefix = "SCS_STATUS_PAGE"

	verbose = "verbose"

	swaggerUIEnabled        = "swagger.ui.enabled"
	swaggerUIEnabledDefault = false

	databaseConnectionString        = "database.connection-string"
	databaseConnectionStringDefault = "host=127.0.0.1 user=postgres dbname=postgres port=5432 password=debug sslmode=disable" //nolint:lll

	listenAddress        = "listen-address"
	listenAddressDefault = ":3000"
	corsOrigins          = "cors-origins"

	provisioningFile        = "provisioning-file"
	provisioningFileDefault = "./provisioning.yaml"
)

var corsOriginsDefault = []string{"127.0.0.1", "localhost"} //nolint:gochecknoglobals

func setDefaults() {
	viper.SetDefault(verbose, 0)

	viper.SetDefault(swaggerUIEnabled, swaggerUIEnabledDefault)

	viper.SetDefault(databaseConnectionString, databaseConnectionStringDefault)

	viper.SetDefault(listenAddress, listenAddressDefault)
	viper.SetDefault(corsOrigins, corsOriginsDefault)

	viper.SetDefault(provisioningFile, provisioningFileDefault)
}

func setFlags() {
	pflag.CountP(verbose, "v", "Increase log level")

	pflag.Bool(swaggerUIEnabled, swaggerUIEnabledDefault, "Enable swagger UI for development.")

	pflag.String(databaseConnectionString, databaseConnectionStringDefault, "Database connection string")

	pflag.String(listenAddress, listenAddressDefault, "Server listen address")
	pflag.StringArray(corsOrigins, corsOriginsDefault, "Server CORS origins to accept")

	pflag.String(provisioningFile, provisioningFileDefault, "YAML file with startup provisioning")
}

func pflagNormalizer(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ReplaceAll(name, ".", "-"))
}

func buildConfig() *Config {
	return &Config{
		CorsOrigins: viper.GetStringSlice(corsOrigins),
		Database: Database{
			ConnectionString: viper.GetString(databaseConnectionString),
		},
		ListenAddress:    viper.GetString(listenAddress),
		ProvisioningFile: viper.GetString(provisioningFile),
		SwaggerEnabled:   viper.GetBool(swaggerUIEnabled),
		Verbose:          viper.GetInt(verbose),
	}
}

// New creates a new configuration.
func New() (*Config, error) {
	// defaults
	setDefaults()

	// envs
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	// flags
	pflag.CommandLine.SetNormalizeFunc(pflagNormalizer)
	setFlags()
	pflag.Parse()

	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("error binding flags: %w", err)
	}

	// new config
	return buildConfig(), nil
}
