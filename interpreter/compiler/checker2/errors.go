package checker2

import (
	"fmt"
	"strconv"
	"kumachan/standalone/util/richtext"
)


func makeErrorDescBlankBlock() richtext.Block {
	var b richtext.Block
	b.WriteSpan("Error: ", richtext.TAG_ERR_EM)
	return b
}
func makeErrorDescBlock(msg string) richtext.Block {
	var b = makeErrorDescBlankBlock()
	b.WriteSpan(msg, richtext.TAG_ERR_NORMAL)
	return b
}

type E_InvalidMetadata struct {
	Reason  string
}
func (e E_InvalidMetadata) DescribeError() richtext.Block {
	var msg = fmt.Sprintf("invalid metadata: %s", e.Reason)
	return makeErrorDescBlock(msg)
}

type E_InvalidTypeName struct {
	Name  string
}
func (e E_InvalidTypeName) DescribeError() richtext.Block {
	var msg = fmt.Sprintf("invalid type name: %s", strconv.Quote(e.Name))
	return makeErrorDescBlock(msg)
}

type E_TypeParametersOnCaseType struct {}
func (e E_TypeParametersOnCaseType) DescribeError() richtext.Block {
	var msg = "cannot specify explicit type parameters on case types"
	return makeErrorDescBlock(msg)
}

