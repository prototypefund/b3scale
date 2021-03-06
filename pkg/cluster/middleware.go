package cluster

import (
	"context"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// Schema is a mapping of variable names and decode hints
type Schema map[string]string

// RequestHandler accepts a bbb request and state. It produces
// a bbb response or an error.
type RequestHandler func(context.Context, *bbb.Request) (bbb.Response, error)

// RequestMiddleware is a plain middleware without a state
type RequestMiddleware func(next RequestHandler) RequestHandler

// RouterHandler accepts a bbb request and a list of
// backends and returns a filtered or sorted
// list of backends.
type RouterHandler func(context.Context, []*Backend, *bbb.Request) ([]*Backend, error)

// A RouterMiddleware accepts a handler function
// and returns a decorated handler function.
type RouterMiddleware func(next RouterHandler) RouterHandler
