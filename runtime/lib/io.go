package lib

import (
	"fmt"
	"io"
	. "kumachan/runtime/common"
	"kumachan/runtime/common/rx"
)


var IO_Functions = map[string] Value {
	"write-line": func(out io.Writer, line []rune) rx.Effect {
		return rx.CreateEffect(func(sender rx.Sender) {
			var _, err = fmt.Fprintln(out, string(line))
			if err != nil {
				sender.Error(err)
			} else {
				sender.Next(nil)
				sender.Complete()
			}
		})
	},
}
