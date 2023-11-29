package db_test

import (
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Phase", func() {
	Describe("PhaseReferenceFromAPI", func() {
		Context("with valid data", func() {
			It("should return phase reference", func() {
				// Arrange
				phaseGeneration := 0
				phaseOrder := 0

				request := &api.PhaseReference{
					Generation: phaseGeneration,
					Order:      phaseOrder,
				}

				expectedResult := &db.Phase{
					Generation: &phaseGeneration,
					Order:      &phaseOrder,
				}

				// Act
				res, err := db.PhaseReferenceFromAPI(request)
				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})

		Context("with no data", func() {
			It("should return phase reference", func() {
				// Arrange
				// Act
				res, err := db.PhaseReferenceFromAPI(nil)
				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})
})
