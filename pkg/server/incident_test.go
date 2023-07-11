package server_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-api/pkg/server"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

const incidentID = "91fd8fa3-4288-4940-bcfb-9e89d82f3522"

var (
	// Sub loggers.
	echoLogger, gormLogger, handlerLogger = test.MustSetupLogging(zerolog.TraceLevel)

	// SQL mocking.
	sqlDB   *sql.DB
	sqlMock sqlmock.Sqlmock

	// Actual functions under test.
	handlers *server.Implementation

	// Time for testing.
	now = time.Now()

	// UUID of the test incident.
	incidentUUID = uuid.MustParse(incidentID)
)

var _ = Describe("Incident", func() {
	var (
		// mocked sql rows
		incidentRows       *sqlmock.Rows
		impactRows         *sqlmock.Rows
		componentRows      *sqlmock.Rows
		phaseRows          *sqlmock.Rows
		incidentUpdateRows *sqlmock.Rows

		// expected SQL
		expectedIncidentsQuery = regexp.
					QuoteMeta(`SELECT * FROM "incidents" WHERE NOT (began_at < $1 AND ended_at < $2) AND NOT (began_at > $3 AND ended_at > $4) OR (ended_at IS NULL AND began_at <= $5)`) //nolint:lll
		expectedIncidentQuery = regexp.
					QuoteMeta(`SELECT * FROM "incidents" WHERE id = $1 ORDER BY "incidents"."id" LIMIT 1`)
		expectedIncidentInsert = regexp.
					QuoteMeta(`INSERT INTO "incidents" ("id","display_name","description","began_at","ended_at","phase_generation","phase_order") VALUES ($1,$2,$3,$4,$5,$6,$7)`) //nolint:lll
		expectedIncidentDelete = regexp.
					QuoteMeta(`DELETE FROM "incidents" WHERE id = $1`)
		expectedIncidentUpdate = regexp.
					QuoteMeta(`UPDATE "incidents" SET "display_name"=$1 WHERE "id" = $2`)
		expectedImpactQuery = regexp.
					QuoteMeta(`SELECT * FROM "impacts" WHERE "impacts"."incident_id" = $1`)
		expectedPhaseQuery = regexp.
					QuoteMeta(`SELECT * FROM "phases" WHERE ("phases"."generation","phases"."order") IN (($1,$2))`)
		expectedIncidentUpdateQuery = regexp.
						QuoteMeta(`SELECT * FROM "incident_updates" WHERE "incident_updates"."incident_id" = $1`)

		// incident time - 5 minutes ago
		incidentHappened = now.Add(-5 * time.Minute)

		// filled test incident
		incident = db.Incident{
			Model: db.Model{
				ID: &incidentUUID,
			},
			DisplayName:     test.Ptr("Disk impact"),
			Description:     test.Ptr("Disk IO low"),
			Affects:         &[]db.Impact{},
			BeganAt:         &incidentHappened,
			PhaseGeneration: test.Ptr(1),
			PhaseOrder:      test.Ptr(0),
			Updates:         &[]db.IncidentUpdate{},
		}
	)

	BeforeEach(func() {
		// setup database and mock before each test
		var gormDB *gorm.DB

		sqlDB, sqlMock, gormDB = test.MustMockGorm(gormLogger)
		handlers = server.New(gormDB, handlerLogger)

		// create mock rows before each test
		incidentRows = sqlmock.
			NewRows([]string{"id", "display_name", "description", "began_at", "ended_at", "phase_generation", "phase_order"})

		impactRows = sqlmock.
			NewRows([]string{"incident_id", "incident_id", "impact_type_id"})

		componentRows = sqlmock.
			NewRows([]string{"id", "display_name", "labels"})

		phaseRows = sqlmock.
			NewRows([]string{"name, generation, order"})

		incidentUpdateRows = sqlmock.
			NewRows([]string{"incident_id", "order", "display_name", "description", "created_at"})
	})

	AfterEach(func() {
		// check every expectation after each test and close database
		Ω(sqlMock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("GetIncidents", func() {
		var (
			ctx               echo.Context
			res               *httptest.ResponseRecorder
			startTime         = now.Add(-10 * time.Minute) // 10 minutes ago
			endTime           = now.Add(5 * time.Minute)   // in 5 minutes
			getIncidentParams = api.GetIncidentsParams{
				Start: startTime,
				End:   endTime,
			}
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				"/incidents",
				nil,
			)
		})

		Context("without data", func() {
			It("should return an empty list of incidents", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedIncidentsQuery).
					WithArgs(startTime, startTime, endTime, endTime, endTime).
					WillReturnRows(incidentRows)

				expectedResult, _ := json.Marshal(api.IncidentListResponse{
					Data: []api.IncidentResponseData{},
				})

				// Act
				err := handlers.GetIncidents(ctx, getIncidentParams)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with valid data", func() {
			It("should return a list of incidents", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedIncidentsQuery).
					WithArgs(startTime, startTime, endTime, endTime, endTime).
					WillReturnRows(
						incidentRows.AddRow(
							incident.ID,              // id
							incident.DisplayName,     // display_name
							incident.Description,     // description
							incident.BeganAt,         // began_at
							incident.EndedAt,         // ended_at
							incident.PhaseGeneration, // phase_generation
							incident.PhaseOrder,      // phase_order
						),
					)
				sqlMock.
					ExpectQuery(expectedImpactQuery).
					WithArgs(incidentID).
					WillReturnRows(impactRows, componentRows)
				sqlMock.
					ExpectQuery(expectedPhaseQuery).
					WithArgs(incident.PhaseGeneration, incident.PhaseOrder).
					WillReturnRows(phaseRows)
				sqlMock.
					ExpectQuery(expectedIncidentUpdateQuery).
					WithArgs(incidentID).
					WillReturnRows(incidentUpdateRows)

				expectedResult, _ := json.Marshal(api.IncidentListResponse{
					Data: []api.IncidentResponseData{
						incident.ToAPIResponse(),
					},
				})

				// Act
				err := handlers.GetIncidents(ctx, getIncidentParams)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with invalid parameters", func() {
			Context("with missing start", func() {
				It("should return 400 bad request", func() {
					// Arrange
					ctx, res = test.MustCreateEchoContextAndResponseWriter(
						echoLogger,
						http.MethodGet,
						"/incidents",
						nil,
					)

					getIncidentParamsMissingStart := api.GetIncidentsParams{
						End: endTime,
					}

					// Act
					err := handlers.GetIncidents(ctx, getIncidentParamsMissingStart)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrBadRequest))
				})
			})

			Context("with missing end", func() {
				It("should return 400 bad request", func() {
					// Arrange
					ctx, res = test.MustCreateEchoContextAndResponseWriter(
						echoLogger,
						http.MethodGet,
						"/incidents",
						nil,
					)

					getIncidentParamsMissingEnd := api.GetIncidentsParams{
						Start: startTime,
					}

					// Act
					err := handlers.GetIncidents(ctx, getIncidentParamsMissingEnd)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrBadRequest))
				})
			})

			Context("with end before start", func() {
				It("should return 400 bad request", func() {
					// Arrange
					ctx, res = test.MustCreateEchoContextAndResponseWriter(
						echoLogger,
						http.MethodGet,
						"/incidents",
						nil,
					)

					getIncidentParamsStartEndSwapped := api.GetIncidentsParams{
						Start: endTime,
						End:   startTime,
					}

					// Act
					err := handlers.GetIncidents(ctx, getIncidentParamsStartEndSwapped)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrBadRequest))
				})
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedIncidentsQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetIncidents(ctx, getIncidentParams)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("CreateIncident", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPost,
				"/incidents",
				api.Incident{
					DisplayName: test.Ptr("Disk impact"),
				},
			)
		})

		Context("with valid request", func() {
			It("should create a incident and return its UUID", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedIncidentInsert).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.CreateIncident(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())

				// parse answer to get uuid
				var response api.IdResponse
				err = json.Unmarshal(res.Body.Bytes(), &response)

				Ω(err).ShouldNot(HaveOccurred())

				// check valid uuid
				_, err = uuid.Parse(response.Id)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusCreated))
			})
		})

		Context("with empty request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(echoLogger, http.MethodPost, "/incidents", nil)

				// Act
				err := handlers.CreateIncident(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedIncidentInsert).WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.CreateIncident(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("with invalid request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(echoLogger, http.MethodPost, "/incidents", api.Incident{
					DisplayName: test.Ptr("Disk impact"),
					Affects: &api.ImpactComponentList{
						{
							Reference: test.Ptr("ABC"), // invalid UUID
						},
					},
				})

				// Act
				err := handlers.CreateIncident(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})
	})

	Describe("DeleteIncident", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodDelete,
				fmt.Sprintf("/incidents/%s", incidentID),
				nil,
			)
		})

		Context("with valid UUID and an affected row", func() {
			It("should return 204 no content", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedIncidentDelete).WithArgs(incidentID).WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteIncident(ctx, incidentID)

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
				sqlMock.ExpectExec(expectedIncidentDelete).WithArgs(incidentID).WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteIncident(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedIncidentDelete).WithArgs(incidentID).WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.DeleteIncident(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("with invalid UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodDelete,
					"/incidents/ABC-123",
					nil,
				)
				// Act
				err := handlers.DeleteIncident(ctx, "ABC-123")

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})
	})

	Describe("GetIncident", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				fmt.Sprintf("/incidents/%s", incidentID),
				nil,
			)
		})

		Context("with valid UUID and valid data", func() {
			It("should return a single incident", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedIncidentQuery).
					WithArgs(incidentID).
					WillReturnRows(
						incidentRows.AddRow(
							incident.ID,
							incident.DisplayName,
							incident.Description,
							incident.BeganAt,
							incident.EndedAt,
							incident.PhaseGeneration,
							incident.PhaseOrder,
						),
					)
				sqlMock.
					ExpectQuery(expectedImpactQuery).
					WithArgs(incidentID).
					WillReturnRows(impactRows, incidentRows)
				sqlMock.
					ExpectQuery(expectedPhaseQuery).
					WithArgs(incident.PhaseGeneration, incident.PhaseOrder).
					WillReturnRows(phaseRows)
				sqlMock.
					ExpectQuery(expectedIncidentUpdateQuery).
					WithArgs(incidentID).
					WillReturnRows(incidentUpdateRows)

				expectedResult, _ := json.Marshal(api.IncidentResponse{
					Data: incident.ToAPIResponse(),
				})

				// Act
				err := handlers.GetIncident(ctx, incidentID)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedIncidentQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetIncident(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("with invalid UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodGet,
					"/incidents/ABC-123",
					nil,
				)

				// Act
				err := handlers.GetIncident(ctx, "ABC-123")

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("without data", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedIncidentQuery).WillReturnRows(incidentRows)

				// Act
				err := handlers.GetIncident(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})

	Describe("UpdateIncident", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPatch,
				fmt.Sprintf("/incidents/%s", incidentID),
				api.Incident{
					DisplayName: test.Ptr("Network impact"),
				},
			)
		})

		Context("with valid UUID and valid request", func() {
			It("should return 204 no conntent", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedIncidentUpdate).
					WithArgs("Network impact", incidentID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateIncident(ctx, incidentID)

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
					fmt.Sprintf("/incidents/%s", incidentID),
					nil,
				)

				// Act
				err := handlers.UpdateIncident(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with invalid UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodPatch,
					fmt.Sprintf("/incidents/%s", "ABC-123"),
					api.Incident{
						DisplayName: test.Ptr("Network impact"),
					},
				)

				// Act
				err := handlers.UpdateIncident(ctx, "ABC-123")

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
					ExpectExec(expectedIncidentUpdate).
					WithArgs("Network impact", incidentID).
					WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.UpdateIncident(ctx, incidentID)

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
					ExpectExec(expectedIncidentUpdate).
					WithArgs("Network impact", incidentID).
					WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateIncident(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})

		Context("with invalid request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodPatch,
					fmt.Sprintf("/incidents/%s", incidentID),
					api.Incident{
						DisplayName: test.Ptr("Disk impact"),
						Affects: &api.ImpactComponentList{
							{
								Reference: test.Ptr("ABC"), // invalid UUID
							},
						},
					},
				)

				// Act
				err := handlers.UpdateIncident(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})
	})
})

var _ = Describe("IncidentUpdate", func() {
	const incidentUpdateOrder = 0

	var (
		// mocked sql rows
		incidentUpdateRows             *sqlmock.Rows
		highestIncidentUpdateOrderRows *sqlmock.Rows

		// expected SQL
		expectedIncidentUpdatesQuery = regexp.
						QuoteMeta(`SELECT * FROM "incident_updates" WHERE incident_id = $1`)
		expectedIncidentUpdateQuery = regexp.
						QuoteMeta(`SELECT * FROM "incident_updates" WHERE incident_id = $1 AND "order" = $2 ORDER BY "incident_updates"."incident_id" LIMIT 1`) //nolint:lll
		expectedIncidentUpdateInsert = regexp.
						QuoteMeta(`INSERT INTO "incident_updates" ("incident_id","order","display_name","description","created_at") VALUES ($1,$2,$3,$4,$5)`) //nolint:lll
		expectedIncidentUpdateDelete = regexp.
						QuoteMeta(`DELETE FROM "incident_updates" WHERE incident_id = $1 AND "order" = $2`)
		expectedIncidentUpdateUpdate = regexp.
						QuoteMeta(`UPDATE "incident_updates" SET "description"=$1 WHERE "incident_id" = $2 AND "order" = $3`)
		expectedHighestIncidentUpdateOrderQuery = regexp.
							QuoteMeta(`SELECT COALESCE(MAX("order"), -1) FROM "incident_updates" WHERE incident_id = $1`)

		// filled test incidentUpdate
		incidentUpdate = db.IncidentUpdate{
			IncidentID:  &incidentUUID,
			Order:       test.Ptr(0),
			DisplayName: test.Ptr("Investigation started"),
			Description: test.Ptr("We started to investigate the impact."),
			CreatedAt:   &now,
		}
	)

	BeforeEach(func() {
		// setup database and mock before each test
		var gormDB *gorm.DB

		sqlDB, sqlMock, gormDB = test.MustMockGorm(gormLogger)
		handlers = server.New(gormDB, handlerLogger)

		// create mock rows before each test
		incidentUpdateRows = sqlmock.
			NewRows([]string{"incident_id", "order", "display_name", "description", "created_at"})

		highestIncidentUpdateOrderRows = sqlmock.NewRows([]string{"coalesce"})
	})

	AfterEach(func() {
		// check every expectation after each test and close database
		Ω(sqlMock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("GetIncidentUpdates", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				fmt.Sprintf("/incidentupdates/%s", incidentID),
				nil,
			)
		})

		Context("without data", func() {
			It("should return an empty list of incident updates", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedIncidentUpdatesQuery).WillReturnRows(incidentUpdateRows)

				expectedResult, _ := json.Marshal(api.IncidentUpdateListResponse{
					Data: []api.IncidentUpdateResponseData{},
				})

				// Act
				err := handlers.GetIncidentUpdates(ctx, incidentID)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with valid data", func() {
			It("should return a list of incidentUpdates", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedIncidentUpdatesQuery).
					WillReturnRows(
						incidentUpdateRows.AddRow(
							incidentUpdate.IncidentID,  // incident_id
							incidentUpdate.Order,       // order
							incidentUpdate.DisplayName, // display_name
							incidentUpdate.Description, // description
							incidentUpdate.CreatedAt,   // created_at
						),
					)

				expectedResult, _ := json.Marshal(api.IncidentUpdateListResponse{
					Data: []api.IncidentUpdateResponseData{
						incidentUpdate.ToAPIResponse(),
					},
				})

				// Act
				err := handlers.GetIncidentUpdates(ctx, incidentID)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with invalid incident UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodGet,
					"/incidentupdates/ABC-123",
					nil,
				)

				// Act
				err := handlers.GetIncidentUpdates(ctx, "ABC-123")

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			It("should return 500 internal server error", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedIncidentUpdatesQuery).WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetIncidentUpdates(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})
	})

	Describe("CreateIncidentUpdate", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPost,
				fmt.Sprintf("/incidentupdates/%s", incidentID),
				api.IncidentUpdate{
					DisplayName: test.Ptr("Investigation started"),
					Description: test.Ptr("We started to investigate the impact."),
				},
			)
		})

		Context("with valid request", func() {
			It("should create a incidentUpdate and return its order", func() {
				// Arrange
				highestOrder := -1

				sqlMock.ExpectBegin()
				sqlMock.
					ExpectQuery(expectedHighestIncidentUpdateOrderQuery).
					WithArgs(incidentID).
					WillReturnRows(highestIncidentUpdateOrderRows.AddRow(highestOrder))
				sqlMock.
					ExpectExec(expectedIncidentUpdateInsert).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.CreateIncidentUpdate(ctx, incidentID)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())

				// parse answer to get order
				var response api.OrderResponse
				err = json.Unmarshal(res.Body.Bytes(), &response)

				Ω(err).ShouldNot(HaveOccurred())

				Ω(response.Order).Should(Equal(highestOrder + 1))
				Ω(res.Code).Should(Equal(http.StatusCreated))
			})
		})

		Context("with empty request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodPost,
					fmt.Sprintf("/incidentupdates/%s", incidentID),
					nil,
				)

				// Act
				err := handlers.CreateIncidentUpdate(ctx, incidentID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with invalid incident UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodPost,
					"/incidentupdates/ABC-123",
					api.IncidentUpdate{
						DisplayName: test.Ptr("Investigation started"),
						Description: test.Ptr("We started to investigate the impact."),
					},
				)

				// Act
				err := handlers.CreateIncidentUpdate(ctx, "ABC-123")

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			Context("while getting highest order", func() {
				It("should return 500 internal server error", func() {
					// Arrange
					sqlMock.ExpectBegin()
					sqlMock.
						ExpectQuery(expectedHighestIncidentUpdateOrderQuery).
						WithArgs(incidentID).
						WillReturnError(test.ErrTestError)
					sqlMock.ExpectRollback()

					// Act
					err := handlers.CreateIncidentUpdate(ctx, incidentID)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrInternalServerError))
				})
			})

			Context("while saving incident update", func() {
				It("should return 500 internal server error", func() {
					// Arrange
					highestOrder := -1

					sqlMock.ExpectBegin()
					sqlMock.
						ExpectQuery(expectedHighestIncidentUpdateOrderQuery).
						WithArgs(incidentID).
						WillReturnRows(highestIncidentUpdateOrderRows.AddRow(highestOrder))
					sqlMock.
						ExpectExec(expectedIncidentUpdateInsert).
						WillReturnError(test.ErrTestError)
					sqlMock.ExpectRollback()

					// Act
					err := handlers.CreateIncidentUpdate(ctx, incidentID)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrInternalServerError))
				})
			})
		})
	})

	Describe("DeleteIncidentUpdate", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodDelete,
				fmt.Sprintf("/incidentupdates/%s/%d", incidentID, incidentUpdateOrder),
				nil,
			)
		})

		Context("with valid incident UUID and an affected row", func() {
			It("should return 204 no content", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.ExpectExec(expectedIncidentUpdateDelete).
					WithArgs(incidentID, incidentUpdateOrder).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

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
					ExpectExec(expectedIncidentUpdateDelete).
					WithArgs(incidentID, incidentUpdateOrder).
					WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.DeleteIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

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
					ExpectExec(expectedIncidentUpdateDelete).
					WithArgs(incidentID, incidentUpdateOrder).
					WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.DeleteIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("with invalid incident UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodDelete,
					fmt.Sprintf("/incidentupdates/ABC-123/%d", incidentUpdateOrder),
					nil,
				)
				// Act
				err := handlers.DeleteIncidentUpdate(ctx, "ABC-123", incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})
	})

	Describe("GetIncidentUpdate", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				fmt.Sprintf("/incidentupdates/%s/%d", incidentID, incidentUpdateOrder),
				nil,
			)
		})

		Context("with valid incident UUID and valid data", func() {
			It("should return a single incidentUpdate", func() {
				// Arrange
				sqlMock.
					ExpectQuery(expectedIncidentUpdateQuery).
					WithArgs(incidentID, incidentUpdateOrder).
					WillReturnRows(
						incidentUpdateRows.AddRow(
							incidentUpdate.IncidentID,  // incident_id
							incidentUpdate.Order,       // order
							incidentUpdate.DisplayName, // display_name
							incidentUpdate.Description, // description
							incidentUpdate.CreatedAt,   // created_at
						),
					)

				expectedResult, _ := json.Marshal(api.IncidentUpdateResponse{
					Data: incidentUpdate.ToAPIResponse(),
				})

				// Act
				err := handlers.GetIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

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
					ExpectQuery(expectedIncidentUpdateQuery).
					WithArgs(incidentID, incidentUpdateOrder).
					WillReturnError(test.ErrTestError)

				// Act
				err := handlers.GetIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrInternalServerError))
			})
		})

		Context("with invalid incident UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodGet,
					fmt.Sprintf("/incidentupdates/ABC-123/%d", incidentUpdateOrder),
					nil,
				)

				// Act
				err := handlers.GetIncidentUpdate(ctx, "ABC-123", incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("without data", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectQuery(expectedIncidentUpdateQuery).WillReturnRows(incidentUpdateRows)

				// Act
				err := handlers.GetIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})

	Describe("UpdateIncidentUpdate", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPatch,
				fmt.Sprintf("/incidentupdates/%s/%d", incidentID, incidentUpdateOrder),
				api.IncidentUpdate{
					Description: test.Ptr("NIC was down"),
				},
			)
		})

		Context("with valid incident UUID and valid request", func() {
			It("should return 204 no conntent", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectExec(expectedIncidentUpdateUpdate).
					WithArgs("NIC was down", incidentID, incidentUpdateOrder).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusNoContent))
				Ω(res.Body.String()).Should(Equal(""))
			})
		})

		Context("with empty request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodPatch,
					fmt.Sprintf("/incidentupdates/%s/%d", incidentID, incidentUpdateOrder),
					nil,
				)

				// Act
				err := handlers.UpdateIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with invalid incident UUID", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, res = test.MustCreateEchoContextAndResponseWriter(
					echoLogger,
					http.MethodPatch,
					fmt.Sprintf("/incidentupdates/ABC-123/%d", incidentUpdateOrder),
					api.IncidentUpdate{
						Description: test.Ptr("NIC was down"),
					},
				)

				// Act
				err := handlers.UpdateIncidentUpdate(ctx, "ABC-123", incidentUpdateOrder)

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
					ExpectExec(expectedIncidentUpdateUpdate).
					WithArgs("NIC was down", incidentID, incidentUpdateOrder).
					WillReturnError(test.ErrTestError)
				sqlMock.ExpectRollback()

				// Act
				err := handlers.UpdateIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

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
					ExpectExec(expectedIncidentUpdateUpdate).
					WithArgs("NIC was down", incidentID, incidentUpdateOrder).
					WillReturnResult(sqlmock.NewResult(0, 0))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.UpdateIncidentUpdate(ctx, incidentID, incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})
	})
})
