package model

import (
	"database/sql"
	"errors"
	"github.com/stretchr/testify/assert"
	"orm/internal/errs"
	"reflect"
	"testing"
)

func TestRegistry_get(t *testing.T) {
	testCases := []struct {
		name      string
		val       any
		wantModel *Model
		wantErr   error
	}{
		{
			// 指针
			name: "pointer",
			val:  &TestModel{},
			wantModel: &Model{
				TableName: "test_model",
				FieldArr: []*Field{
					{
						ColName: "id",
						GoName:  "Id",
						Typ:     reflect.TypeOf(int64(0)),
						Offset:  0,
					},
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(""),
						Offset:  8,
					},
					{
						ColName: "age",
						GoName:  "Age",
						Typ:     reflect.TypeOf(int8(0)),
						Offset:  24,
					},
					{
						ColName: "last_name",
						GoName:  "LastName",
						Typ:     reflect.TypeOf(sql.NullString{}),
						Offset:  32,
					},
				},
			},
		},
		{
			name:    "map",
			val:     map[string]string{},
			wantErr: errors.New("Model: 不支持类型 map"),
		},
		{
			name:    "slice",
			val:     []int{},
			wantErr: errors.New("Model: 不支持类型 slice"),
		},
		{
			name:    "basic type",
			val:     0,
			wantErr: errors.New("Model: 不支持类型 int"),
		},

		// 标签相关测试用例
		{
			name: "column tag",
			val: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type ColumnTag struct {
					ID uint64 `orm:"column=id"`
				}
				return &ColumnTag{}
			}(),
			wantModel: &Model{
				TableName: "column_tag",
				FieldArr: []*Field{
					{
						ColName: "id",
						GoName:  "ID",
						Typ:     reflect.TypeOf(uint64(0)),
					},
				},
			},
		},
		{
			// 如果用户设置了 column，但是传入一个空字符串，那么会用默认的名字
			name: "empty column",
			val: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type EmptyColumn struct {
					FirstName uint64 `orm:"column="`
				}
				return &EmptyColumn{}
			}(),
			wantModel: &Model{
				TableName: "empty_column",
				FieldArr: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(uint64(0)),
					},
				},
			},
		},
		{
			// 如果用户设置了 column，但是没有赋值
			name: "invalid tag",
			val: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type InvalidTag struct {
					FirstName uint64 `orm:"column"`
				}
				return &InvalidTag{}
			}(),
			wantErr: errs.NewErrInvalidTagContent("column"),
		},
		{
			// 如果用户设置了一些奇奇怪怪的内容，这部分内容我们会忽略掉
			name: "ignore tag",
			val: func() any {
				// 我们把测试结构体定义在方法内部，防止被其它用例访问
				type IgnoreTag struct {
					FirstName uint64 `orm:"abc=abc"`
				}
				return &IgnoreTag{}
			}(),
			wantModel: &Model{
				TableName: "ignore_tag",
				FieldArr: []*Field{
					{
						ColName: "first_name",
						GoName:  "FirstName",
						Typ:     reflect.TypeOf(uint64(0)),
					},
				},
			},
		},

		// 利用接口自定义模型信息
		{
			name: "table name",
			val:  &CustomTableName{},
			wantModel: &Model{
				TableName: "custom_table_name_t",
				FieldArr: []*Field{
					{
						ColName: "name",
						GoName:  "Name",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "table name ptr",
			val:  &CustomTableNamePtr{},
			wantModel: &Model{
				TableName: "custom_table_name_ptr_t",
				FieldArr: []*Field{
					{
						ColName: "name",
						GoName:  "Name",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
		{
			name: "empty table name",
			val:  &EmptyTableName{},
			wantModel: &Model{
				TableName: "empty_table_name",
				FieldArr: []*Field{
					{
						ColName: "name",
						GoName:  "Name",
						Typ:     reflect.TypeOf(""),
					},
				},
			},
		},
	}

	r := &registry{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := r.Get(tc.val)
			assert.Equal(t, tc.wantErr, err)
			if err != nil {
				return
			}
			fieldMap := make(map[string]*Field, len(m.Fields))
			columMap := make(map[string]*Field, len(m.Fields))
			for _, f := range tc.wantModel.FieldArr {
				fieldMap[f.GoName] = f
				columMap[f.ColName] = f
			}
			tc.wantModel.Fields = fieldMap
			tc.wantModel.Columns = columMap
			assert.Equal(t, tc.wantModel, m)
		})
	}
}

func Test_underscoreName(t *testing.T) {
	testCases := []struct {
		name    string
		srcStr  string
		wantStr string
	}{
		// 我们这些用例就是为了确保
		// 在忘记 underscoreName 的行为特性之后
		// 可以从这里找回来
		// 比如说过了一段时间之后
		// 忘记了 ID 不能转化为 id
		// 那么这个测试能帮我们确定 ID 只能转化为 i_d
		{
			name:    "upper cases",
			srcStr:  "ID",
			wantStr: "i_d",
		},
		{
			name:    "use number",
			srcStr:  "Table1Name",
			wantStr: "table1_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res := underscoreName(tc.srcStr)
			assert.Equal(t, tc.wantStr, res)
		})
	}
}

type CustomTableName struct {
	Name string
}

func (c CustomTableName) TableName() string {
	return "custom_table_name_t"
}

type CustomTableNamePtr struct {
	Name string
}

func (c *CustomTableNamePtr) TableName() string {
	return "custom_table_name_ptr_t"
}

type EmptyTableName struct {
	Name string
}

func (c *EmptyTableName) TableName() string {
	return ""
}
