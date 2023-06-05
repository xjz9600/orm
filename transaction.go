package orm

import (
	"context"
	"database/sql"
)

type Tx struct {
	tx *sql.Tx
	db *DB
}

var (
	_ Session = &Tx{}
	_ Session = &DB{}
)

func (t *Tx) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) execContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *Tx) Commit() error {
	return t.tx.Commit()
}

func (t *Tx) Rollback() error {
	return t.tx.Rollback()
}

func (t *Tx) getCore() core {
	return t.db.getCore()
}

func (t *Tx) RollBackIfNotCommit() error {
	err := t.tx.Rollback()
	if err == sql.ErrTxDone {
		return nil
	}
	return err
}

type Session interface {
	getCore() core
	queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	execContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
