package reflect

import (
	"fmt"
	"reflect"
)

func IterateArrayOrSlice(entity any) ([]any, error) {
	vals := reflect.ValueOf(entity)
	if vals.Type().Kind() != reflect.Array && vals.Type().Kind() != reflect.Slice {
		return nil, fmt.Errorf("不是数组或切片类型")
	}
	res := make([]any, 0, vals.Len())
	for i := 0; i < vals.Len(); i++ {
		res = append(res, vals.Index(i).Interface())
	}
	return res, nil
}

func IterateMap(entity any) ([]string, []any, error) {
	vals := reflect.ValueOf(entity)
	if vals.Type().Kind() != reflect.Map {
		return nil, nil, fmt.Errorf("不是map类型")
	}
	keys := make([]string, 0, vals.Len())
	values := make([]any, 0, vals.Len())
	itr := vals.MapRange()
	for itr.Next() {
		keys = append(keys, itr.Key().Interface().(string))
		values = append(values, itr.Value().Interface())
	}
	return keys, values, nil
}
