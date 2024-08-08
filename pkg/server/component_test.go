package server_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

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

var _ = Describe("Component", Ordered, func() {
	const (
		componentID        = "7fecf595-6352-4906-a0d8-b3243ee62ec8"
		componentsEndpoint = "/components"
		componentEndpoint  = componentsEndpoint + "/" + componentID
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
		componentRows *sqlmock.Rows
		impactRows    *sqlmock.Rows
		incidentRows  *sqlmock.Rows

		// actual functions under test
		handlers *server.Implementation

		// expected SQL
		expectedComponentsQuery = regexp.QuoteMeta(`SELECT * FROM "components"`)
		expectedComponentQuery  = regexp.QuoteMeta(
			`SELECT *
			FROM "components"
			WHERE id = $1
			ORDER BY "components"."id"
			LIMIT $2`,
		)
		expectedComponentInsert = regexp.QuoteMeta(
			`INSERT INTO "components" ("display_name","labels","id")
			VALUES ($1,$2,$3)`,
		)
		expectedComponentDelete = regexp.QuoteMeta(`DELETE FROM "components" WHERE id = $1`)
		expectedComponentUpdate = regexp.QuoteMeta(`UPDATE "components" SET "display_name"=$1 WHERE "id" = $2`)
		expectedImpactQuery     = `SELECT .+
		FROM "impacts"
		LEFT JOIN "incidents" "Incident" ON "impacts"\."incident_id" = "Incident"\."id"
		WHERE ended_at IS NULL
		AND "impacts"\."component_id" = \$1`
		expectedImpactQueryWithAt = `SELECT .+
		FROM "impacts"
		LEFT JOIN "incidents" "Incident" ON "impacts"\."incident_id" = "Incident"\."id"
		WHERE \(began_at < \$1 AND ended_at > \$2\)
		OR \(began_at < \$3 AND ended_at IS NULL\)
		AND "impacts"\."component_id" = \$4`

		// UUID of the test component
		componentUUID = uuid.MustParse(componentID)

		// filled test component
		component = db.Component{
			Model: db.Model{
				ID: componentUUID,
			},
			DisplayName:        test.Ptr("Storage"),
			ActivelyAffectedBy: &[]db.Impact{},
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
		componentRows = sqlmock.
			NewRows([]string{"id", "display_name", "labels"})

		impactRows = sqlmock.
			NewRows([]string{"incident_id", "component_id", "impact_type_id"})

		incidentRows = sqlmock.
			NewRows([]string{"id", "display_name", "description", "began_at", "ended_at", "phase_generation", "phase_order"})
	})

	AfterEach(func() {
		// check every expectation after each test and close database
		Ω(sqlMock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("GetComponents", func() {
		var (
			ctx    echo.Context
			res    *httptest.ResponseRecorder
			params apiServerDefinition.GetComponentsParams
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				componentsEndpoint,
				nil,
			)

			params = apiServerDefinition.GetComponentsParams{
				At: nil,
			}
		})

		Context("without data", func() {
			It("should return an empty list of components", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedComponentsQuery).WillReturnRows(componentRows)

				expectedResult, _ := json.Marshal(apiServerDefinition.ComponentListResponse{
					Data: []apiServerDefinition.ComponentResponseData{},
				})

				// Act
				err := handlers.GetComponents(ctx, params)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with valid data", func() {
			Context("without at param", func() {
				It("should return a list of components", func() {
					// Arrange
					sqlMock.
						ExpectQuery(expectedComponentsQuery).
						WillReturnRows(
							componentRows.AddRow(component.ID, component.DisplayName, component.Labels),
						)
					sqlMock.
						ExpectQuery(expectedImpactQuery).
						WithArgs(componentID).
						WillReturnRows(impactRows, incidentRows)

					expectedResult, _ := json.Marshal(apiServerDefinition.ComponentListResponse{
						Data: []apiServerDefinition.ComponentResponseData{
							component.ToAPIResponse(),
						},
					})

					// Act
					err := handlers.GetComponents(ctx, params)

					// Assert
					Ω(err).ShouldNot(HaveOccurred())
					Ω(res.Code).Should(Equal(http.StatusOK))
					Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
				})
			})
			Context("with at param", func() {
				It("should return a list of components", func() {
					// Arrange
					now := time.Now()
					params.At = &now
					sqlMock.
						ExpectQuery(expectedComponentsQuery).
						WillReturnRows(
							componentRows.AddRow(component.ID, component.DisplayName, component.Labels),
						)
					sqlMock.
						ExpectQuery(expectedImpactQueryWithAt).
						WithArgs(now, now, now, componentID).
						WillReturnRows(impactRows, incidentRows)

					expectedResult, _ := json.Marshal(apiServerDefinition.ComponentListResponse{
						Data: []apiServerDefinition.ComponentResponseData{
							component.ToAPIResponse(),
						},
					})

					// Act
					err := handlers.GetComponents(ctx, params)

					// Assert
					Ω(err).ShouldNot(HaveOccurred())
					Ω(res.Code).Should(Equal(http.StatusOK))
					Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
				})
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedComponentsQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetComponents(ctx, params)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("CreateComponent", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPost,
				componentsEndpoint,
				apiServerDefinition.Component{
					DisplayName: test.Ptr("Storage"),
				},
			)
		})

		Context("with valid request", func() {
			It("should create a component and return its UUID", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedComponentInsert).WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.CreateComponent(ctx)

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
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(echoLogger, http.MethodPost, componentsEndpoint, nil)

				// Act
				err := handlers.CreateComponent(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedComponentInsert).WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.CreateComponent(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("DeleteComponent", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodDelete,
				componentEndpoint,
				nil,
			)
		})

		Context("with valid UUID and an affected row", func() {
			It("should return 204 no content", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedComponentDelete).WithArgs(componentID).WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteComponent(ctx, componentUUID)

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
				sqlMock.ExpectExec(expectedComponentDelete).WithArgs(componentID).WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteComponent(ctx, componentUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedComponentDelete).WithArgs(componentID).WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.DeleteComponent(ctx, componentUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("GetComponent", func() {
		var (
			ctx    echo.Context
			res    *httptest.ResponseRecorder
			params apiServerDefinition.GetComponentParams
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				componentEndpoint,
				nil,
			)

			params = apiServerDefinition.GetComponentParams{
				At: nil,
			}
		})

		Context("with valid UUID and valid data", func() {
			Context("without at param", func() {
				It("should return a single component", func() {
					// Arrange
					sqlMock.
						ExpectQuery(expectedComponentQuery).
						WithArgs(componentID, 1).
						WillReturnRows(
							componentRows.AddRow(component.ID, component.DisplayName, component.Labels),
						)
					sqlMock.
						ExpectQuery(expectedImpactQuery).
						WithArgs(componentID).
						WillReturnRows(impactRows, incidentRows)

					expectedResult, _ := json.Marshal(apiServerDefinition.ComponentResponse{
						Data: component.ToAPIResponse(),
					})

					// Act
					err := handlers.GetComponent(ctx, componentUUID, params)

					// Assert
					Ω(err).ShouldNot(HaveOccurred())
					Ω(res.Code).Should(Equal(http.StatusOK))
					Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
				})
			})

			Context("with at param", func() {
				It("should return a single component", func() {
					// Arrange
					now := time.Now()
					params.At = &now
					sqlMock.
						ExpectQuery(expectedComponentQuery).
						WithArgs(componentID, 1).
						WillReturnRows(
							componentRows.AddRow(component.ID, component.DisplayName, component.Labels),
						)
					sqlMock.
						ExpectQuery(expectedImpactQueryWithAt).
						WithArgs(now, now, now, componentID).
						WillReturnRows(impactRows, incidentRows)

					expectedResult, _ := json.Marshal(apiServerDefinition.ComponentResponse{
						Data: component.ToAPIResponse(),
					})

					// Act
					err := handlers.GetComponent(ctx, componentUUID, params)

					// Assert
					Ω(err).ShouldNot(HaveOccurred())
					Ω(res.Code).Should(Equal(http.StatusOK))
					Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
				})
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedComponentQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetComponent(ctx, componentUUID, params)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("without data", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedComponentQuery).WillReturnRows(componentRows)

				// Act
				err := handlers.GetComponent(ctx, componentUUID, params)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})

	Describe("UpdateComponent", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPatch,
				componentEndpoint,
				apiServerDefinition.Component{
					DisplayName: test.Ptr("Network"),
				},
			)
		})

		Context("with valid UUID and valid request", func() {
			It("should return 204 no conntent", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedComponentUpdate).
					WithArgs("Network", componentID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateComponent(ctx, componentUUID)

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
					componentEndpoint,
					nil,
				)

				// Act
				err := handlers.UpdateComponent(ctx, componentUUID)

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
					ExpectExec(expectedComponentUpdate).
					WithArgs("Network", componentID).
					WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.UpdateComponent(ctx, componentUUID)

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
					ExpectExec(expectedComponentUpdate).
					WithArgs("Network", componentID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateComponent(ctx, componentUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})
})
