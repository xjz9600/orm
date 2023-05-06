package orm

type Selector[T any] struct {
	where []Predicate
	builder
}

func (s *Selector[T]) Build() (*Query, error) {
	s.sb.WriteString("SELECT * FROM ")
	var (
		t   T
		err error
	)
	s.model, err = parseModel(&t)
	if err != nil {
		return nil, err
	}
	if len(s.tableName) != 0 {
		s.sb.WriteString(s.tableName)
	} else {
		s.sb.WriteByte('`')
		s.sb.WriteString(s.model.tableName)
		s.sb.WriteByte('`')
	}
	if len(s.where) > 0 {
		s.sb.WriteString(" WHERE ")
		err = s.buildPredicates(s.where)
		if err != nil {
			return nil, err
		}
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

func (s *Selector[T]) From(tableName string) *Selector[T] {
	s.tableName = tableName
	return s
}
