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
}
func (w *WrappedConnection) Read(buf ([] byte)) (int, error) {
	err := w.conn.SetReadDeadline(time.Now().Add(w.timeout.ReadTimeout))
	if err != nil { return 0, err }
	n, err := w.conn.Read(buf)
	if err != nil {
		w.closeProperly(err)
	}
	return n, err
}
func (w *WrappedConnection) Write(buf ([] byte)) (int, error) {
	err := w.conn.SetWriteDeadline(time.Now().Add(w.timeout.WriteTimeout))
	if err != nil { return 0, err }
	n, err := w.conn.Write(buf)
	if err != nil {
		w.closeProperly(err)
	}
	return n, err
}
func (w *WrappedConnection) Close() error {
	w.closeProperly(nil)
	return nil
}
func (w *WrappedConnection) Fatal(err error) {
	w.closeProperly(err)
}

func NewConnectionHandler(conn net.Conn, timeout TimeoutPair, logic (func(*WrappedConnection))) Action {
	return Action { func(sched Scheduler, ob *observer) {
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
		}
		go logic(wrapped)
		go ob.context.WaitDispose(func() {
			_ = wrapped.Close()
		})
	} }
}

