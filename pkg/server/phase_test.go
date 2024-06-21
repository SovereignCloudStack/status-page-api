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
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var _ = Describe("Phase", Ordered, func() {
	var (
		// sub loggers
		echoLogger    *zerolog.Logger
		gormLogger    *zerolog.Logger
		handlerLogger *zerolog.Logger

		// sql mocking
		sqlDB   *sql.DB
		sqlMock sqlmock.Sqlmock

		// mocked sql rows
		phaseRows               *sqlmock.Rows
		lastPhaseGenerationRows *sqlmock.Rows

		// actual functions under test
		handlers *server.Implementation

		// expected SQL
		expectedPhaseListQuery = regexp.
					QuoteMeta(`SELECT * FROM "phases" WHERE generation = $1 ORDER BY "order" asc`)
		expectedLastPhaseGenerationQuery = regexp.
							QuoteMeta(`SELECT COALESCE(MAX(generation), 0) FROM "phases"`)
		expectedPhaseListInsert = regexp.
					QuoteMeta(`INSERT INTO "phases" ("name","generation","order") VALUES ($1,$2,$3),($4,$5,$6),($7,$8,$9)`) //nolint:lll

		// filled test phase list
		phaseGeneration = 1

		phases = []db.Phase{
			{
				Name:       test.Ptr("Scheduled"),
				Generation: test.Ptr(phaseGeneration),
				Order:      test.Ptr(0),
			}, {
				Name:       test.Ptr("Investigation ongoing"),
				Generation: test.Ptr(phaseGeneration),
				Order:      test.Ptr(1),
			}, {
				Name:       test.Ptr("Working on it"),
				Generation: test.Ptr(phaseGeneration),
				Order:      test.Ptr(2),
			}, {
				Name:       test.Ptr("Potential fix deployed"),
				Generation: test.Ptr(phaseGeneration),
				Order:      test.Ptr(3),
			}, {
				Name:       test.Ptr("Done"),
				Generation: test.Ptr(phaseGeneration),
				Order:      test.Ptr(4),
			},
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
		phaseRows = sqlmock.
			NewRows([]string{"name", "generation", "order"})
		lastPhaseGenerationRows = sqlmock.NewRows([]string{"coalesce"})
	})

	AfterEach(func() {
		// check every expectation after each test and close database
		Ω(sqlMock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("GetPhaseList", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder

			getPhaseListParams = apiServerDefinition.GetPhaseListParams{}
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodGet,
				"/phases",
				nil,
			)
		})

		Context("without data", func() {
			It("should return an empty list of phases", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectQuery(expectedLastPhaseGenerationQuery).
					WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
				sqlMock.ExpectQuery(expectedPhaseListQuery).WillReturnRows(phaseRows)
				sqlMock.ExpectCommit()

				expectedResult, _ := json.Marshal(apiServerDefinition.PhaseListResponse{
					Data: apiServerDefinition.PhaseListResponseData{
						Generation: phaseGeneration,
						Phases:     []apiServerDefinition.Phase{},
					},
				})

				// Act
				err := handlers.GetPhaseList(ctx, getPhaseListParams)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res.Code).Should(Equal(http.StatusOK))
				Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
			})
		})

		Context("with valid data", func() {
			BeforeEach(func() {
				phaseRows = phaseRows.
					AddRow(phases[0].Name, phases[0].Generation, phases[0].Order).
					AddRow(phases[1].Name, phases[1].Generation, phases[1].Order).
					AddRow(phases[2].Name, phases[2].Generation, phases[2].Order).
					AddRow(phases[3].Name, phases[3].Generation, phases[3].Order).
					AddRow(phases[4].Name, phases[4].Generation, phases[4].Order)
			})

			Context("with given valid generation", func() {
				It("should return a list of phases", func() {
					// Arrange
					sqlMock.ExpectBegin()
					sqlMock.
						ExpectQuery(expectedLastPhaseGenerationQuery).
						WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
					sqlMock.ExpectQuery(expectedPhaseListQuery).WillReturnRows(phaseRows)
					sqlMock.ExpectCommit()

					getPhaseListParamsGivenValidGeneration := apiServerDefinition.GetPhaseListParams{
						Generation: test.Ptr(1),
					}

					expectedResult, _ := json.Marshal(apiServerDefinition.PhaseListResponse{
						Data: apiServerDefinition.PhaseListResponseData{
							Generation: phaseGeneration,
							Phases: []apiServerDefinition.Phase{
								*phases[0].Name,
								*phases[1].Name,
								*phases[2].Name,
								*phases[3].Name,
								*phases[4].Name,
							},
						},
					})

					// Act
					err := handlers.GetPhaseList(ctx, getPhaseListParamsGivenValidGeneration)

					// Assert
					Ω(err).ShouldNot(HaveOccurred())
					Ω(res.Code).Should(Equal(http.StatusOK))
					Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
				})
			})

			Context("without given generation", func() {
				It("should return a list of phases of the current generation", func() {
					// Arrange
					sqlMock.ExpectBegin()
					sqlMock.
						ExpectQuery(expectedLastPhaseGenerationQuery).
						WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
					sqlMock.ExpectQuery(expectedPhaseListQuery).WillReturnRows(phaseRows)
					sqlMock.ExpectCommit()

					expectedResult, _ := json.Marshal(apiServerDefinition.PhaseListResponse{
						Data: apiServerDefinition.PhaseListResponseData{
							Generation: phaseGeneration,
							Phases: []apiServerDefinition.Phase{
								*phases[0].Name,
								*phases[1].Name,
								*phases[2].Name,
								*phases[3].Name,
								*phases[4].Name,
							},
						},
					})

					// Act
					err := handlers.GetPhaseList(ctx, getPhaseListParams)

					// Assert
					Ω(err).ShouldNot(HaveOccurred())
					Ω(res.Code).Should(Equal(http.StatusOK))
					Ω(strings.Trim(res.Body.String(), "\n")).Should(Equal(string(expectedResult)))
				})
			})
		})

		Context("with given invalid generation", func() {
			It("should return 400 bad request", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectQuery(expectedLastPhaseGenerationQuery).
					WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
				sqlMock.ExpectRollback()

				getPhaseListParamsGivenInvalidGeneration := apiServerDefinition.GetPhaseListParams{
					Generation: test.Ptr(-1),
				}

				// Act
				err := handlers.GetPhaseList(ctx, getPhaseListParamsGivenInvalidGeneration)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with generation higher then current", func() {
			It("should return 404 not found", func() {
				// Arrange
				sqlMock.ExpectBegin()
				sqlMock.
					ExpectQuery(expectedLastPhaseGenerationQuery).
					WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
				sqlMock.ExpectRollback()

				getPhaseListParamsGivenLargeGeneration := apiServerDefinition.GetPhaseListParams{
					Generation: test.Ptr(15),
				}

				// Act
				err := handlers.GetPhaseList(ctx, getPhaseListParamsGivenLargeGeneration)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrNotFound))
			})
		})

		Context("with database error", func() {
			Context("while getting current generation", func() {
				It("should return 500 internal server error", func() {
					// Arrange
					sqlMock.ExpectBegin()
					sqlMock.ExpectQuery(expectedLastPhaseGenerationQuery).
						WillReturnError(test.ErrTestError)
					sqlMock.ExpectRollback()

					// Act
					err := handlers.GetPhaseList(ctx, getPhaseListParams)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrInternalServerError))
				})
			})

			Context("while getting phase list", func() {
				It("should return 500 internal server error", func() {
					// Arrange
					sqlMock.ExpectBegin()
					sqlMock.ExpectQuery(expectedLastPhaseGenerationQuery).
						WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
					sqlMock.ExpectQuery(expectedPhaseListQuery).
						WillReturnError(test.ErrTestError)
					sqlMock.ExpectRollback()

					// Act
					err := handlers.GetPhaseList(ctx, getPhaseListParams)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrInternalServerError))
				})
			})
		})
	})

	Describe("CreatePhaseList", func() {
		var (
			ctx echo.Context
			res *httptest.ResponseRecorder
		)

		BeforeEach(func() {
			// setup context and response before every test
			ctx, res = test.MustCreateEchoContextAndResponseWriter(
				echoLogger,
				http.MethodPost,
				"/phases",
				apiServerDefinition.PhaseList{
					Phases: []string{"Phase 1", "Phase 2", "Phase 3"},
				},
			)
		})

		Context("with valid request", func() {
			It("should create a phase list and return its generation", func() {
				// Arrange
				nextPhaseGeneration := phaseGeneration + 1

				sqlMock.ExpectBegin()
				sqlMock.ExpectQuery(expectedLastPhaseGenerationQuery).
					WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
				sqlMock.ExpectExec(expectedPhaseListInsert).
					WithArgs("Phase 1", nextPhaseGeneration, 0, "Phase 2", nextPhaseGeneration, 1, "Phase 3", nextPhaseGeneration, 2).
					WillReturnResult(sqlmock.NewResult(0, 1))
				sqlMock.ExpectCommit()

				// Act
				err := handlers.CreatePhaseList(ctx)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())

				// parse answer to get generation
				var response apiServerDefinition.GenerationResponse
				err = json.Unmarshal(res.Body.Bytes(), &response)

				Ω(err).ShouldNot(HaveOccurred())

				// check generation
				Ω(response.Generation).Should(Equal(phaseGeneration + 1))
				Ω(res.Code).Should(Equal(http.StatusCreated))
			})
		})

		Context("with empty request", func() {
			It("should return 400 bad request", func() {
				// Arrange
				ctx, _ = test.MustCreateEchoContextAndResponseWriter(echoLogger, http.MethodPost, "/phases", nil)

				// Act
				err := handlers.CreatePhaseList(ctx)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(echo.ErrBadRequest))
			})
		})

		Context("with database error", func() {
			Context("while getting current generation", func() {
				It("should return 500 internal server error", func() {
					// Arrange
					sqlMock.ExpectBegin()
					sqlMock.ExpectQuery(expectedLastPhaseGenerationQuery).
						WillReturnError(test.ErrTestError)
					sqlMock.ExpectRollback()

					// Act
					err := handlers.CreatePhaseList(ctx)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrInternalServerError))
				})
			})

			Context("while setting phase list", func() {
				It("should return 500 internal server error", func() {
					// Arrange
					nextPhaseGeneration := phaseGeneration + 1

					sqlMock.ExpectBegin()
					sqlMock.ExpectQuery(expectedLastPhaseGenerationQuery).
						WillReturnRows(lastPhaseGenerationRows.AddRow(phaseGeneration))
					sqlMock.ExpectExec(expectedPhaseListInsert).
						WithArgs("Phase 1", nextPhaseGeneration, 0, "Phase 2", nextPhaseGeneration, 1, "Phase 3", nextPhaseGeneration, 2).
						WillReturnError(test.ErrTestError)
					sqlMock.ExpectRollback()

					// Act
					err := handlers.CreatePhaseList(ctx)

					// Assert
					Ω(err).Should(HaveOccurred())
					Ω(err).Should(Equal(echo.ErrInternalServerError))
				})
			})
		})
	})
})
