package db

type Label struct {
	Slug  string `gorm:"primaryKey" json:"slug"`
	Value string `gorm:"primaryKey" json:"value"`
}

type Labels []Label
