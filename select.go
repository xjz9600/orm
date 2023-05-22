package orm

import (
	"context"
)

type Selectable interface {
	selectable()
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
}

func (s *Selector[T]) Get(ctx context.Context) (*T, error) {
	q, err := s.Build()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, ErrNoRows
	}
	tp := new(T)
	val := s.db.creator(s.Model, tp)
	err = val.SetColumns(rows)
	if err != nil {
		return nil, err
	}
	return tp, err
}

func (s *Selector[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}

func NewSelector[T any](db *DB) *Selector[T] {
	return &Selector[T]{
		builder: builder{db: db},
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
	var (
		t   T
		err error
	)
	s.Model, err = s.db.r.Get(&t)
	if err != nil {
		return nil, err
	}
	s.sb.WriteString("SELECT ")
	err = s.buildColumns()
	if err != nil {
		return nil, err
	}
	s.sb.WriteString(" FROM ")
	if len(s.tableName) != 0 {
		s.sb.WriteString(s.tableName)
	} else {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.Model.TableName)
		s.sb.WriteByte('`')
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

func (s *Selector[T]) From(tableName string) *Selector[T] {
	s.tableName = tableName
	return s
}

func (s *Selector[T]) Having(ps ...Predicate) *Selector[T] {
	s.having = ps
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
