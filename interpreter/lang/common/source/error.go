package source

import (
	"fmt"
	"strings"
	"kumachan/standalone/util/richtext"
)


type Errors ([] Error)
func (es Errors) Error() string {
	if len(es) == 0 {
		panic("invalid operation")
	}
	var all = make([] string, len(es))
	for i, e := range es {
		all[i] = e.Error()
	}
	return strings.Join(all, "\n")
}

type Error struct {
	Location  Location
	Content   ErrorContent
}
func MakeError(loc Location, content ErrorContent) *Error {
	return &Error {
		Location: loc,
		Content:  content,
	}
}
type ErrorContent interface {
	DescribeError() richtext.Block
}
func (e *Error) Error() string {
	var msg = ErrorMessage {
		Description: e.Content.DescribeError(),
		Location:    e.Location,
	}
	return msg.ToFullText().RenderLinear(richtext.RenderOptionsLinear {
		UseAnsiColor: false,
	})
}

type ErrorMessage struct {
	Description  richtext.Block
	Location     Location
}
func (msg ErrorMessage) ToFullText() richtext.Text {
	var t richtext.Text
	var header_block richtext.Block
	var header = fmt.Sprintf("--- %s at %s", msg.Location.PosDesc(), msg.Location.FilePath())
	header_block.WriteLine(header, richtext.TAG_ERR_EM)
	t.Write(header_block)
	t.WriteText(msg.Location.FormatMessage(msg.Description))
	return t
}
func (msg ErrorMessage) ToSerializable() SerializableErrorMessage {
	return SerializableErrorMessage {
		Description: msg.Description,
		FilePath:    msg.Location.File.GetPath(),
		Position:    msg.Location.File.DescribePosition(msg.Location.Pos),
		TokenIndex:  int64(msg.Location.Pos.StartTokenIndex),
	}
}

type SerializableErrorMessage struct {
	Description  richtext.Block  `kmd:"description"`
	FilePath     string          `kmd:"file-path"`
	Position     string          `kmd:"position"`
	TokenIndex   int64           `kmd:"token-index"`
}

