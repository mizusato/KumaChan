package rpc

import (
	"io"
	"fmt"
	"net"
	"bytes"
	"errors"
	"compress/gzip"
	"kumachan/rx"
	"kumachan/rpc/kmd"
)


type ServerOptions struct {
	Listener  net.Listener
	Debugger  ServerDebugger
	KmdApi
}
type ServerDebugger interface {
	LogError(err error, local net.Addr, remote net.Addr)
}
type KmdApi interface {
	SerializeToStream(v kmd.Object, t *kmd.Type, stream io.Writer) error
	DeserializeFromStream(t *kmd.Type, stream io.Reader) (kmd.Object, error)
}

type ServerLogger struct {
	LocalAddr   net.Addr
	RemoteAddr  net.Addr
	Debugger    ServerDebugger
}
func (l ServerLogger) LogError(err error) {
	if l.Debugger != nil {
		l.Debugger.LogError(err, l.LocalAddr, l.RemoteAddr)
	}
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
			Debugger:   opts.Debugger,
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
			err = processMessages(instance, conn, logger, service, opts)
			if err != nil { return fatal(err) }
			return struct{}{}
		}
		return rx.NewConnectionHandler(raw_conn, func(conn *rx.WrappedConnection) {
			handle(conn)
		}).Catch(func(err rx.Object) rx.Action {
			logger.LogError(err.(error))
			return rx.Noop()
		})
	})
}

type ServiceConfirmation struct {
	ServiceName     string
	ServiceVersion  string
}
func receiveServiceConfirmation(conn io.Reader) (*ServiceConfirmation, error) {
	kind, _, payload, err := receiveMessage(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to receive service confirmation: %w", err)
	}
	if kind != "service" {
		return nil, errors.New(fmt.Sprintf("unexpected message kind: %s", kind))
	}
	var buf_reader = bytes.NewReader(payload)
	var name string
	var version string
	_, err = fmt.Fscanf(buf_reader, "%s %s", &name, &version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client metadata: %w", err)
	}
	return &ServiceConfirmation {
		ServiceName:    name,
		ServiceVersion: version,
	}, nil
}
func validateServiceConfirmation(client_info *ServiceConfirmation, service Service) error {
	if client_info.ServiceName != service.Name {
		return errors.New("service name not correct")
	}
	if client_info.ServiceVersion != service.Version {
		return errors.New("service version not correct")
	}
	return nil
}

func receiveConstructorArgument(conn io.Reader, service Service, opts *ServerOptions) (kmd.Object, error) {
	var ctor = service.Constructor
	arg, err := opts.DeserializeFromStream(ctor.ArgType, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to receive ctor argument: %w", err)
	}
	return arg, nil
}
func constructServiceInstance(arg kmd.Object, conn *rx.WrappedConnection, service Service) (kmd.Object, error) {
	var construct = service.Constructor.GetAction(arg)
	var sched = conn.Scheduler()
	var ctx = conn.Context()
	instance, ok := rx.BlockingRunSingle(construct, sched, ctx)
	if !(ok) {
		return nil, errors.New("failed to construct service instance")
	}
	return instance, nil
}

func processMessages(instance kmd.Object, conn *rx.WrappedConnection, logger *ServerLogger, service Service, opts *ServerOptions) error {
	for {
		var kind, id, payload, err = receiveMessage(conn)
		if err != nil { return fmt.Errorf("error receiving message: %w", err) }
		switch kind {
		case "call":
			var method_name = string(payload)
			var method, exists = service.Methods[method_name]
			if !(exists) { return errors.New(fmt.Sprintf(
				"method '%s' does not exist", method_name)) }
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
			conn.Scheduler().RunTopLevel(send_all, rx.Receiver {
				Context: conn.Context(),
			})
		default:
			return errors.New(fmt.Sprintf("unknown message kind: %s", kind))
		}
	}
}
func receiveCallArgument(method ServiceMethod, conn *rx.WrappedConnection, opts *ServerOptions) (kmd.Object, error) {
	decompressed, err := gzip.NewReader(conn)
	if err != nil { panic(err) }
	arg, err := opts.DeserializeFromStream(method.ArgType, decompressed)
	if err != nil {
		return nil, fmt.Errorf("failed to receive method argument: %w", err)
	}
	return arg, nil
}
func sendCallReturnValue(value kmd.Object, id uint64, method ServiceMethod, conn *rx.WrappedConnection, opts *ServerOptions) error {
	err := sendMessage("value", id, ([] byte {}), conn)
	if err != nil {
		return fmt.Errorf("error sending value header: %w", err)
	}
	var compressed = gzip.NewWriter(conn)
	err = opts.SerializeToStream(value, method.RetType, compressed)
	if err != nil {
		return fmt.Errorf("error sending value: %w", err)
	}
	return nil
}
func sendCallException(e error, id uint64, conn *rx.WrappedConnection) error {
	var desc = e.Error()
	var desc_bin = ([] byte)(desc)
	err := sendMessage("error", id, desc_bin, conn)
	if err != nil {
		return fmt.Errorf("error sending exception: %w", err)
	}
	return nil
}
func sendCallCompletion(id uint64, conn *rx.WrappedConnection) error {
	err := sendMessage("complete", id, ([] byte {}), conn)
	if err != nil {
		return fmt.Errorf("error sending completion: %w", err)
	}
	return nil
}

