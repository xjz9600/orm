package orm

import (
	"context"
	"database/sql"
)

type RawQuerier[T any] struct {
	sql  string
	args []any
	core
	sess Session
}

func (r *RawQuerier[T]) Build() (*Query, error) {
	return &Query{
		SQL:  r.sql,
		Args: r.args,
	}, nil
}

func RawQuery[T any](sess Session, query string, args ...any) *RawQuerier[T] {
	core := sess.getCore()
	return &RawQuerier[T]{
		sql:  query,
		args: args,
		sess: sess,
		core: core,
	}
}

func (r *RawQuerier[T]) Exec(ctx context.Context) sql.Result {
	res := exec[T](ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.Model,
	})
	if res.Result != nil {
		return res.Result.(Result)
	}
	return Result{
		err: res.Err,
	}
}

func (r *RawQuerier[T]) Get(ctx context.Context) (*T, error) {
	res := get[T](ctx, r.sess, r.core, &QueryContext{
		Type:    "RAW",
		Builder: r,
		Model:   r.Model,
	})
	if res.Result != nil {
		return res.Result.(*T), res.Err
	}
	return nil, res.Err
}

func (r *RawQuerier[T]) GetMulti(ctx context.Context) ([]*T, error) {
	//TODO implement me
	panic("implement me")
}
