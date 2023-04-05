package db

type Label struct {
	Name  string `gorm:"primaryKey" json:"Name"`
	Value string `gorm:"primaryKey" json:"value"`
}

type Labels []Label
