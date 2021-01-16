package ui

import (
	"kumachan/rx"
	"kumachan/runtime/lib/ui/qt"
	"kumachan/runtime/lib/ui/vdom"
)


func handleEvent(sched rx.Scheduler) {
	var handler = qt.WebUiGetEventHandler()
	var payload = qt.WebUiGetEventPayload()
	var sink = handler.(rx.Sink)
	var handling = sink.Emit(payload)
	rx.ScheduleTask(handling, sched)
}

func scheduleUpdate(sched rx.Scheduler, vdom_source rx.Effect) {
	var single_update = func(root_node rx.Object) rx.Effect {
		return virtualDomUpdate(root_node.(*vdom.Node))
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
}
var virtualDomRoot *vdom.Node = nil
var virtualDomUpdate = func(new_root *vdom.Node) rx.Effect {
	return rx.NewCallback(func(done func(rx.Object)) {
		qt.CommitTask(func() {
			var ctx = virtualDomDeltaNotifier
			var prev_root = virtualDomRoot
			virtualDomRoot = new_root
			vdom.Diff(ctx, nil, prev_root, new_root)
			qt.WebUiPerformActualRendering()
			done(nil)
		})
	})
}

