package logger

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

type EchoLogger struct {
	logger *zerolog.Logger
}

func (el *EchoLogger) RequestLogger(_ echo.Context, values middleware.RequestLoggerValues) error {
	logger := el.logger.Info()
	if values.Error != nil {
		logger = el.logger.Error().Err(values.Error)
	}

	logger.
		Dur("latency", values.Latency).
		Str("ip", values.RemoteIP).
		Str("method", values.Method).
		Str("URI", values.URI).
		Int("status", values.Status).
		Msg("request")

	return nil
}

func NewEchoLogger(logger *zerolog.Logger) *EchoLogger {
	return &EchoLogger{
		logger: logger,
	}
}

func NewEchoLoggerConfig(logger *zerolog.Logger) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(
		middleware.RequestLoggerConfig{ //nolint:exhaustruct
			LogLatency:    true,
			LogRemoteIP:   true,
			LogMethod:     true,
			LogURI:        true,
			LogStatus:     true,
			LogError:      true,
			LogValuesFunc: NewEchoLogger(logger).RequestLogger,
		},
	)
}
