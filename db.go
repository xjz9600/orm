package orm

import (
	"database/sql"
	"orm/internal/valuer"
	"orm/model"
)

type DB struct {
	r       model.Registry
	db      *sql.DB
	creator valuer.Creator
	dialect Dialect
}

type DBOptions func(*DB)

func Open(driverName, dataSourceName string, opts ...DBOptions) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return OpenDB(db, opts...)
}

func OpenDB(db *sql.DB, opts ...DBOptions) (*DB, error) {
	res := &DB{
		r:       model.NewRegistry(),
		db:      db,
		creator: valuer.NewUnsafeValue,
		dialect: mysqlDialect{},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res, nil
}

func DBUserReflect() DBOptions {
	return func(db *DB) {
		db.creator = valuer.NewReflectValue
	}
}

func DBWithDialect(dialect Dialect) DBOptions {
	return func(db *DB) {
		db.dialect = dialect
	}
}
func DBWithRegistry(r model.Registry) DBOptions {
	return func(db *DB) {
		db.r = r
	}
}
