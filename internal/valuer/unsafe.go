package valuer

import (
	"database/sql"
	"orm/internal/errs"
	"orm/model"
	"reflect"
	"unsafe"
)

type unsafeValue struct {
	model   *model.Model
	address unsafe.Pointer
}

var _ Creator = NewUnsafeValue

func NewUnsafeValue(model *model.Model, val any) Value {
	address := reflect.ValueOf(val).UnsafePointer()
	return &unsafeValue{
		model:   model,
		address: address,
	}
}

func (r *unsafeValue) Field(name string) (any, error) {
	fd, ok := r.model.Fields[name]
	if !ok {
		return nil, errs.NewErrUnKnownField(name)
	}
	fdAdress := unsafe.Pointer(uintptr(r.address) + fd.Offset)
	val := reflect.NewAt(fd.Typ, fdAdress)
	return val.Elem().Interface(), nil
}

func (r *unsafeValue) SetColumns(rows *sql.Rows) error {
	cs, err := rows.Columns()
	if err != nil {
		return err
	}

	var vals []any

	for _, c := range cs {
		fd, ok := r.model.Columns[c]
		if !ok {
			return errs.NewErrUnKnownColumn(c)
		}
		val := reflect.NewAt(fd.Typ, unsafe.Pointer(uintptr(r.address)+fd.Offset))
		vals = append(vals, val.Interface())
	}
	return rows.Scan(vals...)
}
