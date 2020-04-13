package lib

import (
	"fmt"
	"io"
	"kumachan/runtime/common/rx"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
)


var IO_Functions = map[string] Value {
	"write-line": func(out io.Writer, line String) rx.Effect {
		return rx.CreateEffect(func(sender rx.Sender) {
			var _, err = fmt.Fprintln(out, string(line))  // TODO: improve
			if err != nil {
				sender.Error(err)
			} else {
				sender.Next(nil)
				sender.Complete()
			}
		})
	},
}
