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

func (r RawExpr) selectedAlias() string {
	return ""
}

func (r RawExpr) fieldName() string {
	return ""
}

func Raw(expr string, args ...any) RawExpr {
	return RawExpr{
		raw:  expr,
		args: args,
	}
}

type SubqueryExpr struct {
	s SubQuery
	// 谓词，ALL，ANY 或者 SOME
	pred string
}

func Any(sub SubQuery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "ANY",
	}
}

func All(sub SubQuery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "ALL",
	}
}

func Some(sub SubQuery) SubqueryExpr {
	return SubqueryExpr{
		s:    sub,
		pred: "SOME",
	}
}

func (r SubqueryExpr) expr() {
	//TODO implement me
	panic("implement me")
}
