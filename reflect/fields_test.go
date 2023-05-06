package reflect

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"orm/reflect/types"
	"reflect"
	"testing"
)

func TestIterateFields(t *testing.T) {

	type User struct {
		Name string
		age  int
	}

	testCase := []struct {
		name    string
		entity  any
		wantErr error
		wantRes map[string]any
	}{
		{
			name:   "struct",
			entity: User{Name: "Tom", age: 18},
			wantRes: map[string]any{
				"Name": "Tom",
				"age":  0,
			}},
		{
			name:   "struct",
			entity: &User{Name: "Tom", age: 18},
			wantRes: map[string]any{
				"Name": "Tom",
				"age":  0,
			}},
		{
			name:    "basic type",
			entity:  18,
			wantErr: errors.New("不支持类型"),
		},
		{
			name: "multiple pointer",
			entity: func() **User {
				res := &User{
					Name: "Tom",
					age:  18,
				}
				return &res
			}(),
			wantRes: map[string]any{
				"Name": "Tom",
				"age":  0,
			}},
		{
			name:    "nil",
			entity:  nil,
			wantErr: errors.New("不支持nil"),
		},
		{
			name:    "nil",
			entity:  (*User)(nil),
			wantErr: errors.New("不支持零值"),
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateFields(tc.entity)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, res, tc.wantRes)
		})
	}
}

func TestSetField(t *testing.T) {

	type User struct {
		Name string
		age  int
	}

	testCase := []struct {
		name       string
		entity     any
		field      string
		newValue   any
		wantErr    error
		wantEntity any
	}{
		{
			name: "struct",
			entity: User{
				Name: "Tom",
			},
			field:    "Name",
			newValue: "Jerry",
			wantErr:  errors.New("不可被修改字段"),
		},
		{
			name: "struct",
			entity: &User{
				Name: "Tom",
			},
			field:    "Name",
			newValue: "Jerry",
			wantEntity: &User{
				Name: "Jerry",
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			err := SetField(tc.entity, tc.field, tc.newValue)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantEntity, tc.entity)
		})
	}
}

func TestIterateFunc(t *testing.T) {
	testCase := []struct {
		name    string
		entity  any
		wantRes map[string]FuncInfo
		wantErr error
	}{
		{
			name:   "struct",
			entity: types.NewUser("Tom", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": {
					Name:       "GetAge",
					OutputType: []reflect.Type{reflect.TypeOf(0)},
					Result:     []any{18},
					InputType:  []reflect.Type{reflect.TypeOf(types.User{})},
				},
				//"ChangeName": {
				//	Name:       "ChangeName",
				//	OutputType: []reflect.Type{reflect.TypeOf("")},
				//},
			},
		},
		{
			name:   "pointer",
			entity: types.NewUserPtr("Tom", 18),
			wantRes: map[string]FuncInfo{
				"GetAge": {
					Name:       "GetAge",
					OutputType: []reflect.Type{reflect.TypeOf(0)},
					Result:     []any{18},
					InputType:  []reflect.Type{reflect.TypeOf(&types.User{})},
				},
				"ChangeName": {
					Name:       "ChangeName",
					InputType:  []reflect.Type{reflect.TypeOf(&types.User{}), reflect.TypeOf("")},
					Result:     []any{},
					OutputType: []reflect.Type{},
				},
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			res, err := IterateFunc(tc.entity)
			assert.Equal(t, err, tc.wantErr)
			if err != nil {
				return
			}
			assert.Equal(t, tc.wantRes, res)
		})
	}
}
