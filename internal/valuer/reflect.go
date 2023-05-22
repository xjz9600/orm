package valuer

import (
	"database/sql"
	"orm/internal/errs"
	"orm/model"
	"reflect"
)

type reflectValue struct {
	model *model.Model
	val   reflect.Value
}

var _ Creator = NewReflectValue

func NewReflectValue(model *model.Model, val any) Value {
	return &reflectValue{
		model: model,
		val:   reflect.ValueOf(val).Elem(),
	}
}

func (r *reflectValue) Field(name string) (any, error) {
	return r.val.FieldByName(name).Interface(), nil
}

func (r *reflectValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}
	vals := make([]any, 0, len(cs))
	valElems := make([]reflect.Value, 0, len(cs))
	for _, c := range cs {
		fd, ok := r.model.Columns[c]
		if !ok {
			return errs.NewErrUnKnownColumn(c)
		}
		val := reflect.New(fd.Typ)
		vals = append(vals, val.Interface())
		valElems = append(valElems, val.Elem())
	}
	err = rows.Scan(vals...)
	if err != nil {
		return err
	}
	for i, c := range cs {
		fd, ok := r.model.Columns[c]
		if !ok {
			return errs.NewErrUnKnownColumn(c)
		}
		r.val.FieldByName(fd.GoName).Set(valElems[i])
	}
	return nil
}
