package orm

type Column struct {
	name  string
	alias string
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
	}
}

func C(name string) Column {
	return Column{name: name}
}

func (c Column) Eq(val any) Predicate {
	return Predicate{
		left:  c,
		op:    opEq,
		right: valueOf(val),
	}
}
