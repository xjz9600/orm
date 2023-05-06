package orm

import (
	"orm/internal/errs"
	"reflect"
	"unicode"
)

type model struct {
	tableName string
	fields    map[string]*field
}

type field struct {
	colName string
}

func parseModel(entity any) (*model, error) {
	typ := reflect.TypeOf(entity)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errs.NewErrUnSupportType(typ.Kind())
	}
	numField := typ.NumField()
	fields := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fieldType := typ.Field(i)
		fields[fieldType.Name] = &field{
			colName: underscoreName(fieldType.Name),
		}
	}
	return &model{
		tableName: underscoreName(typ.Name()),
		fields:    fields,
	}, nil
}

// underscoreName 驼峰转字符串命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}
