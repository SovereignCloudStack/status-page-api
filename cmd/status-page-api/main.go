package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SovereignCloudStack/status-page-api/internal/app/config"
	"github.com/SovereignCloudStack/status-page-api/internal/app/db"
	"github.com/SovereignCloudStack/status-page-api/internal/app/metrics"
	APIServer "github.com/SovereignCloudStack/status-page-api/internal/app/server"
	APIImplementation "github.com/SovereignCloudStack/status-page-api/pkg/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() { //nolint:funlen,cyclop
	// signal handling
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// setup logging
	logger := log.Output(zerolog.ConsoleWriter{ //nolint:exhaustruct
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}).Level(zerolog.WarnLevel)

	// Reading config
	conf, err := config.New()
	if err != nil {
		logger.Fatal().Err(err).Msg("error loading config")
	}

	err = conf.IsValid()
	if err != nil {
		logger.Fatal().Err(err).Msg("config is invalid")
	}

	// leveled logging
	switch conf.Verbose {
	case 1:
		logger = logger.Level(zerolog.InfoLevel)
	case 2: //nolint:gomnd
		logger = logger.Level(zerolog.DebugLevel)
	case 3: //nolint:gomnd
		logger = logger.Level(zerolog.TraceLevel)
	}

	logger.Trace().Interface("config", conf).Send()

	// named logging
	echoLogger := logger.With().Str("component", "echo").Logger()
	gormLogger := logger.With().Str("component", "gorm").Logger()
	handlerLogger := logger.With().Str("component", "handler").Logger()
	metricsLogger := logger.With().Str("component", "metrics").Logger()

	// DB setup
	dbWrapper, err := db.New(conf.Database.ConnectionString, &gormLogger)
	if err != nil {
		logger.Fatal().Err(err).Msg("error creating database wrapper")
	}

	// Initialize "static" DB contents
	err = dbWrapper.Provision(conf.ProvisioningFile)
	if err != nil {
		logger.Fatal().Err(err).Msg("error provisioning data")
	}

	// set up metric server
	metricsServer := metrics.New(&conf.Metrics, &metricsLogger)

	// register api server
	apiServer := APIServer.New(&conf.Server, &echoLogger, metricsServer.GetMiddlewareConfig())
	apiServer.RegisterAPI(APIImplementation.New(dbWrapper.GetDBCon(), &handlerLogger))

	// start metric server
	go func() {
		err := metricsServer.Start()
		if err != nil {
			logger.Warn().Err(err).Msg("error running metrics server")
		}
	}()

	// handle error of api server
	errChan := make(chan error, 1)

	// start api server
	go func() {
		err := apiServer.Start()
		if err != nil {
			errChan <- err
		}
	}()

	// handle shutdown
	select {
	case err := <-errChan:
		logger.Error().Err(err).Msg("error running server, shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), conf.ShutdownTimeout)

		err = metricsServer.Shutdown(ctx)
		if err != nil {
			logger.Warn().Err(err).Msg("error shutting down metrics server")
		}

		cancel()
	case sig := <-shutdownChan:
		logger.Log().Str("signal", sig.String()).Msg("got shutdown signal")

		ctx, cancel := context.WithTimeout(context.Background(), conf.ShutdownTimeout)

		err := metricsServer.Shutdown(ctx)
		if err != nil {
			logger.Warn().Err(err).Msg("error shutting down metrics server")
		}

		err = apiServer.Shutdown(ctx)
		if err != nil {
			logger.Warn().Err(err).Msg("error shutting down server")
		}

		cancel()
	}
}
