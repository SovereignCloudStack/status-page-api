package db

type Phase struct {
	Slug string `gorm:"primaryKey" json:"slug"`
}
