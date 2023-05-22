package model

import (
	"orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

type Registry interface {
	Get(val any) (*Model, error)
	Register(val any, opts ...ModelOpt) (*Model, error)
}

type registry struct {
	Models sync.Map
}

func (r *registry) Get(val any) (*Model, error) {
	typ := reflect.TypeOf(val)
	m, ok := r.Models.Load(typ)
	if ok {
		return m.(*Model), nil
	}
	return r.Register(val)
}

func (r *registry) Register(val any, opts ...ModelOpt) (*Model, error) {
	m, err := r.parseModel(val)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if err := opt(m); err != nil {
			return nil, err
		}
	}
	r.Models.Store(reflect.TypeOf(val), m)
	return m, nil
}

func NewRegistry() *registry {
	return &registry{}
}

func (r *registry) parseModel(entity any) (*Model, error) {
	typ := reflect.TypeOf(entity)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errs.NewErrUnSupportType(typ.Kind())
	}
	numField := typ.NumField()
	fields := make(map[string]*Field, numField)
	columns := make(map[string]*Field, numField)
	fieldArr := make([]*Field, 0, numField)
	for i := 0; i < numField; i++ {
		fieldType := typ.Field(i)
		tags, err := r.parseTag(fieldType.Tag)
		if err != nil {
			return nil, err
		}
		columnName := tags[tagKeyColumn]
		if columnName == "" {
			columnName = underscoreName(fieldType.Name)
		}
		fi := &Field{ColName: columnName, Typ: fieldType.Type, GoName: fieldType.Name, Offset: fieldType.Offset}
		fields[fieldType.Name] = fi
		columns[columnName] = fi
		fieldArr = append(fieldArr, fi)
	}
	var tableName string
	if tn, ok := entity.(TableName); ok {
		tableName = tn.TableName()
	}
	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}
	return &Model{
		TableName: tableName,
		Fields:    fields,
		Columns:   columns,
		FieldArr:  fieldArr,
	}, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		return map[string]string{}, nil
	}
	res := make(map[string]string, 1)
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		res[kv[0]] = kv[1]
	}
	return res, nil
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
