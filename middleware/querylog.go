package middleware

import (
	"context"
	"log"
	"orm"
)

type queryLog struct {
	logFunc func(query string, args []any)
}

func NewQueryLog() *queryLog {
	return &queryLog{
		logFunc: func(query string, args []any) {
			log.Printf("sql：%s, args：%v", query, args)
		},
	}
}

func (m *queryLog) LogFunc(fn func(query string, args []any)) *queryLog {
	m.logFunc = fn
	return m
}

func (m queryLog) Build() orm.Middleware {
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			q, err := qc.Builder.Build()
			if err != nil {
				return &orm.QueryResult{
					Err: err,
				}
			}
			m.logFunc(q.SQL, q.Args)
			res := next(ctx, qc)
			return res
		}
	}
}
