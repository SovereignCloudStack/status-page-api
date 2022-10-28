package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Labels map[string]string

func (l *Labels) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("data is of invalid type: %v", value)
	}
	return json.Unmarshal(data, l)
}

func (l Labels) Value() (driver.Value, error) {
	return json.Marshal(l)
}

type LabelFilterExpression struct {
	column string
	labels Labels
}

func LabelFilter(column string) *LabelFilterExpression {
	return &LabelFilterExpression{column: column}
}

func (lfe *LabelFilterExpression) HasLabels(labels map[string]string) *LabelFilterExpression {
	lfe.labels = labels
	return lfe
}

func (lfe *LabelFilterExpression) Build(builder clause.Builder) {
	stmt, ok := builder.(*gorm.Statement)
	if !ok {
		return
	}
	stmt.WriteQuoted(lfe.column)
	stmt.WriteString("::jsonb @> ")
	stmt.AddVar(builder, lfe.labels)
}
