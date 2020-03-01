package rx

import "context"

type Object = interface{}

type Effect struct {
	action  func(Scheduler, *observer)
}

type Scheduler interface {
	dispatch(event)
	run(Effect, *observer)
}

type observer struct {
	context   *Context
	next      func(Object)
	error     func(Object)
	complete  func()
}

type Observer struct {
	raw    *observer
	sched  Scheduler
}

type Context struct {
	raw       context.Context
	disposed  bool
	children  map[*Context] struct{}
}

type Dispose func()

func Background() *Context {
	return &Context {
		raw:      context.Background(),
		children: make(map[*Context] struct{}),
	}
}

func (ctx *Context) CreateChild() (*Context, Dispose) {
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
		if ctx.disposed { return }
		delete(ctx.children, child)
		dispose_recursively(ctx)
		cancel_raw()
	}
}


func (ob Observer) Context() context.Context {
	return ob.raw.context.raw
}

func (ob Observer) Next(x Object) {
	ob.sched.dispatch(event {
		kind:     ev_next,
		payload:  x,
		observer: ob.raw,
	})
}

func (ob Observer) Error(e Object) {
	ob.sched.dispatch(event {
		kind:     ev_error,
		payload:  e,
		observer: ob.raw,
	})
}

func (ob Observer) Complete() {
	ob.sched.dispatch(event {
		kind:     ev_complete,
		payload:  nil,
		observer: ob.raw,
	})
}

func CreateEffect(action func(Observer)) Effect {
	return Effect { func (sched Scheduler, ob *observer) {
		go action(Observer { sched: sched, raw: ob })
	} }
}

func RunEffect (
	e       Effect,
	sched   Scheduler,
	ctx     *Context,
	values  chan <- Object,
	err     chan <- Object,
) {
	sched.run(e, &observer {
		context:  ctx,
		next: func(x Object) {
			if values != nil {
				values <- x
			}
		},
		error: func(e Object) {
			if err != nil {
				err <- e
				close(err)
			}
		},
		complete: func() {
			if values != nil {
				close(values)
			}
		},
	})
}
