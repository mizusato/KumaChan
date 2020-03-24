package vm

import (
	"fmt"
	. "kumachan/error"
	"os"
	"strings"
)

func PrintRuntimeErrorMessage(err interface{}, ec *ExecutionContext) {
	var err_desc string
	switch e := err.(type) {
	case fmt.Stringer:
		err_desc = e.String()
	default:
		err_desc = "unknown error"
	}
	var buf strings.Builder
	fmt.Fprintf(&buf, "%v*** Runtime Error%v\n*\n", Bold, Reset)
	var L = len(ec.callStack)
	for i := 0; i < L; i += 1 {
		var this = ec.callStack[i]
		var callee CallStackFrame
		if i+1 < L {
			callee = ec.callStack[i+1]
		} else {
			callee = ec.workingFrame
		}
		var callee_name = callee.function.Info.Name
		var frame_msg = fmt.Sprintf("%s called", callee_name)
		buf.WriteString(GenFrameErrMsg(this, frame_msg))
		buf.WriteString("\n*\n")
	}
	var frame_msg = fmt.Sprintf("Runtime Error: %s", err_desc)
	buf.WriteString(GenFrameErrMsg(ec.workingFrame, frame_msg))
	var msg = buf.String()
	var _, _ = fmt.Fprintln(os.Stderr, msg)
}

func GenFrameErrMsg(f CallStackFrame, desc string) string {
	var point = f.function.Info.CodeMap[f.instPtr]
	return point.GenErrMsg(desc)
}
