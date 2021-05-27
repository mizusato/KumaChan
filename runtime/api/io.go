package api

import (
	"io"
	. "kumachan/lang"
)


var IO_Functions = map[string] Value {
	"is-eof": func(err error) EnumValue {
		return ToBool(err == io.EOF)
	},
	"is-closed-pipe": func(err error) EnumValue {
		return ToBool(err == io.ErrClosedPipe)
	},
}
