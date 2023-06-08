package orm

import (
	"orm/internal/errs"
	"strings"
)

type builder struct {
	sb   strings.Builder
	args []any
	sess Session
	core
}

func (b *builder) buildGroupBy(ex []Column) error {
	for i := 0; i < len(ex); i++ {
		if i != 0 {
			b.sb.WriteByte(',')
		}
		err := b.buildColumn(ex[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *builder) buildPredicates(ps []Predicate) error {
	p := ps[0]
	for i := 1; i < len(ps); i++ {
		p = p.And(ps[i])
	}
	return b.buildExpression(p)
}

func (b *builder) quote(name string) {
	b.sb.WriteByte(b.dialect.quoter())
	b.sb.WriteString(name)
	b.sb.WriteByte(b.dialect.quoter())
}

func (b *builder) buildColumn(col Column) error {
	switch table := col.table.(type) {
	case nil:
		if _, ok := b.Model.Fields[col.name]; !ok {
			return errs.NewErrUnKnownField(col.name)
		}
		b.quote(b.Model.Fields[col.name].ColName)
		if len(col.alias) != 0 {
			b.sb.WriteString(" AS ")
			b.quote(col.alias)
		}
		return nil
	case Table:
		m, err := b.r.Get(table.entity)
		if err != nil {
			return err
		}
		if _, ok := m.Fields[col.name]; !ok {
			return errs.NewErrUnKnownField(col.name)
		}
		if table.alias != "" {
			b.quote(table.alias)
			b.sb.WriteByte('.')
		}
		b.quote(m.Fields[col.name].ColName)
		if len(col.alias) != 0 {
			b.sb.WriteString(" AS ")
			b.quote(col.alias)
		}
		return nil
	case SubQuery:
		if len(table.columns) > 0 {
			for _, c := range table.columns {
				if c.selectedAlias() == col.name {
					if table.alias != "" {
						b.quote(table.alias)
						b.sb.WriteByte('.')
					}
					b.quote(col.name)
					if len(col.alias) != 0 {
						b.sb.WriteString(" AS ")
						b.quote(col.alias)
					}
					return nil
				}
				if c.fieldName() == col.name {
					if table.alias != "" {
						b.quote(table.alias)
						b.sb.WriteByte('.')
					}
					return b.buildColumn(Column{
						table: table.tbl,
						name:  col.name,
						alias: col.alias,
					})
				}
			}
			return errs.NewErrUnKnownField(col.name)
		}
		if table.alias != "" {
			b.quote(table.alias)
			b.sb.WriteByte('.')
		}
		return b.buildColumn(Column{
			table: table.tbl,
			name:  col.name,
			alias: col.alias,
		})
	default:
		return errs.NewErrUnSupportedTable(table)
	}
}

func (b *builder) buildOrderBy(order OrderBy) error {
	if _, ok := b.Model.Fields[order.col]; !ok {
		return errs.NewErrUnKnownField(order.col)
	}
	b.quote(b.Model.Fields[order.col].ColName)
	b.sb.WriteByte(' ')
	b.sb.WriteString(order.order)
	return nil
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
		if expr.op != "" {
			b.sb.WriteByte(' ')
			b.sb.WriteString(expr.op.String())
			b.sb.WriteByte(' ')
		}
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
		if len(expr.alias) > 0 {
			return errs.ErrAliasWhere
		}
		err := b.buildColumn(expr)
		if err != nil {
			return err
		}
	case value:
		b.sb.WriteByte('?')
		b.addArg(expr.value)
	case RawExpr:
		b.sb.WriteString(expr.raw)
		b.addArg(expr.args...)
	case Aggregate:
		b.sb.WriteString(expr.fn)
		b.sb.WriteByte('(')
		err := b.buildColumn(Column{name: expr.arg})
		if err != nil {
			return err
		}
		b.sb.WriteByte(')')
	case SubQuery:
		b.buildSubQuery(expr)
	case SubqueryExpr:
		b.sb.WriteString(expr.pred)
		b.sb.WriteByte(' ')
		b.buildSubQuery(expr.s)
	default:
		return errs.NewErrUnSupportExpression(expr)
	}
	return nil
}

func (b *builder) addArg(val ...any) {
	if len(val) == 0 {
		return
	}
	if b.args == nil {
		b.args = make([]any, 0, 8)
	}
	b.args = append(b.args, val...)
	return
}

func (b *builder) buildSubQuery(sub SubQuery) error {
	query, err := sub.s.Build()
	if err != nil {
		return err
	}
	b.sb.WriteByte('(')
	b.sb.WriteString(query.SQL[:len(query.SQL)-1])
	b.sb.WriteByte(')')
	if sub.alias != "" {
		b.sb.WriteString(" AS ")
		b.quote(sub.alias)
	}
	b.args = append(b.args, query.Args...)
	return nil
}
