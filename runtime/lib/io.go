package lib

import (
	"io"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
)


var IO_Functions = map[string] Value {
	"String from I/O::Error": func(err error) String {
		return String(err.Error())
	},
	"is-eof": func(err error) SumValue {
		return ToBool(err == io.EOF)
	},
	"is-closed-pipe": func(err error) SumValue {
		return ToBool(err == io.ErrClosedPipe)
	},
}
