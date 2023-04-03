package db

type Component struct {
	Id          Id         `gorm:"primaryKey" json:"id"`
	AffectedBy  []Incident `gorm:"many2many:component_incidents" json:"affectedBy"`
	DisplayName string     `json:"displayName"`
	Labels      []Label    `gorm:"many2many:component_labels" json:"labels"`
}
