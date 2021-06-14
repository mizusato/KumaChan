package checker2

import (
	"fmt"
	"strconv"
	"kumachan/standalone/util/richtext"
)


const ErrorPrefix = "Error: "

type E_InvalidTypeName struct {
	Name  string
}
func (e E_InvalidTypeName) DescribeError() richtext.Block {
	var b richtext.Block
	var desc = fmt.Sprintf (
		"%s: invalid type name: %s",
		ErrorPrefix, strconv.Quote(e.Name),
	)
	b.WriteLine(desc, richtext.TAG_ERR_NORMAL)
	return b
}


