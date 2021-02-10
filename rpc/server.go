package rpc

import (
	"io"
	"fmt"
	"net"
	"sync"
	"bytes"
	"compress/gzip"
	"kumachan/rx"
	. "kumachan/lang"
)


type ServerOptions struct {
	Listener  net.Listener
	KmdApi
}

func Server(service Service, opts ServerOptions) rx.Action {
	return rx.NewGoroutine(func(sender rx.Sender) {
		var l = opts.Listener
		for {
			if sender.Context().AlreadyCancelled() {
				return
			}
			var conn, err = l.Accept()
			if err != nil {
				sender.Error(err)
				return
			}
			go (func() struct{} {
				var fatal = func(reason string) struct{} {
					_  = sendMessage("fatal", ^uint64(0), ([] byte)(reason), conn)
					_  = conn.Close()
					return struct{}{}
				}
				var line string
				_, err := fmt.Fscanln(conn, &line)
				if err != nil { return fatal("failed to receive metadata") }
				var name = line
				_, err = fmt.Fscanln(conn, &line)
				if err != nil { return fatal("failed to receive metadata") }
				var version = line
				if name != service.Name {
					return fatal("service name not correct")
				}
				if version != service.Version {
					return fatal("service version not correct")
				}
				var ctor = service.Constructor
				ctor_arg, err := opts.DeserializeFromStream(ctor.ArgType, conn)
				if err != nil { return fatal("failed to receive ctor argument") }
				var construct = ctor.GetAction(ctor_arg)
				var sched = sender.Scheduler()
				var ctx = sender.Context()
				instance, ok := rx.BlockingRunSingle(construct, sched, ctx)
				if !(ok) { return fatal("failed to construct service instance") }
				var calls = make(Calls)
				var calls_mutex sync.Mutex
				for {
					var kind, id, payload, err = receiveMessage(conn)
					if err != nil { return fatal("error receiving message") }
					switch kind {
					case "call":
						var method = string(payload)
						var method_data, exists = service.Methods[method]
						if !(exists) { return fatal(fmt.Sprintf(
							"method '%s' does not exist", method)) }
						var reader, writer = io.Pipe()
						var call = &Call {
							Method:    method,
							ArgReader: reader,
							ArgWriter: writer,
						}
						calls_mutex.Lock()
						calls[id] = call
						calls_mutex.Unlock()
						// TODO: task queue
						go (func() {
							var arg_type = method_data.ArgType
							var r, gz_err = gzip.NewReader(reader)
							if gz_err != nil { panic(gz_err) }
							arg, err := opts.DeserializeFromStream(arg_type, r)
							if err != nil { _ = reader.CloseWithError(err) }
							calls_mutex.Lock()
							delete(calls, id)
							calls_mutex.Unlock()
							var action = method_data.GetAction(instance, arg)
							var chan_values = make(chan Value)
							var chan_error = make(chan Value)
							sched.RunTopLevel(action, rx.Receiver {
								Context:   ctx,
								Values:    chan_values,
								Error:     chan_error,
								Terminate: nil,
							})
							var ret_type = method_data.RetType
							for {
								select {
								case value, ok := <-chan_values:
									if ok {
										var buf bytes.Buffer
										var w = gzip.NewWriter(&buf)
										err := opts.SerializeToStream(value, ret_type, w)
										if err != nil {
											panic(err)
										}
										var bin = buf.Bytes()
										const chunk_size = 4096
										for pos := 0; pos < len(bin); pos += chunk_size {
											var chunk []byte
											if (pos + chunk_size) < len(bin) {
												chunk = bin[pos: (pos + chunk_size)]
											} else {
												chunk = bin[pos:]
											}
											_ = sendMessage("val", id, chunk, conn)
										}
										_ = sendMessage("val-end", id, nil, conn)
									} else {
										_ = sendMessage("end", id, nil, conn)
										return
									}
								case err, ok := <-chan_error:
									if ok {
										var err_desc = ([] byte)(err.(error).Error())
										_ = sendMessage("err", id, err_desc, conn)
									} else {
										_ = sendMessage("end", id, nil, conn)
										return
									}
								}
							}
						})()
					case "arg-part":
						calls_mutex.Lock()
						var call, exists = calls[id]
						calls_mutex.Unlock()
						if !(exists) { return fatal("invalid operation") }
						var _, err = call.ArgWriter.Write(payload)
						if err != nil { fatal(fmt.Sprintf(
							"invalid argument for method %s: %s",
							call.Method, err.Error())) }
					case "arg-end":
						calls_mutex.Lock()
						var call, exists = calls[id]
						calls_mutex.Unlock()
						if !(exists) { return fatal("invalid operation") }
						var err = call.ArgWriter.Close()
						if err != nil { fatal(fmt.Sprintf(
							"invalid argument for method %s: %s",
							call.Method, err.Error())) }
					}
				}
			})()
		}
	})
}

