package orm

type Expression interface {
	expr()
}

// RawExpr 代表原生表达式
type RawExpr struct {
	raw  string
	args []any
}

func (r RawExpr) expr() {
	//TODO implement me
	panic("implement me")
}

func (r RawExpr) AsPredicate() Predicate {
	return Predicate{
		left: r,
	}
}

func (r RawExpr) selectable() {

}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}
