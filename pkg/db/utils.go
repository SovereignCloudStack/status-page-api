package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GetHighestIncidentUpdateOrder retrieves the currently highest order for an incident.
func GetHighestIncidentUpdateOrder(dbCon *gorm.DB, incidentID uuid.UUID) (int, error) {
	var order int
	res := dbCon.
		Model(&IncidentUpdate{}). //nolint:exhaustruct
		Select("COALESCE(MAX(\"order\"), -1)").
		Where("incident_id = ?", incidentID).
		Find(&order)

	return order, res.Error
}

// GetCurrentPhaseGeneration retrieves the currently highest generation.
func GetCurrentPhaseGeneration(dbCon *gorm.DB) (int, error) {
	var generation int
	res := dbCon.
		Model(&Phase{}). //nolint:exhaustruct
		Select("COALESCE(MAX(generation), 0)").
		Find(&generation)

	return generation, res.Error
}
