package db_test

import (
	"database/sql"
	"regexp"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

var _ = Describe("Utils", func() {
	const incidentID = "91fd8fa3-4288-4940-bcfb-9e89d82f3522"

	var (
		// sub loggers
		_, gormLogger, _ = test.MustSetupLogging(zerolog.TraceLevel)

		// sql mocking
		sqlDB   *sql.DB
		sqlMock sqlmock.Sqlmock
		gormDB  *gorm.DB

		incidentUUID = uuid.MustParse(incidentID)
	)

	BeforeEach(func() {
		// setup database and mock before each test
		sqlDB, sqlMock, gormDB = test.MustMockGorm(gormLogger)
	})

	AfterEach(func() {
		// check every expectation after each test and close database
		Ω(sqlMock.ExpectationsWereMet()).ShouldNot(HaveOccurred())
		sqlDB.Close()
	})

	Describe("GetHighestIncidentUpdateOrder", func() {
		var (
			highestIncidentUpdateOrderRows *sqlmock.Rows

			expectedHighestIncidentUpdateOrderQuery = regexp.
								QuoteMeta(`SELECT COALESCE(MAX("order"), -1) FROM "incident_updates" WHERE incident_id = $1`)
		)

		BeforeEach(func() {
			highestIncidentUpdateOrderRows = sqlmock.NewRows([]string{"coalesce"})
		})

		Context("with valid data", func() {
			It("should return highest incident update order for incident", func() {
				// Arrange
				incidentUpdateOrder = 0
				sqlMock.
					ExpectQuery(expectedHighestIncidentUpdateOrderQuery).
					WithArgs(incidentUUID).
					WillReturnRows(highestIncidentUpdateOrderRows.AddRow(incidentUpdateOrder))

				// Act
				res, err := db.GetHighestIncidentUpdateOrder(gormDB, incidentUUID)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(incidentUpdateOrder))
			})
		})

		Context("with database error", func() {
			It("should return highest incident update order", func() {
				// Arrange
				incidentUpdateOrder = 0
				sqlMock.
					ExpectQuery(expectedHighestIncidentUpdateOrderQuery).
					WithArgs(incidentUUID).
					WillReturnError(test.ErrTestError)

				// Act
				_, err := db.GetHighestIncidentUpdateOrder(gormDB, incidentUUID)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(test.ErrTestError))
			})
		})
	})

	Describe("GetCurrentPhaseGeneration", func() {
		var (
			lastPhaseGenerationRows *sqlmock.Rows

			expectedLastPhaseGenerationQuery = regexp.
								QuoteMeta(`SELECT COALESCE(MAX(generation), 0) FROM "phases"`)
		)

		BeforeEach(func() {
			lastPhaseGenerationRows = sqlmock.NewRows([]string{"coalesce"})
		})

		Context("with valid data", func() {
			It("should return current phase generation", func() {
				// Arrange
				incidentUpdateOrder = 1
				sqlMock.
					ExpectQuery(expectedLastPhaseGenerationQuery).
					WillReturnRows(lastPhaseGenerationRows.AddRow(incidentUpdateOrder))

				// Act
				res, err := db.GetCurrentPhaseGeneration(gormDB)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(incidentUpdateOrder))
			})
		})

		Context("with database error", func() {
			It("should return highest incident update order", func() {
				// Arrange
				incidentUpdateOrder = 0
				sqlMock.
					ExpectQuery(expectedLastPhaseGenerationQuery).
					WillReturnError(test.ErrTestError)

				// Act
				_, err := db.GetCurrentPhaseGeneration(gormDB)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(test.ErrTestError))
			})
		})
	})
})
