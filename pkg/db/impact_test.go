package db_test

import (
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
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
				expectedResult := api.ImpactTypeResponseData{
					Id:          impactTypeID,
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
				impactTypeRequest := &api.ImpactType{
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
	)

	var (
		componentUUID = uuid.MustParse(componentID)
		incidentUUID  = uuid.MustParse(incidentID)
	)

	Describe("ActivelyAffectedByFromImpactIncidentList", func() {
		Context("with valid data", func() {
			It("should return a impact list", func() {
				// Arrange
				incidentImpacts := &api.ImpactIncidentList{
					{
						Reference: test.Ptr(incidentID),
						Type:      test.Ptr(impactTypeID),
					},
				}

				expectedResult := &[]db.Impact{
					{
						IncidentID:   &incidentUUID,
						ImpactTypeID: &impactTypeUUID,
					},
				}

				// Act
				res, err := db.ActivelyAffectedByFromImpactIncidentList(incidentImpacts)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})

		Context("with no data", func() {
			It("should return an ErrEmptyValue", func() {
				// Arrange
				// Act
				res, err := db.ActivelyAffectedByFromImpactIncidentList(nil)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})

	Describe("AffectsFromImpactComponentList", func() {
		Context("with valid data", func() {
			It("should return a impact list", func() {
				// Arrange
				incidentImpacts := &api.ImpactComponentList{
					{
						Reference: test.Ptr(componentID),
						Type:      test.Ptr(impactTypeID),
					},
				}

				expectedResult := &[]db.Impact{
					{
						ComponentID:  &componentUUID,
						ImpactTypeID: &impactTypeUUID,
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
