package db_test

import (
	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Component", func() {
	const (
		componentID  = "7fecf595-6352-4906-a0d8-b3243ee62ec8"
		incidentID   = "91fd8fa3-4288-4940-bcfb-9e89d82f3522"
		impactTypeID = "c3fc130d-e6c4-4f94-86ba-e51fbdfc5d0c"
	)

	var (
		componentUUID  = uuid.MustParse(componentID)
		incidentUUID   = uuid.MustParse(incidentID)
		impactTypeUUID = uuid.MustParse(impactTypeID)
		component      = db.Component{
			Model: db.Model{
				ID: &componentUUID,
			},
			DisplayName: test.Ptr("Storage"),
			Labels:      &db.Labels{"data-center": "west", "location": "germany"},
			ActivelyAffectedBy: &[]db.Impact{
				{
					IncidentID:   &incidentUUID,
					ComponentID:  &componentUUID,
					ImpactTypeID: &impactTypeUUID,
				},
			},
		}
	)

	Describe("ToAPIResponse", func() {
		Context("with valid data", func() {
			It("should return the api response", func() {
				// Arrange
				expectedResult := api.ComponentResponseData{
					ActivelyAffectedBy: &api.ImpactIncidentList{
						{
							Reference: test.Ptr(incidentID),
							Type:      test.Ptr(impactTypeID),
						},
					},
					DisplayName: test.Ptr("Storage"),
					Id:          componentID,
					Labels:      &api.Labels{"data-center": "west", "location": "germany"},
				}

				// Act
				res := component.ToAPIResponse()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("GetImpactIncidentList", func() {
		Context("with valid data", func() {
			It("should return an impact list", func() {
				// Arrange
				expectedResult := &api.ImpactIncidentList{
					{
						Reference: test.Ptr(incidentID),
						Type:      test.Ptr(impactTypeID),
					},
				}

				// Act
				res := component.GetImpactIncidentList()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("ComponentFromAPI", func() {
		Context("with valid data", func() {
			It("should return a database component", func() {
				// Arrange
				componentRequest := &api.Component{ // ActivelyAffectedBy is read only and will never be part of a request.
					DisplayName: test.Ptr("Storage"),
					Labels:      &api.Labels{"data-center": "west", "location": "germany"},
				}

				expectedResult := &db.Component{
					DisplayName: test.Ptr("Storage"),
					Labels:      &db.Labels{"data-center": "west", "location": "germany"},
				}

				// Act
				res, err := db.ComponentFromAPI(componentRequest)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})

		Context("with no data", func() {
			It("should return an ErrEmptyValue", func() {
				// Arrange
				// Act
				res, err := db.ComponentFromAPI(nil)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})
})
