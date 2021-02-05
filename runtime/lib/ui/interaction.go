package ui

import (
	"os"
	"fmt"
	"reflect"
	"encoding/json"
	"kumachan/rx"
	"kumachan/runtime/lib/ui/qt"
	"kumachan/runtime/lib/ui/vdom"
)


func handleEvent(sched rx.Scheduler) {
	var handler, valid = qt.WebUiGetCurrentEventHandler()
	var payload = qt.WebUiGetCurrentEventPayload()
	if !(valid) {
		// events emitted to detached handlers are ignored
		qt.WebUiConsumeEventPayload(payload,
			func(*qt.WebUiEventPayload) interface{} { return nil })
		return
	}
	var sink = handler.Handler.(rx.Sink)
	var handling = sink.Emit(payload)
	rx.ScheduleActionWaitTerminate(handling, sched)
}

func scheduleUpdate(sched rx.Scheduler, vdom_source rx.Action, debug bool) {
	var single_update = func(root_node rx.Object) rx.Action {
		return virtualDomUpdate(root_node.(*vdom.Node), debug)
	}
	var update = vdom_source.ConcatMap(single_update)
	rx.ScheduleAction(update, sched)
}

var patchOpBuffer = make([] interface {}, 0)
func clearPatchOperationBuffer() {
	patchOpBuffer = patchOpBuffer[0:0]
}
func serializePathOperations() ([] byte) {
	var bin, err = json.Marshal(patchOpBuffer)
	if err != nil { panic("something went wrong") }
	return bin
}
var virtualDomDeltaNotifier = patchOperationCollector(&patchOpBuffer)
var virtualDomRoot *vdom.Node = nil
var virtualDomUpdate = func(new_root *vdom.Node, debug bool) rx.Action {
	return rx.NewCallback(func(done func(rx.Object)) {
		if debug {
			fmt.Fprintf(os.Stderr, "\033[1m<!-- Virtual DOM Update -->\033[0m\n")
			fmt.Fprintf(os.Stderr, "%s", vdom.Inspect(new_root))
		}
		var ctx = virtualDomDeltaNotifier
		var prev_root = virtualDomRoot
		virtualDomRoot = new_root
		clearPatchOperationBuffer()
		vdom.Diff(ctx, nil, prev_root, new_root)
		var patch_data = serializePathOperations()
		if debug {
			fmt.Fprintf(os.Stderr, "\033[1m<!--Patch Operation Sequence -->\033[0m\n")
			_, _ = os.Stderr.Write(patch_data)
			fmt.Fprintf(os.Stderr, "\n\n")
		}
		qt.CommitTask(func() {
			qt.WebUiPatchActualDOM(patch_data)
			done(nil)
		})
	})
}

func patchOperationCollector(buf *([] interface{})) *vdom.DeltaNotifier {
	var writeOperation = func(op string) {
		*buf = append(*buf, op)
	}
	var writeBoolArgument = func(arg bool) {
		*buf = append(*buf, arg)
	}
	var writeStringArgument = func(arg vdom.String) {
		*buf = append(*buf, string(arg))
	}
	var v = reflect.ValueOf(&vdom.DeltaNotifier {})
	var t = v.Elem().Type()
	for i := 0; i < t.NumField(); i += 1 {
		var field = t.Field(i)
		var op = field.Name
		var callback = reflect.MakeFunc(field.Type, func(args ([] reflect.Value)) ([] reflect.Value) {
			writeOperation(op)
			for _, arg_ := range args {
				var arg = arg_.Interface()
				switch arg := arg.(type) {
				case bool:
					writeBoolArgument(arg)
				case vdom.String:
					writeStringArgument(arg)
				case *vdom.EventHandler:
					if op == "AttachEvent" {
						var id = qt.WebUiRegisterEventHandler(arg)
						// when attaching an event,
						// pass the handler argument in the form of ID
						*buf = append(*buf, string(id))
					} else if op == "DetachEvent" {
						qt.WebUiUnregisterEventHandler(arg)
						// when detaching an event,
						// the handler argument is not necessary to pass
					} else {
						panic("something went wrong")
					}
				default:
					panic("something went wrong")
				}
			}
			return nil
		})
		v.Elem().Field(i).Set(callback)
	}
	return v.Interface().(*vdom.DeltaNotifier)
}

