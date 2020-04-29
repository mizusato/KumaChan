package rx

import (
	"context"
)


type Object = interface{}

type Effect struct {
	action  func(Scheduler, *observer)
}

type Scheduler interface {
	dispatch(event)
	run(Effect, *observer)
	RunTopLevel(Effect, Receiver)
}

type observer struct {
	context   *Context
	next      func(Object)
	error     func(Object)
	complete  func()
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
		if child.disposed { return }
		delete(ctx.children, child)
		dispose_recursively(child)
		cancel_raw()
	}
}


type Sender struct {
	ob     *observer
	sched  Scheduler
}

type Receiver struct {
	Context    *Context
	Values     chan <- Object
	Error      chan <- Object
	Terminate  chan <- struct {}
}

func (s Sender) Context() context.Context {
	return s.ob.context.raw
}

func (s Sender) CancelSignal() (<- chan struct{}, bool) {
	var done = s.ob.context.raw.Done()
	if done != nil {
		return done, true
	} else {
		return nil, false
	}
}

func (s Sender) Next(x Object) {
	s.sched.dispatch(event {
		kind:     ev_next,
		payload:  x,
		observer: s.ob,
	})
}

func (s Sender) Error(e Object) {
	s.sched.dispatch(event {
		kind:     ev_error,
		payload:  e,
		observer: s.ob,
	})
}

func (s Sender) Complete() {
	s.sched.dispatch(event {
		kind:     ev_complete,
		payload:  nil,
		observer: s.ob,
	})
}

func CreateEffect(action func(Sender)) Effect {
	return Effect { func (sched Scheduler, ob *observer) {
		go action(Sender { sched: sched, ob: ob })
	} }
}

func CreateQueuedEffect(w *Worker, action func()(Object,bool)) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var sender = Sender { sched: sched, ob: ob }
		w.Do(func() {
			var result, ok = action()
			if ok {
				sender.Next(result)
				sender.Complete()
			} else {
				sender.Error(result)
			}
		})
	} }
}

func CreateBlockingEffect(action func()(Object,bool)) Effect {
	return Effect { func (sched Scheduler, ob *observer) {
		var result, ok = action()
		if ok {
			ob.next(result)
			ob.complete()
		} else {
			ob.error(result)
		}
	} }
}

func CreateBlockingSequenceEffect(action func(func(Object))(bool,Object)) Effect {
	return Effect { func (sched Scheduler, ob *observer) {
		var ok, err = action(ob.next)
		if ok {
			ob.complete()
		} else {
			ob.error(err)
		}
	} }
}
