package db_test

import (
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	apiServerDefinition "github.com/SovereignCloudStack/status-page-openapi/pkg/api/server"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const impactTypeID = "c3fc130d-e6c4-4f94-86ba-e51fbdfc5d0c"

var impactTypeUUID = uuid.MustParse(impactTypeID)

var _ = Describe("ImpactType", func() {
	impactType := db.ImpactType{
		Model: db.Model{
			ID: &impactTypeUUID,
		},
		DisplayName: test.Ptr("Performance Degration"),
		Description: test.Ptr("Performance has been down."),
	}

	Describe("ToAPIResponse", func() {
		Context("with valid data", func() {
			It("should return the api response", func() {
				// Arrange
				expectedResult := apiServerDefinition.ImpactTypeResponseData{
					Id:          impactTypeUUID,
					DisplayName: test.Ptr("Performance Degration"),
					Description: test.Ptr("Performance has been down."),
				}

				// Act
				res := impactType.ToAPIResponse()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("ImpactTypeFromAPI", func() {
		Context("with valid data", func() {
			It("should return the impact type", func() {
				// Arrange
				impactTypeRequest := &apiServerDefinition.ImpactType{
					DisplayName: test.Ptr("Performance Degration"),
					Description: test.Ptr("Performance has been down."),
				}

				expectedResult := &db.ImpactType{
					DisplayName: test.Ptr("Performance Degration"),
					Description: test.Ptr("Performance has been down."),
				}

				// Act
				res, err := db.ImpactTypeFromAPI(impactTypeRequest)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})

		Context("with no data", func() {
			It("should return an ErrEmptyValue", func() {
				// Arrange
				// Act
				res, err := db.ImpactTypeFromAPI(nil)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})
})

var _ = Describe("Impact", func() {
	const (
		componentID = "7fecf595-6352-4906-a0d8-b3243ee62ec8"
		incidentID  = "91fd8fa3-4288-4940-bcfb-9e89d82f3522"
		severity    = 50
	)

	componentUUID := uuid.MustParse(componentID)

	Describe("AffectsFromImpactComponentList", func() {
		Context("with valid data", func() {
			It("should return a impact list", func() {
				// Arrange
				incidentImpacts := &apiServerDefinition.ImpactComponentList{
					{
						Reference: &componentUUID,
						Type:      &impactTypeUUID,
						Severity:  test.Ptr(severity),
					},
				}

				expectedResult := &[]db.Impact{
					{
						ComponentID:  &componentUUID,
						ImpactTypeID: &impactTypeUUID,
						Severity:     test.Ptr(severity),
					},
				}

				// Act
				res, err := db.AffectsFromImpactComponentList(incidentImpacts)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})

		Context("with no data", func() {
			It("should return an ErrEmptyValue", func() {
				// Arrange
				// Act
				res, err := db.AffectsFromImpactComponentList(nil)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})
})
