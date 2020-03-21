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
	var L = len(ec.CallStack)
	for i := 0; i < L; i += 1 {
		var this = ec.CallStack[i]
		var callee CallStackFrame
		if i+1 < L {
			callee = ec.CallStack[i+1]
		} else {
			callee = ec.WorkingFrame
		}
		var callee_name = callee.Function.Info.Name
		var frame_msg = fmt.Sprintf("%s called", callee_name)
		buf.WriteString(GenFrameErrMsg(this, frame_msg))
		buf.WriteString("\n*\n")
	}
	var frame_msg = fmt.Sprintf("Runtime Error: %s", err_desc)
	buf.WriteString(GenFrameErrMsg(ec.WorkingFrame, frame_msg))
	var msg = buf.String()
	var _, _ = fmt.Fprintln(os.Stderr, msg)
}

func GenFrameErrMsg(f CallStackFrame, desc string) string {
	var point = f.Function.Info.CodeMap[f.InstPtr]
	return point.GenErrMsg(desc)
}
