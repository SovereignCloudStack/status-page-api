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

func (db Database) isValid() error {
	if db.ConnectionString == "" {
		return ErrNoDBConnectionString
	}

	return nil
}

// Config holds all application configuration.
type Config struct {
	ProvisioningFile string
	ListenAddress    string
	Database         Database
	AllowedOrigins   []string
	Verbose          int
	SwaggerEnabled   bool
}

// IsValid validates the config by checking own values and calling isValid on sub config objects.
func (c Config) IsValid() error {
	if c.ProvisioningFile == "" {
		return ErrNoProvisioningFile
	}

	if c.ListenAddress == "" {
		return ErrNoListenAddress
	}

	if len(c.AllowedOrigins) == 0 {
		return ErrNoAllowedOrigins
	}

	err := c.Database.isValid()
	if err != nil {
		return fmt.Errorf("error validating database config: %w", err)
	}

	return nil
}

const (
	envPrefix = "SCS_STATUS_PAGE"

	verbose = "verbose"

	swaggerUIEnabled        = "swagger.ui.enabled"
	swaggerUIEnabledDefault = false

	databaseConnectionString        = "database.connection-string"
	databaseConnectionStringDefault = ""

	listenAddress        = "listen-address"
	listenAddressDefault = ":3000"
	allowedOrigins       = "allowed-origins"

	provisioningFile        = "provisioning-file"
	provisioningFileDefault = "./provisioning.yaml"
)

var allowedOriginsDefault = []string{"127.0.0.1", "localhost"} //nolint:gochecknoglobals

func setDefaults() {
	viper.SetDefault(verbose, 0)

	viper.SetDefault(swaggerUIEnabled, swaggerUIEnabledDefault)

	viper.SetDefault(databaseConnectionString, databaseConnectionStringDefault)

	viper.SetDefault(listenAddress, listenAddressDefault)
	viper.SetDefault(allowedOrigins, allowedOriginsDefault)

	viper.SetDefault(provisioningFile, provisioningFileDefault)
}

func setFlags() {
	pflag.CountP(verbose, "v", "Increase log level")

	pflag.Bool(swaggerUIEnabled, swaggerUIEnabledDefault, "Enable swagger UI for development.")

	pflag.String(databaseConnectionString, databaseConnectionStringDefault, "Database connection string")

	pflag.String(listenAddress, listenAddressDefault, "Server listen address")
	pflag.StringArray(allowedOrigins, allowedOriginsDefault, "Server CORS origins to accept")

	pflag.String(provisioningFile, provisioningFileDefault, "YAML file with startup provisioning")
}

func pflagNormalizer(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ReplaceAll(name, ".", "-"))
}

func buildConfig() *Config {
	return &Config{
		AllowedOrigins: viper.GetStringSlice(allowedOrigins),
		Database: Database{
			ConnectionString: strings.TrimSpace(viper.GetString(databaseConnectionString)),
		},
		ListenAddress:    strings.TrimSpace(viper.GetString(listenAddress)),
		ProvisioningFile: strings.TrimSpace(viper.GetString(provisioningFile)),
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
