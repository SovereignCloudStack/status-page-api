package db

type Label struct {
	Name  string `gorm:"primaryKey" json:"name"`
	Value string `gorm:"primaryKey" json:"value"`
}

type Labels []Label
