package lib

import (
	"io"
	"os"
	"runtime"
	"kumachan/runtime/common/rx"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
)


var OS_Constants = map[string] Value {
	"OS::Kind":    String(runtime.GOOS),
	"OS::Arch":    String(runtime.GOARCH),
	"OS::Is64Bit": ToBool(uint64(^uintptr(0)) == ^uint64(0)),
	"OS::Stdin":   io.Reader(os.Stdin),
	"OS::Stdout":  io.Writer(os.Stderr),
	"OS::Stderr":  io.Writer(os.Stderr),
}

var OS_Functions = map[string] Value {
	"exit": func(code uint8) rx.Effect {
		return rx.CreateBlockingEffect(func(_ func(rx.Object)) error {
			os.Exit(int(code))
			panic("process should have exited")
		})
	},
}
