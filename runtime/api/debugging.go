package api

import (
	"os"
	"fmt"
	. "kumachan/lang"
)


var DebuggingFunctions = map[string] interface{} {
	"trace": func(value Value, h InteropContext) Value {
		const bold = "\033[1m"
		const reset = "\033[0m"
		var point = h.ErrorPoint()
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
}
