package lib

import (
	"os"
	"fmt"
	. "kumachan/runtime/common"
)


var DebuggingFunctions = map[string] interface{} {
	"trace": func(value Value, h MachineHandle) Value {
		const bold = "\033[1m"
		const reset = "\033[0m"
		var point = h.GetErrorPoint()
		var source_point = point.Node.Point
		fmt.Fprintf (
			os.Stderr, "--- %vtrace:%v (%d, %d) at %s\n",
			bold, reset, source_point.Row,
			source_point.Col, point.Node.CST.Name,
		)
		var repr = Inspect(value)
		fmt.Fprintf(os.Stderr, "%s\n", repr.String())
		return value
	},
	"panic": func(msg String) struct{} {
		panic("programmed panic: " + GoStringFromString(msg))
	},
}
