package db

// Label represents a label with a name and value. Components can have none or many label.
type Label struct {
	Name  string `gorm:"primaryKey" json:"name"`
	Value string `gorm:"primaryKey" json:"value"`
}

// Labels is a list of [Label].
type Labels []Label
