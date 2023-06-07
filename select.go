package orm

import (
	"context"
	"orm/internal/errs"
)

type Selectable interface {
	selectedAlias() string
	fieldName() string
}

type Selector[T any] struct {
	where []Predicate
	builder
	columns []Selectable
	groupBy []Column
	having  []Predicate
	orderBy []OrderBy
	offset  int
	limit   int
	table   TableReference
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	res := get[T](ctx, s.sess, s.core, &QueryContext{
		Type:    "SELECT",
		Builder: s,
		Model:   s.Model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func NewSelector[T any](sess Session) *Selector[T] {
	core := sess.getCore()
	return &Selector[T]{
		builder: builder{sess: sess, core: core},
	}
}

func (s *Selector[T]) buildColumns() error {
	if len(s.columns) == 0 {
		s.sb.WriteByte('*')
		return nil
	}
	for i, col := range s.columns {
		if i > 0 {
			s.sb.WriteByte(',')
		}
		switch expr := col.(type) {
		case Column:
			err := s.buildColumn(expr)
			if err != nil {
				return err
			}
		case Aggregate:
			s.sb.WriteString(expr.fn)
			s.sb.WriteByte('(')
			err := s.buildColumn(Column{name: expr.arg})
			if err != nil {
				return err
			}
			s.sb.WriteByte(')')
			if len(expr.alias) != 0 {
				s.sb.WriteString(" AS `")
				s.sb.WriteString(expr.alias)
				s.sb.WriteByte('`')
			}
		case RawExpr:
			s.sb.WriteString(expr.raw)
			s.addArg(expr.args...)
		}
	}
	return nil
}

func (s *Selector[T]) Build() (*Query, error) {
	if s.Model == nil {
		var (
			t   T
			err error
		)
		s.Model, err = s.r.Get(&t)
		if err != nil {
			return nil, err
		}
	}
	s.sb.Reset()
	s.sb.WriteString("SELECT ")
	err := s.buildColumns()
	if err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if err := s.buildTable(s.table); err != nil {
		return nil, err
	}
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		err = s.buildPredicates(s.where)
		if err != nil {
			return nil, err
		}
	}
	if len(s.groupBy) > 0 {
		s.sb.WriteString(" GROUP BY ")
		err = s.buildGroupBy(s.groupBy)
		if err != nil {
			return nil, err
		}
	}
	if len(s.having) > 0 {
		s.sb.WriteString(" HAVING ")
		err = s.buildPredicates(s.having)
		if err != nil {
			return nil, err
		}
	}
	if len(s.orderBy) > 0 {
		s.sb.WriteString(" ORDER BY ")
		for i, o := range s.orderBy {
			if i != 0 {
				s.sb.WriteByte(',')
			}
			err := s.buildOrderBy(o)
			if err != nil {
				return nil, err
			}
		}
	}
	if s.limit > 0 {
		s.sb.WriteString(" LIMIT ?")
		s.addArg(s.limit)
	}

	if s.offset > 0 {
		s.sb.WriteString(" OFFSET ?")
		s.addArg(s.offset)
	}
	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Selector[T]) buildTable(table TableReference) error {
	switch t := table.(type) {
	case nil:
		s.quote(s.Model.TableName)
	case Table:
		m, err := s.r.Get(t.entity)
		if err != nil {
			return err
		}
		s.quote(m.TableName)
		if t.alias != "" {
			s.sb.WriteString(" AS ")
			s.quote(t.alias)
		}
	case Join:
		s.sb.WriteByte('(')
		err := s.buildTable(t.left)
		if err != nil {
			return err
		}
		s.sb.WriteByte(' ')
		s.sb.WriteString(t.typ)
		s.sb.WriteByte(' ')
		err = s.buildTable(t.right)
		if err != nil {
			return err
		}
		if len(t.using) > 0 {
			s.sb.WriteString(" USING (")
			for i, u := range t.using {
				if i > 0 {
					s.sb.WriteByte(',')
				}
				err = s.buildColumn(Column{name: u})
				if err != nil {
					return err
				}
			}
			s.sb.WriteByte(')')
		}
		if len(t.on) > 0 {
			s.sb.WriteString(" ON ")
			err = s.buildPredicates(t.on)
			if err != nil {
				return err
			}
		}
		s.sb.WriteByte(')')
	case SubQuery:
		return s.buildSubQuery(t)
	default:
		return errs.NewErrUnSupportedTable(t)
	}
	return nil
}

func (s *Selector[T]) Where(ps ...Predicate) *Selector[T] {
	s.where = ps
	return s
}

func (s *Selector[T]) GroupBy(cols ...Column) *Selector[T] {
	s.groupBy = cols
	return s
}

func (s *Selector[T]) OrderBy(order ...OrderBy) *Selector[T] {
	s.orderBy = order
	return s
}

func (s *Selector[T]) Offset(offset int) *Selector[T] {
	s.offset = offset
	return s
}

func (s *Selector[T]) Limit(limit int) *Selector[T] {
	s.limit = limit
	return s
}

func (s *Selector[T]) Select(cols ...Selectable) *Selector[T] {
	s.columns = cols
	return s
}

func (s *Selector[T]) From(table TableReference) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
	return s
}

func (s *Selector[T]) AsSubQuery() SubQuery {
	var tbl = s.table
	if tbl == nil {
		tbl = TableOf(new(T))
	}
	return SubQuery{
		s:       s,
		tbl:     tbl,
		columns: s.columns,
	}
}

func (s SubQuery) As(alias string) SubQuery {
	s.alias = alias
	return s
}

type OrderBy struct {
	col   string
	order string
}

func Asc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: "ASC",
	}
}

func Desc(col string) OrderBy {
	return OrderBy{
		col:   col,
		order: "DESC",
	}
}
