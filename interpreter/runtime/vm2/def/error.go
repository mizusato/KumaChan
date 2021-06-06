package def

import (
	"fmt"
	. "kumachan/standalone/util/error"
)


type RuntimeError struct {
	Content    interface{}
	FrameAddr  uintptr
	FrameData  AddrSpace
	Location
}
type Location struct {
	Function  *FunctionEntity
	InstPtr   LocalAddr
}

func (err *RuntimeError) Error() string {
	var err_desc string
	switch e := err.Content.(type) {
	case fmt.Stringer:
		err_desc = e.String()
	case string:
		err_desc = e
	default:
		err_desc = "unknown error"
	}
	var buf = make(ErrorMessage, 0)
	buf.WriteText(TS_BOLD, fmt.Sprintf("*** Frame {0x%x}", err.FrameAddr))
	buf.WriteText(TS_NORMAL, "\n*\n")
	// NOTE: currently, backtrace is not available (only 1 frame will be listed)
	var frame_msg = make(ErrorMessage, 0)
	frame_msg.WriteText(TS_NORMAL, fmt.Sprintf("Runtime Error: %s", err_desc))
	var point = err.Function.SrcMap[err.InstPtr]
	buf.WriteAll(FormatErrorAt(point, frame_msg))
	return buf.String()
}


