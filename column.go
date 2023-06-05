package orm

type Column struct {
	name  string
	alias string
	table TableReference
}

func (c Column) assign() {

}

func (c Column) selectable() {

}

func (c Column) expr() {

}

func (c Column) As(alias string) Column {
	return Column{
		name:  c.name,
		alias: alias,
		table: c.table,
	}
}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) EQ(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(val),
	}
}

func (c Column) LT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opLT,
		right: valueOf(val),
	}
}

func (c Column) GT(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opGT,
		right: valueOf(val),
	}
}
