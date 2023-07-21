package rpc

import "context"

type oneWayKey struct{}

func CtxWithOneWay(ctx context.Context) context.Context {
	return context.WithValue(ctx, oneWayKey{}, true)
}

func isOneWay(ctx context.Context) bool {
	val := ctx.Value(oneWayKey{})
	oneWay, ok := val.(bool)
	return ok && oneWay
}
