package logger

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	logger *zerolog.Logger
}

func NewGormLogger(logger *zerolog.Logger) *GormLogger {
	return &GormLogger{
		logger: logger,
	}
}

func (gl *GormLogger) LogMode(logger.LogLevel) logger.Interface { //nolint:ireturn
	return gl
}

func (gl *GormLogger) Info(_ context.Context, message string, data ...interface{}) {
	gl.logger.Info().Interface("data", data).Msg(message)
}

func (gl *GormLogger) Warn(_ context.Context, message string, data ...interface{}) {
	gl.logger.Warn().Interface("data", data).Msg(message)
}

func (gl *GormLogger) Error(_ context.Context, message string, data ...interface{}) {
	gl.logger.Error().Interface("data", data).Msg(message)
}

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
