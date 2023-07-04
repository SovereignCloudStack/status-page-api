package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
)

// GormLogger handles logging of gorm logs.
type GormLogger struct {
	logger *zerolog.Logger
}

// NewGormLogger creates a new [GormLogger].
func NewGormLogger(logger *zerolog.Logger) *GormLogger {
	return &GormLogger{
		logger: logger,
	}
}

// LogMode should change the loggers log level.
// zerlog does not need this change on the fly.
// The function just exists to satisfy the [logger.Interface].
func (gl *GormLogger) LogMode(logger.LogLevel) logger.Interface { //nolint:ireturn
	return gl
}

// Info logs on info level.
func (gl *GormLogger) Info(_ context.Context, message string, data ...interface{}) {
	gl.logger.Info().Msg(fmt.Sprintf(message, data...))
}

// Warn logs on warn level.
func (gl *GormLogger) Warn(_ context.Context, message string, data ...interface{}) {
	gl.logger.Warn().Msg(fmt.Sprintf(message, data...))
}

// Error logs on error level.
func (gl *GormLogger) Error(_ context.Context, message string, data ...interface{}) {
	gl.logger.Error().Msg(fmt.Sprintf(message, data...))
}

// Trace logs on trace level.
// This logging function has a lot more information.
// It is used to log SQL statements generated by gorm, for example.
func (gl *GormLogger) Trace(
	_ context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		if errors.Is(err, logger.ErrRecordNotFound) {
			gl.logger.Warn().Err(err).Dur("time", elapsed).Int64("rows", rows).Str("sql", sql).Send()
		} else {
			gl.logger.Error().Err(err).Dur("time", elapsed).Int64("rows", rows).Str("sql", sql).Send()
		}

		return
	}

	gl.logger.Trace().Err(err).Dur("time", elapsed).Int64("rows", rows).Str("sql", sql).Send()
}
