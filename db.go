package orm

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"log"
	"orm/internal/errs"
	"orm/internal/valuer"
	"orm/model"
)

type DB struct {
	db *sql.DB
	core
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
		core: core{
			r:       model.NewRegistry(),
			creator: valuer.NewUnsafeValue,
			dialect: mysqlDialect{},
		},
		db: db,
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

func DBWithMiddleware(mdls ...Middleware) DBOptions {
	return func(db *DB) {
		db.mdls = mdls
	}
}

func (db *DB) getCore() core {
	return db.core
}

func (db *DB) DoTx(ctx context.Context, fn func(ctx context.Context, tx *Tx) error, opts *sql.TxOptions) (err error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	panicked := true
	defer func() {
		if panicked || err != nil {
			e := tx.Rollback()
			err = errs.NewErrFailedToRollBackTx(err, e, panicked)
		} else {
			err = tx.Commit()
		}
	}()
	err = fn(ctx, tx)
	panicked = false
	return err
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, db: db}, nil
}

func (db *DB) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.db.QueryContext(ctx, query, args...)
}

func (db *DB) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.db.ExecContext(ctx, query, args...)
}

func (db *DB) Wait() error {
	err := db.db.Ping()
	for err == driver.ErrBadConn {
		log.Println("数据库启动中")
		err = db.db.Ping()
	}
	return err
}
