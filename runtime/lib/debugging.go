package lib

import (
	"os"
	"fmt"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
	"kumachan/runtime/common/rx"
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
		panic("programmed panic: " + string(msg))
	},
	"crash": func(msg String, h MachineHandle) rx.Effect {
		const bold = "\033[1m"
		const red = "\033[31m"
		const reset = "\033[0m"
		var point = h.GetErrorPoint()
		var source_point = point.Node.Point
		return rx.CreateBlockingEffect(func() (rx.Object, bool) {
			fmt.Fprintf (
				os.Stderr, "%v*** Crash: (%d, %d) at %s%v\n",
				bold+red,
				source_point.Row, source_point.Col, point.Node.CST.Name,
				reset,
			)
			fmt.Fprintf (
				os.Stderr, "%v%s%v\n",
				bold+red, string(msg), reset,
			)
			os.Exit(255)
			panic("program should have crashed")
		})
	},
}
