package orm

type Aggregate struct {
	fn    string
	arg   string
	alias string
}

func (a Aggregate) selectedAlias() string {
	return a.alias
}

func (a Aggregate) fieldName() string {
	return a.arg
}

func (a Aggregate) As(alias string) Aggregate {
	return Aggregate{
		fn:    a.fn,
		arg:   a.arg,
		alias: alias,
	}
}

func (a Aggregate) expr() {
}

func (a Aggregate) EQ(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opEq,
		right: valueOf(arg),
	}
}

func (a Aggregate) LT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opLT,
		right: valueOf(arg),
	}
}

func (a Aggregate) GT(arg any) Predicate {
	return Predicate{
		left:  a,
		op:    opGT,
		right: valueOf(arg),
	}
}

func Avg(col string) Aggregate {
	return Aggregate{
		fn:  "AVG",
		arg: col,
	}
}

func Max(col string) Aggregate {
	return Aggregate{
		fn:  "MAX",
		arg: col,
	}
}

func Count(col string) Aggregate {
	return Aggregate{
		fn:  "COUNT",
		arg: col,
	}
}
func Min(col string) Aggregate {
	return Aggregate{
		fn:  "MIN",
		arg: col,
	}
}
