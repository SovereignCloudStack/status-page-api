package server_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-api/pkg/server"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var _ = Describe("Impact", Ordered, func() {
	const (
		impactTypeID        = "c3fc130d-e6c4-4f94-86ba-e51fbdfc5d0c"
		impactTypesEndpoint = "/impactTypes"
		impactTypeEndpoint  = impactTypesEndpoint + "/" + impactTypeID
	)

	var (
		// sub loggers
		echoLogger    *zerolog.Logger
		gormLogger    *zerolog.Logger
		handlerLogger *zerolog.Logger

		// sql mocking
		sqlDB   *sql.DB
		sqlMock sqlmock.Sqlmock

		// mocked sql rows
		impactTypeRows *sqlmock.Rows

		// actual functions under test
		handlers *server.Implementation

		// expected SQL
		expectedImpactTypesQuery         = regexp.QuoteMeta(`SELECT * FROM "impact_types"`)
		expectedImpactTypeQuery          = regexp.QuoteMeta(`SELECT * FROM "impact_types" WHERE id = $1 ORDER BY "impact_types"."id" LIMIT $2`)                  //nolint:lll
		expectedImpactTypeQueryWithTable = regexp.QuoteMeta(`SELECT * FROM "impact_types" WHERE "impact_types"."id" = $1 ORDER BY "impact_types"."id" LIMIT $2`) //nolint:lll
		expectedImpactTypeInsert         = regexp.QuoteMeta(`INSERT INTO "impact_types" ("display_name","description","id") VALUES ($1,$2,$3)`)                  //nolint:lll
		expectedImpactTypeDelete         = regexp.QuoteMeta(`DELETE FROM "impact_types" WHERE id = $1`)
		expectedImpactTypeUpdate         = regexp.QuoteMeta(`UPDATE "impact_types" SET "display_name"=$1 WHERE "id" = $2`)

		// UUID of the test impact type
		impactTypeUUID = uuid.MustParse(impactTypeID)

		// filled test impact type
		impactType = db.ImpactType{
			Model: db.Model{
				ID: impactTypeUUID,
			},
			DisplayName: test.Ptr("Performance degration"),
			Description: test.Ptr("Performance has been decreased in some parts of the system."),
		}
	)

	BeforeAll(func() {
		// setup loggers once
		echoLogger, gormLogger, handlerLogger = test.MustSetupLogging(zerolog.TraceLevel)
	})

	BeforeEach(func() {
		// setup database and mock before each test
		var gormDB *gorm.DB

		sqlDB, sqlMock, gormDB = test.MustMockGorm(gormLogger)
		handlers = server.New(gormDB, handlerLogger)

		// create mock rows before each test
		impactTypeRows = sqlmock.
			NewRows([]string{"id", "display_name", "description"})
	})

	AfterEach(func() {
		// check every expectation after each test and close database
		Ω(sqlMock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("GetImpactTypes", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				impactTypesEndpoint,
				nil,
			)
		})

		Context("without data", func() {
			It("should return an empty list of impact types", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedImpactTypesQuery).WillReturnRows(impactTypeRows)

				expectedResult, _ := json.Marshal(apiServerDefinition.ImpactTypeListResponse{
					Data: []apiServerDefinition.ImpactTypeResponseData{},
				})

				// Act
				err := handlers.GetImpactTypes(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with valid data", func() {
			It("should return a list of impactTypes", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedImpactTypesQuery).
					WillReturnRows(
						impactTypeRows.AddRow(impactType.ID, impactType.DisplayName, impactType.Description),
					)

				expectedResult, _ := json.Marshal(apiServerDefinition.ImpactTypeListResponse{
					Data: []apiServerDefinition.ImpactTypeResponseData{
						impactType.ToAPIResponse(),
					},
				})

				// Act
				err := handlers.GetImpactTypes(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedImpactTypesQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetImpactTypes(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("CreateImpactType", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPost,
				impactTypesEndpoint,
				apiServerDefinition.ImpactType{
					DisplayName: test.Ptr("Performance degration"),
				},
			)
		})

		Context("with valid request", func() {
			It("should create an impact type and return its UUID", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedImpactTypeInsert).WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.CreateImpactType(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())

				// parse answer to get uuid
				var response apiServerDefinition.IdResponse
				err = json.Unmarshal(res.Body.Bytes(), &response)

				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusCreated))
			})
		})

		Context("with empty request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(echoLogger, http.MethodPost, impactTypesEndpoint, nil)

				// Act
				err := handlers.CreateImpactType(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedImpactTypeInsert).WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.CreateImpactType(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("DeleteImpactType", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodDelete,
				impactTypeEndpoint,
				nil,
			)
		})

		Context("with valid UUID and an affected row", func() {
			It("should return 204 no content", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedImpactTypeDelete).WithArgs(impactTypeID).WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteImpactType(ctx, impactTypeUUID)

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
				sqlMock.ExpectExec(expectedImpactTypeDelete).WithArgs(impactTypeID).WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteImpactType(ctx, impactTypeUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedImpactTypeDelete).WithArgs(impactTypeID).WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.DeleteImpactType(ctx, impactTypeUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("GetImpactType", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				impactTypeEndpoint,
				nil,
			)
		})

		Context("with valid UUID and valid data", func() {
			It("should return a single impactType", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedImpactTypeQuery).
					WithArgs(impactTypeID, 1).
					WillReturnRows(
						impactTypeRows.AddRow(impactType.ID, impactType.DisplayName, impactType.Description),
					)

				expectedResult, _ := json.Marshal(apiServerDefinition.ImpactTypeResponse{
					Data: impactType.ToAPIResponse(),
				})

				// Act
				err := handlers.GetImpactType(ctx, impactTypeUUID)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedImpactTypeQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetImpactType(ctx, impactTypeUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("without data", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedImpactTypeQuery).WillReturnRows(impactTypeRows)

				// Act
				err := handlers.GetImpactType(ctx, impactTypeUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})

	Describe("UpdateImpactType", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPatch,
				impactTypeEndpoint,
				apiServerDefinition.ImpactType{
					DisplayName: test.Ptr("Connectivity problems"),
				},
			)
		})

		Context("with valid UUID and valid request", func() {
			It("should return 204 no content", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectQuery(expectedImpactTypeQueryWithTable).
					WithArgs(impactTypeID, 1).
					WillReturnRows(
						impactTypeRows.AddRow(
							impactType.ID,
							impactType.DisplayName,
							impactType.Description,
						),
					)
				sqlMock.
					ExpectExec(expectedImpactTypeUpdate).
					WithArgs("Connectivity problems", impactTypeID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateImpactType(ctx, impactTypeUUID)

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
					impactTypeEndpoint,
					nil,
				)

				// Act
				err := handlers.UpdateImpactType(ctx, impactTypeUUID)

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
					ExpectQuery(expectedImpactTypeQueryWithTable).
					WithArgs(impactTypeID, 1).
					WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.UpdateImpactType(ctx, impactTypeUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("without existing impact type", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectQuery(expectedImpactTypeQueryWithTable).
					WithArgs(impactTypeID, 1).
					WillReturnRows(impactTypeRows)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.UpdateImpactType(ctx, impactTypeUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})
})
