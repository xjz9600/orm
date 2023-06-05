package middleware

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"orm"
	"time"
)

type prometheusBuilder struct {
	name      string
	subSystem string
	nameSpace string
	help      string
}

func NewPrometheusBuilder(name, subSystem, nameSpace, help string) *prometheusBuilder {
	return &prometheusBuilder{name, subSystem, nameSpace, help}
}

func (p *prometheusBuilder) Build() orm.Middleware {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:      p.name,
		Subsystem: p.subSystem,
		Namespace: p.nameSpace,
		Help:      p.help,
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})
	prometheus.MustRegister(vector)
	return func(next orm.Handler) orm.Handler {
		return func(ctx context.Context, qc *orm.QueryContext) *orm.QueryResult {
			startTime := time.Now()
			defer func() {
				duration := time.Now().Sub(startTime).Milliseconds()
				vector.WithLabelValues(qc.Type, qc.Model.TableName).Observe(float64(duration))
			}()
			return next(ctx, qc)
		}
	}
}
