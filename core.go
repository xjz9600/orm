package orm

import (
	"context"
	"orm/internal/valuer"
	"orm/model"
)

type core struct {
	r       model.Registry
	creator valuer.Creator
	dialect Dialect
	mdls    []Middleware
	Model   *model.Model
}

func get[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var (
		t   T
		err error
	)
	c.Model, err = c.r.Get(&t)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return getHandler[T](ctx, sess, c, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}

func getHandler[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	rows, err := sess.queryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	if !rows.Next() {
		return &QueryResult{
			Err: ErrNoRows,
		}
	}
	tp := new(T)
	val := c.creator(c.Model, tp)
	err = val.SetColumns(rows)
	if err != nil {
		return &QueryResult{
			Err: err,
		}
	}
	return &QueryResult{
		Err:    err,
		Result: tp,
	}
}

func execHandler(ctx context.Context, sess Session, qc *QueryContext) *QueryResult {
	q, err := qc.Builder.Build()
	if err != nil {
		return &QueryResult{
			Result: Result{
				err: err,
			},
		}
	}
	res, err := sess.execContext(ctx, q.SQL, q.Args...)
	return &QueryResult{
		Result: Result{
			err: err,
			res: res,
		},
	}
}

func exec[T any](ctx context.Context, sess Session, c core, qc *QueryContext) *QueryResult {
	var (
		t   T
		err error
	)
	c.Model, err = c.r.Get(&t)
	if err != nil {
		return &QueryResult{
			Result: Result{
				err: err,
			},
		}
	}
	var root Handler = func(ctx context.Context, qc *QueryContext) *QueryResult {
		return execHandler(ctx, sess, qc)
	}
	for i := len(c.mdls) - 1; i >= 0; i-- {
		root = c.mdls[i](root)
	}
	return root(ctx, qc)
}
