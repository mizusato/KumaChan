package rpc

import (
	"io"
	"fmt"
	"net"
	"time"
	"errors"
	"kumachan/rx"
	"kumachan/rpc/kmd"
)


type ServerOptions struct {
	Listener     net.Listener
	DebugOutput  io.Writer
	Limits
	KmdApi
}

func Server(service Service, opts *ServerOptions) rx.Action {
	return rx.NewGoroutine(func(sender rx.Sender) {
		var l = opts.Listener
		go sender.Context().WaitDispose(func() {
			_ = l.Close()
		})
		for {
			var conn, err = l.Accept()
			if err != nil {
				_ = l.Close()
				sender.Error(err)
				return
			}
			sender.Next(conn)
		}
	}).MergeMap(func(raw_conn_ rx.Object) rx.Action {
		var raw_conn = raw_conn_.(net.Conn)
		var logger = &ServerLogger {
			LocalAddr:  raw_conn.LocalAddr(),
			RemoteAddr: raw_conn.RemoteAddr(),
			Output:     opts.DebugOutput,
		}
		var handle = func(conn *rx.WrappedConnection) struct{} {
			var fatal = func(err error) struct{} {
				conn.Fatal(err)
				return struct{}{}
			}
			client_info, err := receiveServiceConfirmation(conn)
			if err != nil { return fatal(err) }
			err = validateServiceConfirmation(client_info, service)
			if err != nil { return fatal(err) }
			arg, err := receiveConstructorArgument(conn, service, opts)
			if err != nil { return fatal(err) }
			instance, err := constructServiceInstance(arg, conn, service)
			if err != nil { return fatal(err) }
			err = serverProcessMessages(instance, conn, logger, service, opts)
			if err != nil { return fatal(err) }
			return struct{}{}
		}
		var timeout = rx.TimeoutPair {
			ReadTimeout:  opts.RecvTimeout,
			WriteTimeout: opts.SendTimeout,
		}
		return rx.NewConnectionHandler(raw_conn, timeout, func(conn *rx.WrappedConnection) {
			handle(conn)
		}).Catch(func(err rx.Object) rx.Action {
			logger.LogError(err.(error))
			return rx.Noop()
		})
	})
}


func receiveServiceConfirmation(conn io.Reader) (ServiceIdentifier, error) {
	kind, _, payload, err := receiveMessage(conn)
	if err != nil {
		return ServiceIdentifier{},
			fmt.Errorf("failed to receive service confirmation: %w", err)
	}
	if kind != MSG_SERVICE {
		return ServiceIdentifier{},
			errors.New(fmt.Sprintf("unexpected message kind: %s", kind))
	}
	id, err := ParseServiceIdentifier(string(payload))
	if err != nil {
		return ServiceIdentifier{},
			fmt.Errorf("failed to parse service confirmation: %w", err)
	}
	return id, nil
}
func validateServiceConfirmation(id ServiceIdentifier, service Service) error {
	if id != service.ServiceIdentifier {
		return errors.New("service id not correct")
	}
	return nil
}

func receiveConstructorArgument(conn io.Reader, service Service, opts *ServerOptions) (kmd.Object, error) {
	var ctor = service.Constructor
	var limit = opts.RecvMaxObjectSize
	arg, err := receiveObject(ctor.ArgType, conn, limit, opts.KmdApi)
	if err != nil {
		return nil, fmt.Errorf("failed to receive ctor argument: %w", err)
	}
	return arg, nil
}
func constructServiceInstance(arg kmd.Object, conn *rx.WrappedConnection, service Service) (kmd.Object, error) {
	var construct = service.Constructor.GetAction(arg, conn)
	var sched = conn.Scheduler()
	var ctx = conn.Context()
	opt_instance, ok := rx.ScheduleSingle(construct, sched, ctx)
	if !(ok) {
		return nil, errors.New("service instance construction was cancelled")
	}
	if !(opt_instance.HasValue) {
		var e = opt_instance.Value.(error)
		err := sendConstructorException(e, conn)
		if err != nil { return nil, err }
		return nil, errors.New("service instance construction failed")
	}
	var instance = opt_instance.Value
	var destruct_on_close = conn.OnClose().Catch(func(_ rx.Object) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			return nil, true
		})
	}).Then(func(_ rx.Object) rx.Action {
		return service.Destructor.GetAction(instance)
	})
	rx.ScheduleBackground(destruct_on_close, conn.Scheduler())
	err := sendInstanceCreated(conn)
	if err != nil { return nil, err }
	return instance, nil
}
func sendConstructorException(e error, conn *rx.WrappedConnection) error {
	err := sendError(e, ^uint64(0), conn)
	if err != nil {
		return fmt.Errorf("error sending constructor exception: %w", err)
	}
	return nil
}
func sendInstanceCreated(conn *rx.WrappedConnection) error {
	err := sendMessage(MSG_CREATED, ^uint64(0), ([] byte {}), conn)
	if err != nil {
		return fmt.Errorf("error sending instance created notification: %w", err)
	}
	return nil
}

func serverProcessMessages(instance kmd.Object, conn *rx.WrappedConnection, logger *ServerLogger, service Service, opts *ServerOptions) error {
	var interval = opts.RecvInterval
	for {
		if interval != 0 {
			<- time.After(interval)
		}
		var kind, id, payload, err = receiveMessage(conn)
		if err != nil { return fmt.Errorf("error receiving message: %w", err) }
		switch kind {
		case MSG_CALL, MSG_CALL_MULTI:
			var method_name = string(payload)
			var method, exists = service.Methods[method_name]
			if !(exists) {
				return errors.New(fmt.Sprintf("method '%s' does not exist",
					method_name))
			}
			if (kind == MSG_CALL && method.MultiValue) ||
				(kind == MSG_CALL_MULTI && !(method.MultiValue)) {
				return errors.New(fmt.Sprintf("wrong quantifier (method: '%s')",
					method_name))
			}
			arg, err := receiveCallArgument(method, conn, opts)
			if err != nil { return err }
			var action = method.GetAction(instance, arg)
			var worker = conn.Worker()
			var with_worker = func(do (func() error)) rx.Action {
				return rx.NewQueuedNoValue(worker, func() (bool, rx.Object) {
					err := do()
					if err != nil { return false, err }
					return true, nil
				})
			}
			var send_value = func(value kmd.Object) rx.Action {
				return with_worker(func() error {
					return sendCallReturnValue(value, id, method, conn, opts)
				})
			}
			var send_exception = func(e kmd.Object) rx.Action {
				var e_as_error, e_is_error = e.(error)
				if !(e_is_error) { panic("invalid exception") }
				return with_worker(func() error {
					return sendCallException(e_as_error, id, conn)
				})
			}
			var send_completion = func(err_val kmd.Object) rx.Action {
				return with_worker(func() error {
					return sendCallCompletion(id, conn)
				})
			}
			var send_all =
				action.
				Catch(send_exception).
				ConcatMap(send_value).
				WaitComplete().
				Then(send_completion).
				Catch(func(err_ rx.Object) rx.Action {
					logger.LogError(err.(error))
					return rx.Noop()
				})
			rx.Schedule(send_all, conn.Scheduler(), rx.Receiver {
				Context: conn.Context(),
			})
		default:
			return errors.New(fmt.Sprintf("unknown message kind: %s", kind))
		}
	}
}
func receiveCallArgument(method ServiceMethod, conn *rx.WrappedConnection, opts *ServerOptions) (kmd.Object, error) {
	var limit = opts.RecvMaxObjectSize
	arg, err := receiveObject(method.ArgType, conn, limit, opts.KmdApi)
	if err != nil {
		return nil, fmt.Errorf("failed to receive method argument: %w", err)
	}
	return arg, nil
}
func sendCallReturnValue(value kmd.Object, id uint64, method ServiceMethod, conn *rx.WrappedConnection, opts *ServerOptions) error {
	err := sendMessage(MSG_VALUE, id, ([] byte {}), conn)
	if err != nil {
		return fmt.Errorf("error sending value event header: %w", err)
	}
	err = sendObject(value, method.RetType, conn, opts.KmdApi)
	if err != nil {
		return fmt.Errorf("error sending value event object: %w", err)
	}
	return nil
}
func sendCallException(e error, id uint64, conn *rx.WrappedConnection) error {
	err := sendError(e, id, conn)
	if err != nil {
		return fmt.Errorf("error sending exception event: %w", err)
	}
	return nil
}
func sendCallCompletion(id uint64, conn *rx.WrappedConnection) error {
	err := sendMessage(MSG_COMPLETE, id, ([] byte {}), conn)
	if err != nil {
		return fmt.Errorf("error sending completion event: %w", err)
	}
	return nil
}

