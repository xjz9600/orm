package cache

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
	"log"
	"time"
)

var (
	ErrFailedToRefreshCache = errors.New("刷新缓存失败")
)

type ReadThroughCache struct {
	Cache
	LoadFunc   func(ctx context.Context, key string) (any, error)
	Expiration time.Duration
	g          singleflight.Group
}

func (r *ReadThroughCache) Get(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			err = r.Cache.Set(ctx, key, val, r.Expiration)
			if err != nil {
				return val, fmt.Errorf("%w, 原因：%s", ErrFailedToRefreshCache, err.Error())
			}
		}
	}
	return val, err
}

func (r *ReadThroughCache) GetV1(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		go func() {
			val, err = r.LoadFunc(ctx, key)
			if err != nil {
				log.Fatalln(err)
			}
			err = r.Cache.Set(ctx, key, val, r.Expiration)
			if err != nil {
				log.Fatalln(err)
			}
		}()
	}
	return val, err
}

func (r *ReadThroughCache) GetV2(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err = r.LoadFunc(ctx, key)
		if err == nil {
			go func() {
				er := r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					log.Fatalln(er)
				}
			}()
		}
	}
	return val, err
}

func (r *ReadThroughCache) GetV3(ctx context.Context, key string) (any, error) {
	val, err := r.Cache.Get(ctx, key)
	if err == errKeyNotFound {
		val, err, _ = r.g.Do(key, func() (interface{}, error) {
			v, er := r.LoadFunc(ctx, key)
			if er == nil {
				er = r.Cache.Set(ctx, key, val, r.Expiration)
				if er != nil {
					return v, fmt.Errorf("%w, 原因：%s", ErrFailedToRefreshCache, er.Error())
				}
			}
			return v, er
		})
	}
	return val, err
}
