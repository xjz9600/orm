package orm

import (
	"context"
	"orm/model"
)

type QueryContext struct {
	Type    string
	Builder QueryBuilder
	Model   *model.Model
}

type QueryResult struct {
	Result any
	Err    error
}

type Handler func(ctx context.Context, qc *QueryContext) *QueryResult

type Middleware func(next Handler) Handler
