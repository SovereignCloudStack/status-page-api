package config

import "errors"

var (
	// ErrNoDBConnectionString is an error, raised when no database connection string is configured.
	ErrNoDBConnectionString = errors.New("no database connection string")

	// ErrNoProvisioningFile is an error, raised when no provisioning file is configured.
	ErrNoProvisioningFile = errors.New("no provisioning file")

	// ErrNoServerAddress is an error, raised when no server address is configured.
	ErrNoServerAddress = errors.New("no server address")
	// ErrNoAllowedOrigins is an error, raised when no allowed origins is configured.
	ErrNoAllowedOrigins = errors.New("no allowed origins")

	// ErrNoMetricNamespace is an error, raised when no metric namespace is configured.
	ErrNoMetricNamespace = errors.New("no metrics namespace")
	// ErrNoMetricSubsystem is an error, raised when no metric subsystem is configured.
	ErrNoMetricSubsystem = errors.New("no metrics subsystem")
)
