package lib

import (
	"fmt"
	. "kumachan/runtime/common"
	"os"
)


var DebuggingFunctions = map[string] interface{} {
	"trace": func(value Value, h MachineHandle) Value {
		var point = h.GetErrorPoint()
		var source_point = point.Node.Point
		fmt.Fprintf (
			os.Stderr, "--- \033[1mtrace:\033[0m (%d, %d) at %s\n",
			source_point.Row, source_point.Col, point.CST.Name,
		)
		var repr = Inspect(value)
		fmt.Fprintf(os.Stderr, "%s\n", repr.String())
		return value
	},
	"panic": func(msg []rune) struct{} {
		panic("programmed panic: " + string(msg))
	},
}
