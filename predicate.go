package orm

type op string

const (
	opEq    op = "="
	opNot   op = "NOT"
	opAnd   op = "AND"
	opOr    op = "OR"
	opLT    op = "<"
	opGT    op = ">"
	opIN    op = "IN"
	opExist op = "EXIST"
)

func (o op) String() string {
	return string(o)
}

type Predicate struct {
	left  Expression
	op    op
	right Expression
}

func (left Predicate) expr() {

}

func Not(p Predicate) Predicate {
	return Predicate{
		op:    opNot,
		right: p,
	}
}

func Exist(sub SubQuery) Predicate {
	return Predicate{
		op:    opExist,
		right: sub,
	}
}

func (left Predicate) And(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opAnd,
		right: right,
	}
}

func (left Predicate) Or(right Predicate) Predicate {
	return Predicate{
		left:  left,
		op:    opOr,
		right: right,
	}
}

type value struct {
	value any
}

func (v value) expr() {
}

func valueOf(val any) Expression {
	switch val.(type) {
	case Expression:
		return val.(Expression)
	default:
		return value{value: val}
	}
}
