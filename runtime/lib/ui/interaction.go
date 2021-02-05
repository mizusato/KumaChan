package ui

import (
	"os"
	"fmt"
	"kumachan/rx"
	"kumachan/runtime/lib/ui/qt"
	"kumachan/runtime/lib/ui/vdom"
)


var updating = false

func handleEvent(sched rx.Scheduler) {
	var handler = qt.WebUiGetEventHandler()
	var payload = qt.WebUiGetEventPayload()
	if updating {
		// in order to preserve data consistency,
		// events emitted during updating are ignored
		qt.WebUiConsumeEventPayload(payload,
			func(*qt.WebUiEventPayload) interface{} { return nil })
		return
	}
	var sink = handler.(rx.Sink)
	var handling = sink.Emit(payload)
	rx.ScheduleTaskWaitTerminate(handling, sched)
}

func scheduleUpdate(sched rx.Scheduler, vdom_source rx.Action, debug bool) {
	var single_update = func(root_node rx.Object) rx.Action {
		return virtualDomUpdate(root_node.(*vdom.Node), debug)
	}
	var update = vdom_source.ConcatMap(single_update)
	rx.ScheduleTask(update, sched)
}

var virtualDomDeltaNotifier = &vdom.DeltaNotifier {
	ApplyStyle:  qt.WebUiApplyStyle,
	EraseStyle:  qt.WebUiEraseStyle,
	SetAttr:     qt.WebUiSetAttr,
	RemoveAttr:  qt.WebUiRemoveAttr,
	AttachEvent: qt.WebUiAttachEvent,
	ModifyEvent: qt.WebUiModifyEvent,
	DetachEvent: qt.WebUiDetachEvent,
	SetText:     qt.WebUiSetText,
	AppendNode:  qt.WebUiAppendNode,
	RemoveNode:  qt.WebUiRemoveNode,
	UpdateNode:  qt.WebUiUpdateNode,
	ReplaceNode: qt.WebUiReplaceNode,
	SwapNode:    qt.WebUiSwapNode,
	MoveNode:    qt.WebUiMoveNode,
}
var virtualDomRoot *vdom.Node = nil
var virtualDomUpdate = func(new_root *vdom.Node, debug bool) rx.Action {
	return rx.NewCallback(func(done func(rx.Object)) {
		if debug {
			fmt.Fprintf(os.Stderr, "\033[1m<!-- Virtual DOM Update -->\033[0m\n")
			fmt.Fprintf(os.Stderr, "%s\n", vdom.Inspect(new_root))
		}
		updating = true
		qt.CommitTask(func() {
			var ctx = virtualDomDeltaNotifier
			var prev_root = virtualDomRoot
			virtualDomRoot = new_root
			vdom.Diff(ctx, nil, prev_root, new_root)
			qt.WebUiPerformActualRendering()
			done(nil)
		})
	}).Then(func(_ rx.Object) rx.Action {
		return rx.NewSync(func() (rx.Object, bool) {
			updating = false
			return nil, true
		})
	})
}

