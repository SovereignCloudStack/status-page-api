package test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SovereignCloudStack/status-page-api/internal/app/logging"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/gomega" //nolint:revive,stylecheck
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ErrTestError is a test error.
var ErrTestError = errors.New("test error")

// MustSetupLogging creates child loggers for echo, gorm and handlers.
func MustSetupLogging(loglevel zerolog.Level) (*zerolog.Logger, *zerolog.Logger, *zerolog.Logger) {
	log := log.Level(loglevel).Output(zerolog.ConsoleWriter{ //nolint:exhaustruct
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	echoLogger := log.With().Str("component", "echo").Logger()
	gormLogger := log.With().Str("component", "gorm").Logger()
	handlerLogger := log.With().Str("component", "handler").Logger()

	return &echoLogger, &gormLogger, &handlerLogger
}

// MustMockGorm creates SQL mock and connects gorm to the mock.
// Fails matching in tests, when an error occures.
func MustMockGorm(gormLogger *zerolog.Logger) (*sql.DB, sqlmock.Sqlmock, *gorm.DB) { //nolint:ireturn,nolintlint
	// mock sql connection
	sqlDB, sqlMock, err := sqlmock.New()
	Ω(err).ShouldNot(HaveOccurred())

	// connect gorm to mock db
	dbCon, err := gorm.Open(postgres.New(postgres.Config{ //nolint:exhaustruct
		Conn: sqlDB,
	}), &gorm.Config{ //nolint:exhaustruct
		Logger: logging.NewGormLogger(gormLogger),
	})
	Ω(err).ShouldNot(HaveOccurred())

	return sqlDB, sqlMock, dbCon
}

// MustCreateRequestAndResponseWriter creates a http request and a response writer in form of a response recorder.
// Fails matching in tests, when an error occures.
func MustCreateRequestAndResponseWriter(
	method string,
	target string,
	body interface{},
) (
	*http.Request,
	*httptest.ResponseRecorder,
) {
	// create request and response
	var req *http.Request

	if body == nil {
		req = httptest.NewRequest(method, target, nil)
	} else {
		jsonBody, err := json.Marshal(body)
		Ω(err).ShouldNot(HaveOccurred())

		req = httptest.NewRequest(method, target, bytes.NewReader(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}

	rec := httptest.NewRecorder()

	return req, rec
}

// MustCreateEchoContext creates an echo context with the given logger and request and response writer.
func MustCreateEchoContext( //nolint:ireturn,nolintlint
	echoLogger *zerolog.Logger,
	request *http.Request,
	responseWriter http.ResponseWriter,
) echo.Context {
	echoServer := echo.New()
	echoServer.Use(logging.NewEchoZerlogLogger(echoLogger))
	ctx := echoServer.NewContext(request, responseWriter)

	return ctx
}

// MustCreateEchoContextAndResponseWriter is a convenience function combining [MustCreateEchoContext]
// and [MustCreateRequestAndResponseWriter] when access to the request is not needed.
func MustCreateEchoContextAndResponseWriter( //nolint:ireturn,nolintlint
	echoLogger *zerolog.Logger,
	method string,
	target string,
	body interface{},
) (
	echo.Context,
	*httptest.ResponseRecorder,
) {
	req, res := MustCreateRequestAndResponseWriter(method, target, body)
	ctx := MustCreateEchoContext(echoLogger, req, res)

	return ctx, res
}

// Ptr returns a pointer to the value.
// This is used for testing when defining values in code
// and they need to be a pointer for structs.
func Ptr[T any](value T) *T {
	return &value
}
