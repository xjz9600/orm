package orm

import (
	"orm/internal/errs"
)

type Dialect interface {
	quoter() byte
	buildUpsert(build *builder, upsert *Upsert) error
}

type standardSQL struct {
}

func (s standardSQL) quoter() byte {
	//TODO implement me
	panic("implement me")
}

func (s standardSQL) buildUpsert(build *builder, upsert *Upsert) error {
	//TODO implement me
	panic("implement me")
}

type mysqlDialect struct {
	standardSQL
}

func (s mysqlDialect) buildUpsert(build *builder, upsert *Upsert) error {
	build.sb.WriteString(" ON DUPLICATE KEY UPDATE ")
	for idx, assign := range upsert.assigns {
		if idx > 0 {
			build.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			build.sb.WriteByte('`')
			if _, ok := build.Model.Fields[a.col]; !ok {
				return errs.NewErrUnKnownField(a.col)
			}
			build.sb.WriteString(build.Model.Fields[a.col].ColName)
			build.sb.WriteByte('`')
			build.sb.WriteString("=?")
			build.addArg(a.val)
		case Column:
			if _, ok := build.Model.Fields[a.name]; !ok {
				return errs.NewErrUnKnownField(a.name)
			}
			build.quote(build.Model.Fields[a.name].ColName)
			build.sb.WriteString("=VALUES(")
			build.quote(build.Model.Fields[a.name].ColName)
			build.sb.WriteByte(')')
		default:
			return errs.NewErrUnSupportAssignable(a)
		}
	}
	return nil
}

func (s mysqlDialect) quoter() byte {
	return '`'
}

type sqliteDialect struct {
	standardSQL
}

func (s sqliteDialect) buildUpsert(build *builder, upsert *Upsert) error {
	build.sb.WriteString(" ON CONFLICT(")
	for i, col := range upsert.conflictColumns {
		if i > 0 {
			build.sb.WriteByte(',')
		}
		err := build.buildColumn(Column{name: col})
		if err != nil {
			return err
		}
	}
	build.sb.WriteString(") DO UPDATE SET ")
	for idx, assign := range upsert.assigns {
		if idx > 0 {
			build.sb.WriteByte(',')
		}
		switch a := assign.(type) {
		case Assignment:
			build.sb.WriteByte('`')
			if _, ok := build.Model.Fields[a.col]; !ok {
				return errs.NewErrUnKnownField(a.col)
			}
			build.sb.WriteString(build.Model.Fields[a.col].ColName)
			build.sb.WriteByte('`')
			build.sb.WriteString("=?")
			build.addArg(a.val)
		case Column:
			if _, ok := build.Model.Fields[a.name]; !ok {
				return errs.NewErrUnKnownField(a.name)
			}
			build.quote(build.Model.Fields[a.name].ColName)
			build.sb.WriteString("=exclude.")
			build.quote(build.Model.Fields[a.name].ColName)
		default:
			return errs.NewErrUnSupportAssignable(a)
		}
	}
	return nil
}

func (s sqliteDialect) quoter() byte {
	return '`'
}

type postgreDialect struct {
	standardSQL
}
