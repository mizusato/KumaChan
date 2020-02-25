package rx

import "context"

type Object = interface{}

type Effect struct {
	Action  func(EffectRunner, *Observer)
}

type EffectRunner interface {
	Run(Effect, *Observer)
}

type Observer struct {
	Context   *Context
	Next      func(Object)
	Error     func(Object)
	Complete  func()
}

type Context struct {
	raw       context.Context
	disposed  bool
	children  map[*Context] struct{}
}

func Background() *Context {
	return &Context {
		raw:      context.Background(),
		children: make(map[*Context] struct{}),
	}
}

type Dispose func()

func (ctx *Context) NewChild() (*Context, Dispose) {
	var dispose_recursively func(*Context)
	dispose_recursively = func(ctx *Context) {
		ctx.disposed = true
		for child, _ := range ctx.children {
			dispose_recursively(child)
		}
	}
	var child_raw, cancel_raw = context.WithCancel(ctx.raw)
	var child = &Context {
		raw:      child_raw,
		disposed: false,
		children: make(map[*Context] struct{}),
	}
	ctx.children[child] = struct{}{}
	return child, func() {
		delete(ctx.children, child)
		dispose_recursively(ctx)
		cancel_raw()
	}
}

func (ctx *Context) NewRawChild() (context.Context, context.CancelFunc) {
	return context.WithCancel(ctx.raw)
}
