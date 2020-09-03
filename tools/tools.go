package tools

import (
	"io"
	"strings"
	"encoding/json"
	"kumachan/runtime/rx"
)


type DirtyBuffer struct {
	Path  string   `json:"path"`
	Text  string   `json:"text"`
}

func Server(input io.Reader, output io.Writer) error {
	for {
		var line_runes, err = rx.WellBehavedScanLine(input)
		if err != nil { return err }
		var line = string(line_runes)
		var i = strings.Index(line, " ")
		if i == -1 { i = len(line) }
		var cmd = line[:i]
		var arg = strings.TrimPrefix(line[i:], " ")
		switch cmd {
		case "quit":
			return nil
		case "autocomplete":
			var req AutoCompleteRequest
			err := json.Unmarshal(([]byte)(arg), &req)
			if err != nil { return err }
			var res = AutoComplete(req)
			raw_res, err := json.Marshal(&res)
			if err != nil { return err }
			_, err = output.Write(raw_res)
			if err != nil { return err }
			_, err = output.Write(([]byte)("\n"))
			if err != nil { return err }
		}
	}
}
