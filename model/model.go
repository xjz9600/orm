package model

import (
	"orm/internal/errs"
	"reflect"
)

type Model struct {
	TableName string
	Fields    map[string]*Field
	Columns   map[string]*Field
	FieldArr  []*Field
}

const (
	tagKeyColumn = "column"
)

type ModelOpt func(*Model) error
type Field struct {
	ColName string
	GoName  string
	Typ     reflect.Type
	Offset  uintptr
	Alias   string
}

func WithTableName(tableName string) ModelOpt {
	return func(model *Model) error {
		model.TableName = tableName
		return nil
	}
}

func WithColumnName(field string, columnName string) ModelOpt {
	return func(model *Model) error {
		fd, ok := model.Fields[field]
		if !ok {
			return errs.NewErrUnKnownField(field)
		}
		// 注意，这里我们根本没有检测 ColName 会不会是空字符串
		// 因为正常情况下，用户都不会写错
		// 即便写错了，也很容易在测试中发现
		fd.ColName = columnName
		return nil
	}
}

type TableName interface {
	TableName() string
}
