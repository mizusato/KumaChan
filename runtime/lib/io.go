package lib

import (
	"io"
	. "kumachan/runtime/common"
)


var IO_Functions = map[string] Value {
	"is-eof": func(err error) SumValue {
		return ToBool(err == io.EOF)
	},
	"is-closed-pipe": func(err error) SumValue {
		return ToBool(err == io.ErrClosedPipe)
	},
}
