package unsafe

import (
	"errors"
	"reflect"
	"unsafe"
)

type UnsafeAccessor struct {
	fields  map[string]*FieldMeta
	address unsafe.Pointer
}

type FieldMeta struct {
	Offset uintptr
	typ    reflect.Type
}

func NewUnsafeAccessor(entity any) *UnsafeAccessor {
	typ := reflect.TypeOf(entity)
	typ = typ.Elem()
	numField := typ.NumField()
	fields := make(map[string]*FieldMeta, numField)
	for i := 0; i < numField; i++ {
		fd := typ.Field(i)
		fields[fd.Name] = &FieldMeta{
			Offset: fd.Offset,
			typ:    fd.Type,
		}
	}
	val := reflect.ValueOf(entity)
	return &UnsafeAccessor{
		fields:  fields,
		address: val.UnsafePointer(),
	}
}

func (a *UnsafeAccessor) Field(field string) (any, error) {
	fd, ok := a.fields[field]
	if !ok {
		return nil, errors.New("非法字段")
	}
	fdAddress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	return reflect.NewAt(fd.typ, fdAddress).Elem().Interface(), nil
}

func (a *UnsafeAccessor) SetField(field string, val any) error {
	fd, ok := a.fields[field]
	if !ok {
		return errors.New("非法字段")
	}
	fdAddress := unsafe.Pointer(uintptr(a.address) + fd.Offset)
	reflect.NewAt(fd.typ, fdAddress).Elem().Set(reflect.ValueOf(val))
	return nil
}
