package rx

import (
	"net"
	"time"
)


type WrappedConnection struct {
	conn     net.Conn
	timeout  TimeoutPair
	sched    Scheduler
	ob       *observer
	worker   *Worker
	context  *Context
	dispose  disposeFunc
	closed   chan struct{}
	result   Promise
}
type TimeoutPair struct {
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
}
func (w *WrappedConnection) Context() *Context {
	return w.context
}
func (w *WrappedConnection) Scheduler() Scheduler {
	return w.sched
}
func (w *WrappedConnection) Worker() *Worker {
	return w.worker
}
func (w *WrappedConnection) closeProperly(err error) {
	select {
	case <- w.closed:
		return
	default:
	}
	close(w.closed)
	_ = w.conn.Close()
	w.worker.Dispose()
	w.sched.commit(func() {
		if err != nil {
			w.ob.error(err)
		} else {
			w.ob.complete()
		}
		w.dispose(behaviour_cancel)
	})
	if err != nil {
		w.result.Reject(err, w.sched)
	} else {
		w.result.Resolve(nil, w.sched)
	}
}
func (w *WrappedConnection) Read(buf ([] byte)) (int, error) {
	var timeout = w.timeout.ReadTimeout
	if timeout != 0 {
		err := w.conn.SetReadDeadline(time.Now().Add(timeout))
		if err != nil { return 0, err }
	}
	n, err := w.conn.Read(buf)
	if err != nil {
		w.closeProperly(err)
	}
	return n, err
}
func (w *WrappedConnection) Write(buf ([] byte)) (int, error) {
	var timeout = w.timeout.WriteTimeout
	if timeout != 0 {
		err := w.conn.SetWriteDeadline(time.Now().Add(timeout))
		if err != nil { return 0, err }
	}
	n, err := w.conn.Write(buf)
	if err != nil {
		w.closeProperly(err)
	}
	return n, err
}
func (w *WrappedConnection) Fatal(err error) {
	w.closeProperly(err)
}
func (w *WrappedConnection) Close() error {
	w.closeProperly(nil)
	return nil
}
func (w *WrappedConnection) OnClose() Observable {
	return w.result.Outcome()
}

func NewConnectionHandler(conn net.Conn, timeout TimeoutPair, logic (func(*WrappedConnection))) Observable {
	return Observable { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var wrapped = &WrappedConnection {
			conn:    conn,
			timeout: timeout,
			sched:   sched,
			ob:      ob,
			worker:  CreateWorker(),
			context: ctx,
			dispose: dispose,
			closed:  make(chan struct{}),
			result:  CreatePromise(),
		}
		go logic(wrapped)
		go ob.context.WaitDispose(func() {
			_ = wrapped.Close()
		})
	} }
}

