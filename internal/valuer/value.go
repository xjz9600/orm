package valuer

import (
	"database/sql"
	"orm/model"
)

type Value interface {
	SetColumns(rows *sql.Rows) error
	Field(string) (any, error)
}

type Creator func(model *model.Model, entity any) Value
