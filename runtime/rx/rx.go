package rx


type Object = interface{}

type Effect struct {
	action  func(Scheduler, *observer)
}

type Scheduler interface {
	dispatch(event)
	commit(task)
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
	children   map[*Context] struct{}
	disposed   bool
	cancel     chan struct{}
	terminate  chan struct{}
}

type disposeFunc func(disposeBehaviour)
type disposeBehaviour int
const (
	behaviour_cancel disposeBehaviour = iota
	behaviour_terminate
)

var background = &Context {
	children:  make(map[*Context] struct{}),
	disposed:  false,
	cancel:    nil,
	terminate: nil,
}

func Background() *Context {
	return background
}

func (ctx *Context) disposable() bool {
	return (ctx != background)
}

func (ctx *Context) dispose_recursively(behaviour disposeBehaviour) {
	if !(ctx.disposable()) { panic("something went wrong") }
	if !(ctx.disposed) {
		ctx.disposed = true
		for child, _ := range ctx.children {
			child.dispose_recursively(behaviour)
		}
		switch behaviour {
		case behaviour_cancel:
			close(ctx.cancel)
		case behaviour_terminate:
			close(ctx.terminate)
		default:
			panic("impossible branch")
		}
	}
}

func (ctx *Context) create_disposable_child() (*Context, disposeFunc) {
	var child = &Context {
		children:  make(map[*Context] struct{}),
		disposed:  false,
		cancel:    make(chan struct{}),
		terminate: make(chan struct{}),
	}
	ctx.children[child] = struct{}{}
	return child, func(behaviour disposeBehaviour) {
		if !(child.disposed) {
			delete(ctx.children, child)
			child.dispose_recursively(behaviour)
		}
	}
}

func (ctx *Context) AlreadyCancelled() bool {
	if ctx.disposable() {
		select {
		case <- ctx.cancel:
			return true
		default:
			return false
		}
	} else {
		return false
	}
}

func (ctx *Context) WaitDispose(handleCancel func()) {
	if ctx.disposable() {
		select {
		case <- ctx.cancel:
			handleCancel()
		case <- ctx.terminate:
			// do nothing
		}
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

func (s Sender) Context() *Context {
	return s.ob.context
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

func CreateValueCallbackEffect(action func(func(Object))) Effect {
	return Effect { func(sched Scheduler, ob *observer) {
		var sender = Sender { sched: sched, ob: ob }
		action(func(value Object) {
			sender.Next(value)
			sender.Complete()
		})
	}}
}

