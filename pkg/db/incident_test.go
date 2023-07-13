package db_test

import (
	"time"

	"github.com/SovereignCloudStack/status-page-api/internal/app/util/test"
	"github.com/SovereignCloudStack/status-page-api/pkg/db"
	"github.com/SovereignCloudStack/status-page-openapi/pkg/api"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const incidentID = "91fd8fa3-4288-4940-bcfb-9e89d82f3522"

var (
	incidentUUID = uuid.MustParse(incidentID)

	incidentUpdateOrder = 0

	now           = time.Now()
	updateCreated = now.Add(-5 * time.Minute)
)

var _ = Describe("Incident", func() {
	const (
		componentID  = "7fecf595-6352-4906-a0d8-b3243ee62ec8"
		impactTypeID = "c3fc130d-e6c4-4f94-86ba-e51fbdfc5d0c"
	)

	var (
		componentUUID  = uuid.MustParse(componentID)
		impactTypeUUID = uuid.MustParse(impactTypeID)

		beganAt = now.Add(-10 * time.Minute)
		endedAt = now.Add(5 * time.Minute)

		phaseGeneration = 1
		phaseOrder      = 1

		incident = db.Incident{
			Model: db.Model{
				ID: &incidentUUID,
			},
			DisplayName: test.Ptr("Disk incident"),
			Description: test.Ptr("Disk performance decrease."),
			Affects: &[]db.Impact{
				{
					IncidentID:   &incidentUUID,
					ComponentID:  &componentUUID,
					ImpactTypeID: &impactTypeUUID,
				},
			},
			BeganAt:         &beganAt,
			EndedAt:         &endedAt,
			PhaseGeneration: &phaseGeneration,
			PhaseOrder:      &phaseOrder,
			Phase: &db.Phase{
				Name:       test.Ptr("Investigation ongoing"),
				Generation: &phaseGeneration,
				Order:      &phaseOrder,
			},
			Updates: &[]db.IncidentUpdate{
				{
					IncidentID:  &incidentUUID,
					Order:       &incidentUpdateOrder,
					DisplayName: test.Ptr("Investigation started"),
					Description: test.Ptr("We started to investigate the issue."),
					CreatedAt:   &updateCreated,
				},
			},
		}
	)

	Describe("ToAPIResponse", func() {
		Context("with valid data", func() {
			It("should return the api response", func() {
				// Arrange
				expectedResult := api.IncidentResponseData{
					Affects: &api.ImpactComponentList{
						{
							Reference: test.Ptr(componentID),
							Type:      test.Ptr(impactTypeID),
						},
					},
					BeganAt:     &beganAt,
					DisplayName: test.Ptr("Disk incident"),
					Description: test.Ptr("Disk performance decrease."),
					EndedAt:     &endedAt,
					Id:          incidentID,
					Phase: &api.PhaseReference{
						Generation: phaseGeneration,
						Order:      phaseOrder,
					},
					Updates: &api.IncrementalList{
						incidentUpdateOrder,
					},
				}

				// Act
				res := incident.ToAPIResponse()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("GetImpactComponentList", func() {
		Context("with valid data", func() {
			It("should return a list of impacts", func() {
				// Arrange
				expectedResult := &api.ImpactComponentList{
					{
						Reference: test.Ptr(componentID),
						Type:      test.Ptr(impactTypeID),
					},
				}

				// Act
				res := incident.GetImpactComponentList()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("GetIncidentUpdates", func() {
		Context("with valid data", func() {
			It("should return a list of incident updates", func() {
				// Arrange
				expectedResult := &api.IncrementalList{
					incidentUpdateOrder,
				}

				// Act
				res := incident.GetIncidentUpdates()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("IncidentFromAPI", func() {
		Context("with valid data", func() {
			It("should return a incident", func() {
				// Arrange
				incidentRequest := &api.Incident{
					Affects: &api.ImpactComponentList{
						{
							Reference: test.Ptr(componentID),
							Type:      test.Ptr(impactTypeID),
						},
					},
					BeganAt:     &beganAt,
					DisplayName: test.Ptr("Disk incident"),
					Description: test.Ptr("Disk performance decrease."),
					EndedAt:     &endedAt,
					Phase: &api.PhaseReference{
						Generation: phaseGeneration,
						Order:      phaseOrder,
					},
					Updates: &api.IncrementalList{
						incidentUpdateOrder,
					},
				}

				expectedResult := &db.Incident{
					DisplayName: test.Ptr("Disk incident"),
					Description: test.Ptr("Disk performance decrease."),
					Affects: &[]db.Impact{
						{
							ComponentID:  &componentUUID,
							ImpactTypeID: &impactTypeUUID,
						},
					},
					BeganAt: &beganAt,
					EndedAt: &endedAt,
					Phase: &db.Phase{
						Generation: &phaseGeneration,
						Order:      &phaseOrder,
					},
				}

				// Act
				res, err := db.IncidentFromAPI(incidentRequest)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})

		Context("with no data", func() {
			It("should return an ErrEmptyValue", func() {
				// Arrange
				// Act
				res, err := db.IncidentFromAPI(nil)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})
})

var _ = Describe("IncidentUpdate", func() {
	incidentUpdate := db.IncidentUpdate{
		IncidentID:  &incidentUUID,
		Order:       &incidentUpdateOrder,
		DisplayName: test.Ptr("Investigation started"),
		Description: test.Ptr("We started to investigate the issue."),
		CreatedAt:   &updateCreated,
	}

	Describe("ToAPIResponse", func() {
		Context("with valid data", func() {
			It("should return the api response", func() {
				// Arrange
				expectedResult := api.IncidentUpdateResponseData{
					CreatedAt:   &updateCreated,
					DisplayName: test.Ptr("Investigation started"),
					Description: test.Ptr("We started to investigate the issue."),
					Order:       incidentUpdateOrder,
				}

				// Act
				res := incidentUpdate.ToAPIResponse()

				// Assert
				Ω(res).Should(Equal(expectedResult))
			})
		})
	})

	Describe("IncidentUpdateFromAPI", func() {
		Context("with valid data", func() {
			It("should return a incident update", func() {
				// Arrange
				incidentUpdateRequest := &api.IncidentUpdate{
					CreatedAt:   &updateCreated,
					DisplayName: test.Ptr("Investigation started"),
					Description: test.Ptr("We started to investigate the issue."),
				}

				expectedResult := &db.IncidentUpdate{
					IncidentID:  &incidentUUID,
					Order:       &incidentUpdateOrder,
					DisplayName: test.Ptr("Investigation started"),
					Description: test.Ptr("We started to investigate the issue."),
					CreatedAt:   &updateCreated,
				}

				// Act
				res, err := db.IncidentUpdateFromAPI(incidentUpdateRequest, incidentUUID, incidentUpdateOrder)

				// Assert
				Ω(err).ShouldNot(HaveOccurred())
				Ω(res).Should(Equal(expectedResult))
			})
		})

		Context("with no data", func() {
			It("should return an ErrEmptyValue", func() {
				// Arrange
				// Act
				res, err := db.IncidentUpdateFromAPI(nil, incidentUUID, incidentUpdateOrder)

				// Assert
				Ω(err).Should(HaveOccurred())
				Ω(err).Should(Equal(db.ErrEmptyValue))
				Ω(res).Should(BeNil())
			})
		})
	})
})
