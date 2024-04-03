package db_test

import (
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Severity", func() {
	Describe("NewSeverity", func() {
		Context("with valid value", func() {
			It("should return a severity", func() {
				// Arrange
				expectedResult := &db.Severity{
					DisplayName: test.Ptr("broken"),
					Value:       test.Ptr(50),
				}

				// Act
				severity, err := db.NewSeverity("broken", 50)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(severity).Should(Equal(expectedResult))
			})
		})

		Context("with low value", func() {
			It("should return an ErrSeverityValueOutOfRange", func() {
				// Arrange
				// Act
				severity, err := db.NewSeverity("broken", -1)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrSeverityValueOutOfRange))
				Ω(severity).Should(BeNil())
			})
		})

		Context("with high value", func() {
			It("should return an ErrSeverityValueOutOfRange", func() {
				// Arrange
				// Act
				severity, err := db.NewSeverity("broken", 101)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrSeverityValueOutOfRange))
				Ω(severity).Should(BeNil())
			})
		})
	})

	Describe("ToAPIResponse", func() {
		Context("with valid data", func() {
			It("should return the api response", func() {
				// Arrange
				severity, _ := db.NewSeverity("broken", 50)

				expectedResult := api.Severity{
					DisplayName: test.Ptr("broken"),
					Value:       test.Ptr(50),
				}

				// Act
				res := severity.ToAPIResponse()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("SeverityFromAPI", func() {
		Context("with valid data", func() {
			It("should return a database severity", func() {
				// Arrange
				severityRequest := &api.Severity{
					DisplayName: test.Ptr("broken"),
					Value:       test.Ptr(50),
				}

				expectedResult := &db.Severity{
					DisplayName: test.Ptr("broken"),
					Value:       test.Ptr(50),
				}

				// Act
				res, err := db.SeverityFromAPI(severityRequest)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})
		Context("with no data", func() {
			It("should return an ErrEmptyValue", func() {
				// Arrange
				// Act
				res, err := db.SeverityFromAPI(nil)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})
})
