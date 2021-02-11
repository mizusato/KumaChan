package rpc

import (
	"io"
	"fmt"
	"net"
	"bytes"
	"errors"
	"compress/gzip"
	"kumachan/rx"
	. "kumachan/lang"
)


type ServerOptions struct {
	Listener  net.Listener
	Debugger  ServerDebugger
	KmdApi
}
type ServerDebugger interface {
	LogError(err error, local net.Addr, remote net.Addr)
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
	}).ConcatMap(func(raw_conn_ rx.Object) rx.Action {
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
			metadata, err := receiveClientMetadata(conn)
			if err != nil { return fatal(err) }
			err = validateClientMetadata(metadata, service)
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

const ClientMetadataLength = 1024
type ClientMetadata struct {
	ServiceName     string
	ServiceVersion  string
}
func receiveClientMetadata(conn io.Reader) (*ClientMetadata, error) {
	var buf_ ([ClientMetadataLength] byte)
	var buf = buf_[:]
	_, err := io.ReadFull(conn, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to receive client metadata: %w", err)
	}
	var buf_reader = bytes.NewReader(buf)
	var name string
	var version string
	_, err = fmt.Fscanf(buf_reader, "%s %s", &name, &version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse client metadata: %w", err)
	}
	return &ClientMetadata {
		ServiceName:    name,
		ServiceVersion: version,
	}, nil
}
func validateClientMetadata(metadata *ClientMetadata, service Service) error {
	if metadata.ServiceName != service.Name {
		return errors.New("service name not correct")
	}
	if metadata.ServiceVersion != service.Version {
		return errors.New("service version not correct")
	}
	return nil
}

func receiveConstructorArgument(conn io.Reader, service Service, opts *ServerOptions) (Value, error) {
	var ctor = service.Constructor
	arg, err := opts.DeserializeFromStream(ctor.ArgType, conn)
	if err != nil {
		return nil, fmt.Errorf("failed to receive ctor argument: %w", err)
	}
	return arg, nil
}
func constructServiceInstance(arg Value, conn *rx.WrappedConnection, service Service) (Value, error) {
	var construct = service.Constructor.GetAction(arg)
	var sched = conn.Scheduler()
	var ctx = conn.Context()
	instance, ok := rx.BlockingRunSingle(construct, sched, ctx)
	if !(ok) {
		return nil, errors.New("failed to construct service instance")
	}
	return instance, nil
}

func processMessages(instance Value, conn *rx.WrappedConnection, logger *ServerLogger, service Service, opts *ServerOptions) error {
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
			var worker = rx.CreateWorker()
			var send_value = func(value Value) rx.Action {
				return rx.NewQueued(worker, func() (rx.Object, bool) {
					err := sendCallReturnValue(value, id, method, conn, opts)
					if err != nil { return err, false }
					return nil, true
				})
			}
			var send_exception = func(e Value) rx.Action {
				return rx.NewQueued(worker, func() (rx.Object, bool) {
					var e_as_error, e_is_error = e.(error)
					if !(e_is_error) {
						panic("invalid exception value thrown by rpc call")
					}
					err := sendCallException(e_as_error, id, conn)
					if err != nil { return err, false }
					return nil, true
				})
			}
			var send_completion = func(err_val Value) rx.Action {
				return rx.NewQueued(worker, func() (rx.Object, bool) {
					err := sendCallCompletion(id, conn)
					if err != nil { return err, false }
					return nil, true
				})
			}
			var watch_normal =
				action.ConcatMap(send_value).
				WaitComplete().
				Then(send_completion)
			var watch_exception =
				action.Catch(send_exception)
			var watch_all =
				rx.Merge([] rx.Action { watch_normal, watch_exception }).
				Catch(func(err_ rx.Object) rx.Action {
					logger.LogError(err.(error))
					return rx.Noop()
				})
			rx.ScheduleAction(watch_all, conn.Scheduler())
		default:
			return errors.New(fmt.Sprintf("unknown error kind: %s", kind))
		}
	}
}
func receiveCallArgument(method ServiceMethod, conn *rx.WrappedConnection, opts *ServerOptions) (Value, error) {
	decompressed, err := gzip.NewReader(conn)
	if err != nil { panic(err) }
	arg, err := opts.DeserializeFromStream(method.ArgType, decompressed)
	if err != nil {
		return nil, fmt.Errorf("failed to receive method argument: %w", err)
	}
	return arg, nil
}
func sendCallReturnValue(value Value, id uint64, method ServiceMethod, conn *rx.WrappedConnection, opts *ServerOptions) error {
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

