package rx

import (
	"net"
	"io"
)


type WrappedConnection struct {
	conn     net.Conn
	sched    Scheduler
	ob       *observer
	context  *Context
	dispose  disposeFunc
	closed   chan struct{}
}
func (w *WrappedConnection) Context() *Context {
	return w.context
}
func (w *WrappedConnection) Scheduler() Scheduler {
	return w.sched
}
func (w *WrappedConnection) closeProperly(err error) {
	select {
	case <- w.closed:
		return
	default:
	}
	close(w.closed)
	_ = w.conn.Close()
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
	var n, err = w.conn.Read(buf)
	if err != nil && err != io.EOF {
		w.closeProperly(err)
		return n, err
	}
	return n, err
}
func (w *WrappedConnection) Write(buf ([] byte)) (int, error) {
	var n, err = w.conn.Write(buf)
	if err != nil {
		w.closeProperly(err)
		return n, err
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

func NewConnectionHandler(conn net.Conn, logic (func(*WrappedConnection))) Action {
	return Action { func(sched Scheduler, ob *observer) {
		var ctx, dispose = ob.context.create_disposable_child()
		var wrapped = &WrappedConnection {
			conn:    conn,
			sched:   sched,
			ob:      ob,
			context: ctx,
			dispose: dispose,
			closed:  make(chan struct{}),
		}
		go logic(wrapped)
	} }
}

