package ui

import (
	"os"
	"fmt"
	"reflect"
	"encoding/json"
	"kumachan/misc/rx"
	"kumachan/runtime/lib/ui/qt"
	"kumachan/runtime/lib/ui/vdom"
)


func handleEvent(view qt.Widget, sched rx.Scheduler) {
	var payload = qt.WebViewGetCurrentEventPayload(view)
	var handler_id = qt.WebViewGetCurrentEventHandler(view)
	var handler, exists = lookupEventHandler(string(handler_id))
	if !(exists) {
		// events emitted to detached handlers are ignored
		qt.WebViewConsumeEventPayload(payload,
			func(*qt.WebViewEventPayload) interface{} { return nil })
		return
	}
	var sink = handler.Handler.(rx.Sink)
	var handling = sink.Emit(payload)
	rx.ScheduleBackgroundWaitTerminate(handling, sched)
}

func scheduleUpdate(view qt.Widget, sched rx.Scheduler, vdom_source rx.Observable, debug bool) {
	var single_update = virtualDomUpdate(view, debug)
	var update = vdom_source.ConcatMap(func(new_root rx.Object) rx.Observable {
		return single_update(new_root.(*vdom.Node))
	})
	rx.ScheduleBackground(update, sched)
}

var virtualDomUpdate = func(view qt.Widget, debug bool) func(*vdom.Node)(rx.Observable) {
	var patchOpBuffer = make([] interface {}, 0)
	var serializePatchOperations = func() ([] byte) {
		var bin, err = json.Marshal(patchOpBuffer)
		if err != nil { panic("something went wrong") }
		patchOpBuffer = patchOpBuffer[0:0]
		return bin
	}
	var patchOperationCollector = getPatchOperationCollector(&patchOpBuffer)
	var virtualDomRoot *vdom.Node = nil
	return func(new_root *vdom.Node) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			var prev_root = virtualDomRoot
			virtualDomRoot = new_root
			vdom.Diff(patchOperationCollector, nil, prev_root, new_root)
			var patch_data = serializePatchOperations()
			qt.CommitTask(func() {
				qt.WebViewPatchActualDOM(view, patch_data)
			})
			if debug {
				var ctx = vdom.InspectContext { GetHandlerId: getEventHandlerId }
				fmt.Fprintf(os.Stderr, "\033[1m<!-- Virtual DOM Update -->\033[0m\n")
				fmt.Fprintf(os.Stderr, "%s", vdom.Inspect(new_root, ctx))
				fmt.Fprintf(os.Stderr, "\033[1m<!-- Patch Operation Sequence -->\033[0m\n")
				_, _ = os.Stderr.Write(patch_data)
				fmt.Fprintf(os.Stderr, "\n\n")
			}
			return nil, true
		})
	}
}

func getPatchOperationCollector(buf *([] interface{})) *vdom.DeltaNotifier {
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
						var id = registerEventHandler(arg)
						// when attaching an event,
						// pass the handler argument in the form of ID
						*buf = append(*buf, id)
					} else if op == "DetachEvent" {
						unregisterEventHandler(arg)
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

