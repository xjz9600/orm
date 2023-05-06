package orm

import (
	"orm/internal/errs"
	"strings"
)

type builder struct {
	sb        strings.Builder
	args      []any
	model     *model
	tableName string
}

func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p.And(ps[i])
	}
	return b.buildExpression(p)
}

func (b *builder) buildExpression(exp Expression) error {
	switch expr := exp.(type) {
	case nil:
	case Predicate:
		_, ok := expr.left.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(expr.left); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
		b.sb.WriteByte(' ')
		b.sb.WriteString(expr.op.String())
		b.sb.WriteByte(' ')
		_, ok = expr.right.(Predicate)
		if ok {
			b.sb.WriteByte('(')
		}
		if err := b.buildExpression(expr.right); err != nil {
			return err
		}
		if ok {
			b.sb.WriteByte(')')
		}
	case Column:
		b.sb.WriteByte('`')
		if _, ok := b.model.fields[expr.name]; !ok {
			return errs.NewErrUnKnownField(expr.name)
		}
		b.sb.WriteString(b.model.fields[expr.name].colName)
		b.sb.WriteByte('`')
	case value:
		b.sb.WriteByte('?')
		b.addArg(expr.value)
	default:
		return errs.NewErrUnSupportExpression(expr)
	}
	return nil
}

func (b *builder) addArg(val any) *builder {
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, val)
	return b
}
