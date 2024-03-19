package logging

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// defaultZerlogRequestLoggerConfig is the default config for logging a request.
var defaultZerlogRequestLoggerConfig = middleware.RequestLoggerConfig{ //nolint:gochecknoglobals,exhaustruct
	Skipper:     middleware.DefaultSkipper,
	LogLatency:  true,
	LogRemoteIP: true,
	LogMethod:   true,
	LogURI:      true,
	LogStatus:   true,
	LogError:    true,
}

// NewZerlogRequestLogger generates the logger function being used by the logging middleware.
func NewZerlogRequestLogger(
	logger *zerolog.Logger,
) func(ctx echo.Context, values middleware.RequestLoggerValues) error {
	return func(_ echo.Context, values middleware.RequestLoggerValues) error {
		event := logger.Info() //nolint:zerologlint
		if values.Error != nil {
			event = logger.Error().Err(values.Error) //nolint:zerologlint
		} else if values.Status >= 400 && values.Status < 600 {
			// warn if status code is in 400 (client error) or 500 (server error) range
			// but did complete the request.
			event = logger.Warn() //nolint:zerologlint
		}

		event.
			Dur("latency", values.Latency).
			Str("ip", values.RemoteIP).
			Str("method", values.Method).
			Str("URI", values.URI).
			Int("status", values.Status).
			Msg("request")

		return nil
	}
}

// NewEchoZerlogLogger builds a logging middleware with default config.
func NewEchoZerlogLogger(logger *zerolog.Logger) echo.MiddlewareFunc {
	return NewEchoZerlogLoggerWithConfig(logger, defaultZerlogRequestLoggerConfig)
}

// NewEchoZerlogLoggerWithConfig builds a logging middleware with custom config.
func NewEchoZerlogLoggerWithConfig(logger *zerolog.Logger, config middleware.RequestLoggerConfig) echo.MiddlewareFunc {
	config.LogValuesFunc = NewZerlogRequestLogger(logger)

	return middleware.RequestLoggerWithConfig(config)
}
