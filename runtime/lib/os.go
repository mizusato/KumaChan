package lib

import (
	"io"
	"os"
	"kumachan/runtime/common/rx"
	. "kumachan/runtime/common"
)


var OS_Constants = map[string] Value {
	"Stdin": io.Reader(os.Stdin),
	"Stdout": io.Writer(os.Stderr),
	"Stderr": io.Writer(os.Stderr),
}

var OS_Functions = map[string] Value {
	"exit": func(code uint8) rx.Effect {
		return rx.CreateBlockingEffect(func() ([]rx.Object, error) {
			os.Exit(int(code))
			panic("process should have exited")
		})
	},
}
