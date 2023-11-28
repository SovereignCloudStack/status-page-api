package config

import "errors"

var (
	// ErrNoDBConnectionString is an error, raised when no database connection string is configured.
	ErrNoDBConnectionString = errors.New("no database connection string")

	// ErrNoProvisioningFile is an error, raised when no provisioning file is configured.
	ErrNoProvisioningFile = errors.New("no provisioning file")
	// ErrNoListenAddress is an error, raised when no listen address is configured.
	ErrNoListenAddress = errors.New("no listen address")
	// ErrNoAllowedOrigins is an error, raised when no allowed origins is configured.
	ErrNoAllowedOrigins = errors.New("no allowed origins")
)
