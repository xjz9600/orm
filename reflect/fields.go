package reflect

import (
	"fmt"
	"reflect"
)

func IterateFields(entity any) (map[string]any, error) {
	if entity == nil {
		return nil, fmt.Errorf("不支持nil")
	}
	typ := reflect.TypeOf(entity)
	val := reflect.ValueOf(entity)
	val.IsValid()
	if val.IsZero() {
		return nil, fmt.Errorf("不支持零值")
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		val = val.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("不支持类型")
	}
	numFields := typ.NumField()
	res := make(map[string]any, numFields)
	for i := 0; i < numFields; i++ {
		fieldType := typ.Field(i)
		fieldValue := val.Field(i)
		if fieldType.IsExported() {
			res[fieldType.Name] = fieldValue.Interface()
		} else {
			res[fieldType.Name] = reflect.Zero(fieldType.Type).Interface()
		}
	}
	return res, nil
}

func SetField(entity any, field string, newValue any) error {
	val := reflect.ValueOf(entity)
	for val.Type().Kind() == reflect.Pointer {
		val = val.Elem()
	}
	fieldVal := val.FieldByName(field)
	if !fieldVal.CanSet() {
		return fmt.Errorf("不可被修改字段")
	}
	fieldVal.Set(reflect.ValueOf(newValue))
	return nil
}
