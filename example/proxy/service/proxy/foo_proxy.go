// Code generated by simple-gen-proxy. DO NOT EDIT.
// source: example/proxy/service/foo.go

package proxy

import (
	"context"

	service "github.com/ISSuh/simple-gen-proxy/example/proxy/service"
)

type FooProxyMiddleware func(func(context.Context) error) func(context.Context) error

type FooProxy struct {
	target      service.Foo
	middlewares []FooProxyMiddleware
}

func NewFooProxy(target service.Foo, middlewares ...FooProxyMiddleware) *FooProxy {
	return &FooProxy{
		target:      target,
		middlewares: middlewares,
	}
}

func (p *FooProxy) Logic(needEmitErr bool) (string, error) {
	var (
		r0  string
		err error
	)

	f := func(context.Context) error {
		r0, err = p.target.Logic(needEmitErr)
		if err != nil {
			return err
		}
		return nil
	}

	for i := range p.middlewares {
		index := len(p.middlewares) - i - 1
		f = p.middlewares[index](f)
	}

	f(context.TODO())
	return r0, err
}
