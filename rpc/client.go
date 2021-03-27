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


type ClientOptions struct {
	Connection           net.Conn
	DebugOutput          io.Writer
	ConstructorArgument  kmd.Object
	InstanceConsumer     func(*ClientInstance) rx.Observable
	Limits
	KmdApi
}

func Client(service ServiceInterface, opts *ClientOptions) rx.Observable {
	var raw_conn = opts.Connection
	var logger = ClientLogger {
		LocalAddr:  raw_conn.LocalAddr(),
		RemoteAddr: raw_conn.RemoteAddr(),
		Output:     opts.DebugOutput,
	}
	var handle = func(conn *rx.WrappedConnection) struct{} {
		var fatal = func(err error) struct{} {
			conn.Fatal(err)
			return struct{}{}
		}
		err := sendServiceConfirmation(conn, service)
		if err != nil { return fatal(err) }
		err = sendConstructorArgument(conn, service, opts)
		if err != nil { return fatal(err) }
		err = receiveInstanceCreated(conn)
		if err != nil { return fatal(err) }
		var instance = createClientInstance(conn, logger, service, opts)
		consumeClientInstance(instance, conn, opts)
		err = clientProcessMessages(instance, conn, opts)
		if err != nil { return fatal(err) }
		return struct{}{}
	}
	var timeout = rx.TimeoutPair {
		ReadTimeout:  opts.RecvTimeout,
		WriteTimeout: opts.SendTimeout,
	}
	return rx.NewConnectionHandler(raw_conn, timeout, func(conn *rx.WrappedConnection) {
		handle(conn)
	}).Catch(func(err rx.Object) rx.Observable {
		logger.LogError(err.(error))
		return rx.Throw(err)
	})
}


type ClientInstance struct {
	connection  *rx.WrappedConnection
	requester   *rx.Worker
	logger      ClientLogger
	service     ServiceInterface
	options     *ClientOptions
	state       ClientInstanceState
}
type ClientInstanceState struct {
	mutator     *rx.Worker
	calls       map[uint64] Call
	nextCallId  uint64
}
type Call struct {
	sender   rx.Sender
	retType  *kmd.Type
}
func createClientInstance(conn *rx.WrappedConnection, logger ClientLogger, service ServiceInterface, opts *ClientOptions) *ClientInstance {
	return &ClientInstance {
		connection: conn,
		requester:  rx.CreateWorker(),
		logger:     logger,
		service:    service,
		options:    opts,
		state: ClientInstanceState {
			mutator:    rx.CreateWorker(),
			calls:      make(map[uint64] Call),
			nextCallId: 0,
		},
	}
}
func (instance *ClientInstance) Call(method_name string, arg kmd.Object) rx.Observable {
	var method, exists = instance.service.Methods[method_name]
	if !(exists) { panic("something went wrong") }
	return rx.NewSyncWithSender(func(sender rx.Sender) {
		instance.state.mutator.Do(func() {
			var id = instance.state.nextCallId
			instance.state.nextCallId += 1
			instance.state.calls[id] = Call {
				sender:  sender,
				retType: method.RetType,
			}
			var send_request = func() struct{} {
				var conn = instance.connection
				var fatal = func(err error) struct{} {
					var wrapped = fmt.Errorf("error sending call request: %w", err)
					conn.Fatal(wrapped)
					instance.logger.LogError(wrapped)
					return struct{}{}
				}
				var method_name_bin = ([] byte)(method_name)
				var msg_kind = (func() string {
					if method.MultiValue {
						return MSG_CALL_MULTI
					} else {
						return MSG_CALL
					}
				})()
				err := sendMessage(msg_kind, id, method_name_bin, conn)
				if err != nil { return fatal(err) }
				err = sendCallArgument(arg, method, conn, instance.options)
				if err != nil { return fatal(err) }
				return struct{}{}
			}
			instance.requester.Do(func() {
				send_request()
			})
		})
	})
}
func (instance *ClientInstance) lookupCall(id uint64) (Call, bool) {
	var call, exists = instance.state.calls[id]
	if !(exists) {
		var err = errors.New(fmt.Sprintf(
			"inconsistent server message: call %d does not exist", id))
		instance.connection.Fatal(err)
		instance.logger.LogError(err)
		return Call{}, false
	}
	return call, true
}
func (instance *ClientInstance) getCallReturnValueType(id uint64) *kmd.Type {
	var wait = make(chan *kmd.Type)
	instance.state.mutator.Do(func() {
		var call, ok = instance.lookupCall(id)
		if !(ok) { return }
		wait <- call.retType
	})
	return <- wait
}
func (instance *ClientInstance) next(id uint64, value kmd.Object) {
	instance.state.mutator.Do(func() {
		var call, ok = instance.lookupCall(id)
		if !(ok) { return }
		call.sender.Next(value)
	})
}
func (instance *ClientInstance) error(id uint64, e error) {
	instance.state.mutator.Do(func() {
		var call, ok = instance.lookupCall(id)
		if !(ok) { return }
		call.sender.Error(e)
	})
}
func (instance *ClientInstance) complete(id uint64) {
	instance.state.mutator.Do(func() {
		var call, ok = instance.lookupCall(id)
		if !(ok) { return }
		delete(instance.state.calls, id)
		call.sender.Complete()
	})
}

func sendServiceConfirmation(conn io.Writer, service ServiceInterface) error {
	var service_id = DescribeServiceIdentifier(service.ServiceIdentifier)
	return sendMessage(MSG_SERVICE, ^uint64(0), ([] byte)(service_id), conn)
}

func sendConstructorArgument(conn io.Writer, service ServiceInterface, opts *ClientOptions) error {
	var ctor = service.Constructor
	var arg = opts.ConstructorArgument
	return sendObject(arg, ctor.ArgType, conn, opts.KmdApi)
}

func receiveInstanceCreated(conn io.Reader) error {
	kind, _, payload, err := receiveMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to receive instance created notification: %w", err)
	}
	if kind != MSG_CREATED {
		if kind == MSG_ERROR {
			var desc = string(payload)
			return errors.New(desc)
		} else {
			return errors.New(fmt.Sprintf("unexpected message kind: %s", kind))
		}
	}
	return nil
}

func consumeClientInstance(instance *ClientInstance, conn *rx.WrappedConnection, opts *ClientOptions) {
	var consume = opts.InstanceConsumer(instance)
	var consume_and_dispose = consume.WaitComplete().Then(func(_ rx.Object) rx.Observable {
		_ = conn.Close()
		return rx.Noop()
	})
	rx.Schedule(consume_and_dispose, conn.Scheduler(), rx.Receiver {
		Context:   conn.Context(),
	})
}

func sendCallArgument(arg kmd.Object, method ServiceMethodInterface, conn *rx.WrappedConnection, opts *ClientOptions) error {
	return sendObject(arg, method.ArgType, conn, opts.KmdApi)
}

func clientProcessMessages(instance *ClientInstance, conn *rx.WrappedConnection, opts *ClientOptions) error {
	var interval = opts.RecvInterval
	for {
		if interval != 0 {
			<- time.After(interval)
		}
		var kind, id, payload, err = receiveMessage(conn)
		if err != nil { return fmt.Errorf("error receving server message: %w", err) }
		switch kind {
		case MSG_VALUE:
			var ret_type = instance.getCallReturnValueType(id)
			var limit = opts.RecvMaxObjectSize
			value, err := receiveObject(ret_type, conn, limit, opts.KmdApi)
			if err != nil { return fmt.Errorf("error receiving value object: %w", err) }
			instance.next(id, value)
		case MSG_ERROR:
			var desc = string(payload)
			var e = errors.New(desc)
			instance.error(id, e)
		case MSG_COMPLETE:
			instance.complete(id)
		default:
			return errors.New(fmt.Sprintf("unknown message kind: %s", kind))
		}
	}
}

