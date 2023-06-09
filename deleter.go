package orm

type Deleter[T any] struct {
	builder
	where     []Predicate
	tableName string
}

func (s *Deleter[T]) Build() (*Query, error) {
	s.sb.WriteString("DELETE FROM ")
	var (
		t   T
		err error
	)
	s.Model, err = s.r.Get(&t)
	if err != nil {
		return nil, err
	}
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
	s.sb.WriteByte(';')
	return &Query{
		SQL:  s.sb.String(),
		Args: s.args,
	}, nil
}

func (s *Deleter[T]) Where(ps ...Predicate) *Deleter[T] {
	s.where = ps
	return s
}

func (s *Deleter[T]) From(tableName string) *Deleter[T] {
	s.tableName = tableName
	return s
}

func NewDeleter[T any](sess Session) *Deleter[T] {
	core := sess.getCore()
	return &Deleter[T]{
		builder: builder{sess: sess, core: core},
	}
}
