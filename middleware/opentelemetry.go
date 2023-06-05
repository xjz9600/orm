package middleware

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"orm"
)

const instrumentationName = "github.com/xjz9600/orm/middleware/tracing"

type opentelemetry struct {
	tracer trace.Tracer
}

func (m opentelemetry) Build() orm.Middleware {
	if m.tracer == nil {
		m.tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			tbl := qc.Model.TableName
			spanCtx, span := m.tracer.Start(ctx, fmt.Sprintf("%s-%s", qc.Type, tbl))
			defer span.End()
			q, _ := qc.Builder.Build()
			if q != nil {
				span.SetAttributes(attribute.String("sql", q.SQL))
			}
			span.SetAttributes(attribute.String("component", "orm"))
			res := next(spanCtx, qc)
			if res.Err != nil {
				span.RecordError(res.Err)
			}
			return res
		}
	}
}
