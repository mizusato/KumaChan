package atom

import (
	"io"
	"fmt"
	"bufio"
	"strings"
	"encoding/json"
	"kumachan/util"
)


type LangServerContext struct {
	DebugLog     func(info string)
}

func LangServer(input io.Reader, output io.Writer, debug io.Writer) error {
	input = bufio.NewReader(input)
	var write_line = func(line ([] byte)) error {
		_, err := output.Write(line)
		if err != nil { return err }
		_, err = output.Write(([]byte)("\n"))
		if err != nil { return err }
		return nil
	}
	var ctx = LangServerContext{
		DebugLog: func(info string) { _, _ = fmt.Fprintln(debug, info) },
	}
	for {
		var line_runes, err = util.WellBehavedReadLine(input)
		if err != nil { return err }
		var line = string(line_runes)
		var i = strings.Index(line, " ")
		if i == -1 { i = len(line) }
		var cmd = line[:i]
		var arg = strings.TrimPrefix(line[i:], " ")
		switch cmd {
		case "quit":
			return nil
		case "lint":
			var raw_req = ([] byte)(arg)
			var req LintRequest
			err := json.Unmarshal(raw_req, &req)
			if err != nil { return err }
			var res = Lint(req, ctx)
			raw_res, err := json.Marshal(&res)
			if err != nil { return err }
			err = write_line(raw_res)
			if err != nil { return err }
		case "autocomplete":
			var raw_req = ([] byte)(arg)
			var req AutoCompleteRequest
			err := json.Unmarshal(raw_req, &req)
			if err != nil { return err }
			var res = AutoComplete(req, ctx)
			raw_res, err := json.Marshal(&res)
			if err != nil { return err }
			err = write_line(raw_res)
			if err != nil { return err }
		}
	}
}

