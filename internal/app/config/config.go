package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Database holds configuration regarding the database connection.
type Database struct {
	ConnectionString string `json:"-"` // do not leak database password when logging.
}

func (db Database) isValid() error {
	if db.ConnectionString == "" {
		return ErrNoDBConnectionString
	}

	return nil
}

// Server holds configuration regarding the api server.
type Server struct {
	Address        string
	AllowedOrigins []string
	SwaggerEnabled bool
}

func (s Server) isValid() error {
	if s.Address == "" {
		return ErrNoServerAddress
	}

	if len(s.AllowedOrigins) == 0 {
		return ErrNoAllowedOrigins
	}

	return nil
}

// Metrics holds configuration regarding the metrics server.
type Metrics struct {
	Namespace string
	Subsystem string
	Address   string
}

func (m Metrics) isValid() error {
	if m.Namespace == "" {
		return ErrNoMetricNamespace
	}

	if m.Subsystem == "" {
		return ErrNoMetricSubsystem
	}

	return nil
}

// Config holds all application configuration.
type Config struct {
	ProvisioningFile string
	Metrics          Metrics
	Database         Database
	Server           Server
	Verbose          int
	ShutdownTimeout  time.Duration
}

// IsValid validates the config by checking own values and calling isValid on sub config objects.
func (c Config) IsValid() error {
	if c.ProvisioningFile == "" {
		return ErrNoProvisioningFile
	}

	err := c.Metrics.isValid()
	if err != nil {
		return fmt.Errorf("error validating metrics config: %w", err)
	}

	err = c.Database.isValid()
	if err != nil {
		return fmt.Errorf("error validating database config: %w", err)
	}

	err = c.Server.isValid()
	if err != nil {
		return fmt.Errorf("error validating server config: %w", err)
	}

	return nil
}

const (
	envPrefix = "STATUS_PAGE"

	verbose = "verbose"

	databaseConnectionString        = "database.connection-string"
	databaseConnectionStringDefault = ""

	metricsNamespace        = "metrics.namespace"
	metricsNamespaceDefault = "status_page"
	metricsSubsystem        = "metrics.subsystem"
	metricsSubsystemDefault = "api"
	metricsAddress          = "metrics.address"
	metricsAddressDefault   = ""

	serverAddress                 = "server.address"
	serverAddressDefault          = ":3000"
	serverAllowedOrigins          = "server.allowed-origins"
	serverSwaggerUIEnabled        = "server.swagger.ui.enabled"
	serverSwaggerUIEnabledDefault = false

	provisioningFile        = "provisioning-file"
	provisioningFileDefault = "./provisioning.yaml"

	shutdownTimeout        = "shutdown-timeout"
	shutdownTimeoutDefault = 50 * time.Millisecond
)

var serverAllowedOriginsDefault = []string{"127.0.0.1", "localhost"} //nolint:gochecknoglobals

func setDefaults() {
	viper.SetDefault(verbose, 0)

	viper.SetDefault(serverSwaggerUIEnabled, serverSwaggerUIEnabledDefault)

	viper.SetDefault(databaseConnectionString, databaseConnectionStringDefault)

	viper.SetDefault(metricsNamespace, metricsNamespaceDefault)
	viper.SetDefault(metricsSubsystem, metricsSubsystemDefault)
	viper.SetDefault(metricsAddress, metricsAddressDefault)

	viper.SetDefault(serverAddress, serverAddressDefault)
	viper.SetDefault(serverAllowedOrigins, serverAllowedOriginsDefault)

	viper.SetDefault(provisioningFile, provisioningFileDefault)

	viper.SetDefault(shutdownTimeout, shutdownTimeoutDefault)
}

func setFlags() {
	pflag.CountP(verbose, "v", "Increase log level")

	pflag.String(databaseConnectionString, databaseConnectionStringDefault, "Database connection string.")

	pflag.String(metricsNamespace, metricsNamespace, "Metrics namespace.")
	pflag.String(metricsSubsystem, metricsSubsystem, "Metrics sub system name.")
	pflag.String(metricsAddress, metricsAddressDefault, "Metrics server listen address.")

	pflag.String(serverAddress, serverAddressDefault, "Server listen address.")
	pflag.StringArray(serverAllowedOrigins, serverAllowedOriginsDefault, "Server CORS origins to accept.")
	pflag.Bool(serverSwaggerUIEnabled, serverSwaggerUIEnabledDefault, "Enable swagger UI for development.")

	pflag.String(provisioningFile, provisioningFileDefault, "YAML file with startup provisioning.")

	pflag.Duration(shutdownTimeout, shutdownTimeoutDefault, "Duration to wait for the server to gracefully shutdown.")
}

func pflagNormalizer(_ *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(strings.ReplaceAll(name, ".", "-"))
}

func buildConfig() *Config {
	return &Config{
		Database: Database{
			ConnectionString: strings.TrimSpace(viper.GetString(databaseConnectionString)),
		},
		Server: Server{
			Address:        strings.TrimSpace(viper.GetString(serverAddress)),
			AllowedOrigins: viper.GetStringSlice(serverAllowedOrigins),
			SwaggerEnabled: viper.GetBool(serverSwaggerUIEnabled),
		},
		Metrics: Metrics{
			Namespace: strings.TrimSpace(viper.GetString(metricsNamespace)),
			Subsystem: strings.TrimSpace(viper.GetString(metricsSubsystem)),
			Address:   strings.TrimSpace(viper.GetString(metricsAddress)),
		},
		ProvisioningFile: strings.TrimSpace(viper.GetString(provisioningFile)),
		Verbose:          viper.GetInt(verbose),
		ShutdownTimeout:  viper.GetDuration(shutdownTimeout),
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
