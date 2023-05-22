package sql

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JsonColum[T any] struct {
	Val   T
	Valid bool
}

func (j *JsonColum[T]) Scan(src any) error {
	var bs []byte
	switch data := src.(type) {
	case []byte:
		bs = data
	case string:
		bs = []byte(data)
	case nil:
		return nil
	default:
		return fmt.Errorf("ekit：JsonColumn.Scan 不支持 src 类型 %v", src)
	}
	err := json.Unmarshal(bs, &j.Val)
	if err == nil {
		j.Valid = true
	}
	return nil
}

func (j JsonColum[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}
	return json.Marshal(j.Val)
}
