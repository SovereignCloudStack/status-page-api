package server_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-api/pkg/server"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var _ = Describe("Severity", func() {
	var (
		// sub loggers
		echoLogger, gormLogger, handlerLogger = test.MustSetupLogging(zerolog.TraceLevel)

		// sql mocking
		sqlDB   *sql.DB
		sqlMock sqlmock.Sqlmock

		// mocked sql rows
		severityRows *sqlmock.Rows

		// actual functions under test
		handlers *server.Implementation

		// expected SQL
		expectedSeveritiesQuery = regexp.
					QuoteMeta(`SELECT * FROM "severities"`)
		expectedSeverityQuery = regexp.
					QuoteMeta(`SELECT * FROM "severities" WHERE display_name = $1 ORDER BY "severities"."display_name" LIMIT $2`)
		expectedSeverityInsert = regexp.
					QuoteMeta(`INSERT INTO "severities" ("display_name","value") VALUES ($1,$2)`)
		expectedSeverityDelete = regexp.
					QuoteMeta(`DELETE FROM "severities" WHERE display_name = $1`)
		expectedSeverityUpdate = regexp.
					QuoteMeta(`UPDATE "severities" SET "display_name"=$1 WHERE display_name = $2`)

		// filled test severity
		severity = db.Severity{
			DisplayName: test.Ptr("broken"),
			Value:       test.Ptr(50),
		}
	)

	BeforeEach(func() {
		// setup database and mock before each test
		var gormDB *gorm.DB

		sqlDB, sqlMock, gormDB = test.MustMockGorm(gormLogger)
		handlers = server.New(gormDB, handlerLogger)

		// create mock rows before each test
		severityRows = sqlmock.
			NewRows([]string{"display_name", "value"})
	})

	AfterEach(func() {
		// check every expectation after each test and close database
		Ω(sqlMock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("GetSeverities", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				"/severities",
				nil,
			)
		})

		Context("without data", func() {
			It("should return an empty list of severities", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedSeveritiesQuery).
					WillReturnRows(severityRows)

				expectedResult, _ := json.Marshal(api.SeverityListResponse{
					Data: []api.Severity{},
				})
				// Act
				err := handlers.GetSeverities(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with valid data", func() {
			It("should return a list of severities", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedSeveritiesQuery).
					WillReturnRows(
						severityRows.AddRow(severity.DisplayName, severity.Value),
					)

				expectedResult, _ := json.Marshal(api.SeverityListResponse{
					Data: []api.Severity{
						severity.ToAPIResponse(),
					},
				})

				// Act
				err := handlers.GetSeverities(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedSeveritiesQuery).
					WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetSeverities(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("CreateSeverity", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPost,
				"/severities",
				api.Severity{
					DisplayName: test.Ptr("broken"),
					Value:       test.Ptr(50),
				},
			)
		})

		Context("with valid request", func() {
			It("should create a severity", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedSeverityInsert).WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.CreateSeverity(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusNoContent))
			})
		})

		Context("with empty request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(echoLogger, http.MethodPost, "/severities", nil)

				// Act
				err := handlers.CreateSeverity(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedSeverityInsert).WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.CreateSeverity(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("DeleteSeverity", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodDelete,
				fmt.Sprintf("/severities/%s", *severity.DisplayName),
				nil,
			)
		})

		Context("with affected row", func() {
			It("should return 204 no content", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedSeverityDelete).
					WithArgs(severity.DisplayName).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusNoContent))
				Ω(res.Body.String()).Should(Equal(""))
			})
		})

		Context("without affected row", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedSeverityDelete).
					WithArgs(severity.DisplayName).
					WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedSeverityDelete).
					WithArgs(severity.DisplayName).
					WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.DeleteSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("GetSeverity", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				fmt.Sprintf("/severities/%s", *severity.DisplayName),
				nil,
			)
		})

		Context("with valid data", func() {
			It("should return a single severity", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedSeverityQuery).
					WithArgs(severity.DisplayName, 1).
					WillReturnRows(
						severityRows.AddRow(severity.DisplayName, severity.Value),
					)

				expectedResult, _ := json.Marshal(api.SeverityResponse{
					Data: severity.ToAPIResponse(),
				})

				// Act
				err := handlers.GetSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedSeverityQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("without data", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedSeverityQuery).WillReturnRows(severityRows)

				// Act
				err := handlers.GetSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})

	Describe("UpdateSeverity", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPatch,
				fmt.Sprintf("/severities/%s", *severity.DisplayName),
				api.Severity{
					DisplayName: test.Ptr("impacted"),
				},
			)
		})

		Context("with valid request", func() {
			It("should return 204 no conntent", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedSeverityUpdate).
					WithArgs("impacted", severity.DisplayName).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusNoContent))
				Ω(res.Body.String()).Should(Equal(""))
			})
		})

		Context("with empty request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodPatch,
					fmt.Sprintf("/severities/%s", *severity.DisplayName),
					nil,
				)

				// Act
				err := handlers.UpdateSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedSeverityUpdate).
					WithArgs("impacted", severity.DisplayName).
					WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.UpdateSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("without affected rows", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedSeverityUpdate).
					WithArgs("impacted", severity.DisplayName).
					WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateSeverity(ctx, *severity.DisplayName)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})
})
