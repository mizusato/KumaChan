package vm

import (
	"os"
	"fmt"
	"reflect"
	. "kumachan/error"
)

func PrintRuntimeErrorMessage(err interface{}, ec *ExecutionContext) {
	var ec_addr = reflect.ValueOf(ec).Pointer()
	var err_desc string
	switch e := err.(type) {
	case fmt.Stringer:
		err_desc = e.String()
	case string:
		err_desc = e
	default:
		err_desc = "unknown error"
	}
	var buf = make(ErrorMessage, 0)
	buf.WriteText(TS_BOLD, fmt.Sprintf("*** Execution Context {0x%x}", ec_addr))
	buf.WriteText(TS_NORMAL, "\n*\n")
	var frame_msg = make(ErrorMessage, 0)
	frame_msg.WriteText(TS_NORMAL, fmt.Sprintf("Runtime Error: %s", err_desc))
	buf.WriteAll(FormatErrorAtFrame(ec.workingFrame, frame_msg))
	var L = len(ec.callStack)
	for i := (L-1); i >= 1; i -= 1 {
		buf.WriteText(TS_NORMAL, "\n*\n")
		var callee = ec.callStack[i]
		var callee_name = callee.function.Info.Name
		var frame_msg = make(ErrorMessage, 0)
		frame_msg.WriteText(TS_NORMAL, fmt.Sprintf("call from %s", callee_name))
		buf.WriteAll(FormatErrorAtFrame(callee, frame_msg))
	}
	var msg = buf.String()
	var _, _ = fmt.Fprintln(os.Stderr, msg)
}

func FormatErrorAtFrame(f CallStackFrame, desc ErrorMessage) ErrorMessage {
	return FormatErrorAt(GetFrameErrorPoint(f), desc)
}

func GetFrameErrorPoint(f CallStackFrame) ErrorPoint {
	if f.instPtr == 0 { panic("something went wrong") }
	if f.function == nil { panic("something went wrong") }
	var last_executed = (f.instPtr - 1)
	return f.function.Info.SourceMap[last_executed]
}
