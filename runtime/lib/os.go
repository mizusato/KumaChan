package lib

import (
	"io"
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
	. "kumachan/runtime/lib/container"
	"os"
	"runtime"
)


var OS_Constants = map[string] Value {
	"OS::Kind":    String(runtime.GOOS),
	"OS::Arch":    String(runtime.GOARCH),
	"OS::Is64Bit": ToBool(uint64(^uintptr(0)) == ^uint64(0)),
	"OS::Env":     GetEnv(),
	"OS::Args":    GetArgs(),
	"OS::Stdin":   io.Reader(os.Stdin),
	"OS::Stdout":  io.Writer(os.Stdout),
	"OS::Stderr":  io.Writer(os.Stderr),
}

func GetEnv() Map {
	var m = NewMap(func(v1 Value, v2 Value) Ordering {
		var s1 = v1.(String)
		var s2 = v2.(String)
		return StringCompare(s1, s2)
	})
	for _, item := range os.Environ() {
		var str = String(item)
		var k = make(String, 0)
		var v = make(String, 0)
		var cut = false
		for _, r := range str {
			if !cut && r == '=' {
				cut = true
				continue
			}
			if cut {
				v = append(v, r)
			} else {
				k = append(k, r)
			}
		}
		m = m.Insert(k, v)
	}
	return m
}

func GetArgs() ([] String) {
	// TODO: further process may be useful to extinguish interpreter arguments
	var args = make([] String, len(os.Args))
	for i, raw := range os.Args {
		args[i] = ([] rune)(raw)
	}
	return args
}

var OS_Functions = map[string] Value {
	"exit": func(code uint8) rx.Effect {
		return rx.CreateBlockingEffect(func() (rx.Object, bool) {
			os.Exit(int(code))
			panic("process should have exited")
		})
	},
}
