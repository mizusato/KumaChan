package rx


type Object = interface{}

type Action struct {
	action  func(Scheduler, *observer)
}

type Scheduler interface {
	dispatch(event)
	commit(task)
	run(Action, *observer)
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
	hooks      [] func()
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
	if ctx.disposable() && ctx.disposed {
		child.dispose_recursively(behaviour_cancel)
		return child, func(disposeBehaviour) {}
	}
	ctx.children[child] = struct{}{}
	return child, func(behaviour disposeBehaviour) {
		if !(child.disposed) {
			delete(ctx.children, child)
			child.dispose_recursively(behaviour)
			for len(child.hooks) > 0 {
				var l = len(child.hooks)
				if behaviour == behaviour_cancel {
					child.hooks[l-1]()
				}
				child.hooks[l-1] = nil
				child.hooks = child.hooks[:l-1]
			}
		}
	}
}

func (ctx *Context) CancelSignal() (<- chan struct{}) {
	if ctx.disposable() {
		return ctx.cancel
	} else {
		return nil
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

func (ctx *Context) push_cancel_hook(h func()) {
	if !(ctx.disposable()) { return }
	if !(ctx.disposed) {
		ctx.hooks = append(ctx.hooks, h)
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
	Terminate  chan <- bool
}

func (s Sender) Context() *Context {
	return s.ob.context
}

func (s Sender) Scheduler() Scheduler {
	return s.sched
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


func Schedule(action Action, sched Scheduler, r Receiver) {
	sched.commit(func() {
		sched.run(action, &observer {
			context:  r.Context,
			next: func(x Object) {
				if r.Values != nil {
					r.Values <- x
				}
			},
			error: func(e Object) {
				if r.Error != nil {
					r.Error <- e
					close(r.Error)
				}
				if r.Terminate != nil {
					r.Terminate <- false
				}
			},
			complete: func() {
				if r.Values != nil {
					close(r.Values)
				}
				if r.Terminate != nil {
					r.Terminate <- true
				}
			},
		})
	})
}

func ScheduleBackground(action Action, sched Scheduler) {
	Schedule(action, sched, Receiver {
		Context:   Background(),
	})
}

func ScheduleBackgroundWaitTerminate(action Action, sched Scheduler) bool {
	var wait = make(chan bool)
	Schedule(action, sched, Receiver {
		Context:   Background(),
		Terminate: wait,
	})
	return <- wait
}

func Noop() Action {
	return Action { func(sched Scheduler, ob *observer) {
		ob.complete()
	} }
}

func NewGoroutine(action func(Sender)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		go action(Sender { sched: sched, ob: ob })
	} }
}

func NewGoroutineSingle(action func(ctx *Context)(Object,bool)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var sender = Sender { sched: sched, ob: ob }
		go (func() {
			var result, ok = action(sender.Context())
			if ok {
				sender.Next(result)
				sender.Complete()
			} else {
				sender.Error(result)
			}
		})()
	}}
}

func NewQueued(w *Worker, action func()(Object,bool)) Action {
	return Action { func(sched Scheduler, ob *observer) {
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

func NewQueuedNoValue(w *Worker, action func()(bool,Object)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var sender = Sender { sched: sched, ob: ob }
		w.Do(func() {
			var ok, err = action()
			if ok {
				sender.Complete()
			} else {
				sender.Error(err)
			}
		})
	} }
}

func NewCallback(action func(func(Object))) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var sender = Sender { sched: sched, ob: ob }
		action(func(value Object) {
			sender.Next(value)
			sender.Complete()
		})
	}}
}

func NewSubscription(action func(func(Object))(func())) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var h = action(ob.next)
		if h != nil {
			ob.context.push_cancel_hook(h)
		}
	} }
}

func NewSubscriptionWithSender(action func(Sender)(func())) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var h = action(Sender { sched: sched, ob: ob })
		if h != nil {
			ob.context.push_cancel_hook(h)
		}
	} }
}

func NewSync(action func()(Object,bool)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var result, ok = action()
		if ok {
			ob.next(result)
			ob.complete()
		} else {
			ob.error(result)
		}
	} }
}

func NewSyncSequence(action func(func(Object))(bool,Object)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var ok, err = action(ob.next)
		if ok {
			ob.complete()
		} else {
			ob.error(err)
		}
	} }
}

func NewSyncWithSender(action func(Sender)) Action {
	return Action { func(sched Scheduler, ob *observer) {
		action(Sender { sched: sched, ob: ob })
	} }
}

func NewConstant(values... Object) Action {
	return Action { func(sched Scheduler, ob *observer) {
		for _, value := range values {
			ob.next(value)
		}
		ob.complete()
	} }
}

