package db_test

import (
	"encoding/json"
	"errors"

	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Label", func() {
	Describe("Scan", func() {
		Context("with valid data", func() {
			It("should parse json", func() {
				// Arrange
				labels := db.Labels{}
				expectedResult := db.Labels{"location": "west"}
				data, _ := json.Marshal(expectedResult)

				// Act
				err := labels.Scan(data)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(labels).Should(Equal(expectedResult))
			})
		})

		Context("with invalid data", func() {
			It("should return ErrInvalidLabelData", func() {
				// Arrange
				labels := db.Labels{}
				data := 842376

				// Act
				err := labels.Scan(data)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(errors.Unwrap(err)).Should(Equal(db.ErrInvalidLabelData))
			})
		})
	})
})
