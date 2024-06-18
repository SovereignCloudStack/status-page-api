package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ID is the generally used identifier type. String type for UUIDs.
type ID = uuid.UUID

// Model sets the basic Data for all true database resources.
type Model struct {
	ID *ID `gorm:"primaryKey;type:uuid;"`
}

// BeforeCreate is a gorm hook to fill the ID field with a new UUID,
// before an insert statement is send to the database.
func (m *Model) BeforeCreate(_ *gorm.DB) error {
	if m.ID == nil {
		// pointer to id is nil
		id := uuid.New()
		m.ID = &id
	} else if *m.ID == uuid.Nil {
		// points to id but is empty id
		id := uuid.New()
		m.ID = &id
	}

	return nil
}
