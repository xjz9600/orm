package orm

import (
	"context"
	"orm/internal/errs"
	"orm/model"
)

type UpsertBuilder[T any] struct {
	i               *Inserter[T]
	conflictColumns []string
}

type Upsert struct {
	assigns         []Assignable
	conflictColumns []string
}

func (i *Inserter[T]) OnDuplicateKey() *UpsertBuilder[T] {
	return &UpsertBuilder[T]{
		i: i,
	}
}

func (o *UpsertBuilder[T]) ConflictColum(cols ...string) *UpsertBuilder[T] {
	o.conflictColumns = cols
	return o
}

func (o *UpsertBuilder[T]) Update(assigns ...Assignable) *Inserter[T] {
	o.i.onDuplicateKey = &Upsert{
		assigns:         assigns,
		conflictColumns: o.conflictColumns,
	}
	return o.i
}

type Assignable interface {
	assign()
}

type Inserter[T any] struct {
	val []*T
	builder
	columns        []string
	onDuplicateKey *Upsert
	tableName      string
}

func NewInserter[T any](sess Session) *Inserter[T] {
	core := sess.getCore()
	return &Inserter[T]{
		builder: builder{sess: sess, core: core},
	}
}

func (i *Inserter[T]) Values(vals ...*T) *Inserter[T] {
	i.val = vals
	return i
}

func (i *Inserter[T]) Columns(col ...string) *Inserter[T] {
	i.columns = col
	return i
}

func (i *Inserter[T]) Build() (*Query, error) {
	if len(i.val) == 0 {
		return nil, errs.ErrInsertZeroRow
	}
	i.sb.WriteString("INSERT INTO ")
	if i.Model == nil {
		var err error
		i.Model, err = i.r.Get(i.val[0])
		if err != nil {
			return nil, err
		}
	}
	if len(i.tableName) != 0 {
		i.sb.WriteString(i.tableName)
	} else {
		i.quote(i.Model.TableName)
	}
	i.sb.WriteByte('(')
	fields := i.Model.FieldArr
	if len(i.columns) > 0 {
		fields = make([]*model.Field, 0, len(i.columns))
		for _, c := range i.columns {
			fd, ok := i.Model.Fields[c]
			if !ok {
				return nil, errs.NewErrUnKnownField(c)
			}
			fields = append(fields, fd)
		}
	}
	for index, f := range fields {
		if index > 0 {
			i.sb.WriteByte(',')
		}
		i.quote(f.ColName)
	}
	i.sb.WriteByte(')')
	i.sb.WriteString(" VALUES ")
	for j, v := range i.val {
		if j > 0 {
			i.sb.WriteByte(',')
		}
		i.sb.WriteByte('(')
		val := i.creator(i.Model, v)
		for index, field := range fields {
			if index > 0 {
				i.sb.WriteByte(',')
			}
			i.sb.WriteByte('?')
			arg, err := val.Field(field.GoName)
			if err != nil {
				return nil, err
			}
			i.addArg(arg)
		}
		i.sb.WriteByte(')')
	}
	if i.onDuplicateKey != nil {
		err := i.dialect.buildUpsert(&i.builder, i.onDuplicateKey)
		if err != nil {
			return nil, err
		}
	}
	i.sb.WriteByte(';')
	return &Query{
		SQL:  i.sb.String(),
		Args: i.args,
	}, nil
}

func (i *Inserter[T]) From(tableName string) *Inserter[T] {
	i.tableName = tableName
	return i
}

func (i *Inserter[T]) Exec(ctx context.Context) Result {
	res := exec[T](ctx, i.sess, i.core, &QueryContext{
		Type:    "INSERT",
		Builder: i,
		Model:   i.Model,
	})
	if res.Result != nil {
		return res.Result.(Result)
	}
	return Result{
		err: res.Err,
	}
}
