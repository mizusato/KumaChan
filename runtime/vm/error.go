package vm

import (
	"fmt"
	. "kumachan/error"
	"os"
)

func PrintRuntimeErrorMessage(err interface{}, ec *ExecutionContext) {
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
	buf.WriteText(TS_BOLD, "*** Runtime Error")
	buf.WriteText(TS_NORMAL, "\n*\n")
	var L = len(ec.callStack)
	for i := 1; i < L; i += 1 {
		var callee = ec.callStack[i]
		var callee_name = callee.function.Info.Name
		var frame_msg = make(ErrorMessage, 0)
		frame_msg.WriteText(TS_NORMAL, fmt.Sprintf("%s called", callee_name))
		buf.WriteAll(GenFrameErrMsg(callee, frame_msg))
		buf.WriteText(TS_NORMAL, "\n*\n")
	}
	var frame_msg = make(ErrorMessage, 0)
	frame_msg.WriteText(TS_NORMAL, fmt.Sprintf("Runtime Error: %s", err_desc))
	buf.WriteAll(GenFrameErrMsg(ec.workingFrame, frame_msg))
	buf.Write(T_LF)
	var msg = buf.String()
	var _, _ = fmt.Fprintln(os.Stderr, msg)
}

func GenFrameErrMsg(f CallStackFrame, desc ErrorMessage) ErrorMessage {
	var last_executed = (f.instPtr - 1)
	return FormatErrorAt(ErrorPoint {
		CST:  f.function.Info.DeclPoint.CST,
		Node: *(f.function.Info.SourceMap[last_executed]),
	}, desc)
}
