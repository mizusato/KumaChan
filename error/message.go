package error

import (
	"fmt"
	"strings"
)


type ErrorMessage  [] StyledText

type TextStyle int
const (
	TS_NORMAL  TextStyle  =  iota
	TS_BOLD
	TS_ERROR
	TS_WARNING
	TS_INFO
	TS_SPOT
	TS_INLINE
	TS_INLINE_CODE
)

type StyledText struct {
	Style  TextStyle
	Text   string
}
func T(style TextStyle, text string) StyledText {
	return StyledText { style, text }
}
var T_LF = StyledText { Style: TS_NORMAL, Text:  "\n", }
var T_SPACE = StyledText { Style: TS_NORMAL, Text:  " ", }
var T_INDENT = StyledText { Style: TS_NORMAL, Text:  "    ", }


func (t StyledText) String() string {
	const Bold = "\033[1m"
	const Red = "\033[31m"
	const Green = "\033[32m"
	const Orange = "\033[33m"
	const Blue = "\033[34m"
	const Magenta = "\033[35m"
	const Cyan = "\033[36m"
	const Reset = "\033[0m"
	switch t.Style {
	case TS_NORMAL:
		return t.Text
	case TS_BOLD:
		return Bold + t.Text + Reset
	case TS_ERROR:
		return Bold + t.Text + Reset
	case TS_WARNING:
		return Bold + t.Text + Reset
	case TS_INFO:
		return Orange + Bold + t.Text + Reset
	case TS_SPOT:
		return Red + Bold + t.Text + Reset
	case TS_INLINE:
		return Red + Bold + t.Text + Reset
	case TS_INLINE_CODE:
		return Magenta + Bold + t.Text + Reset
	default:
		return t.Text
	}
}

func (t StyledText) StringMarkdown() string {
	switch t.Style {
	case TS_NORMAL:
		return t.Text
	case TS_BOLD:
		return fmt.Sprintf("**%s**", t.Text)
	case TS_ERROR:
		return fmt.Sprintf("**%s**", t.Text)
	case TS_WARNING:
		return fmt.Sprintf("**%s**", t.Text)
	case TS_INFO:
		return fmt.Sprintf("**<font color='orange'>%s</font>**", t.Text)
	case TS_SPOT:
		return fmt.Sprintf("**<font color='red'>%s</font>**", t.Text)
	case TS_INLINE:
		return fmt.Sprintf("**<font color='red'>%s</font>**", t.Text)
	case TS_INLINE_CODE:
		return fmt.Sprintf("**<font color='magenta'>%s</font>**", t.Text)
	default:
		return t.Text
	}
}

func (t StyledText) StringPlain() string {
	return t.Text
}


func (msg ErrorMessage) String() string {
	var buf strings.Builder
	for _, segment := range msg {
		buf.WriteString(segment.String())
	}
	return buf.String()
}

func (msg ErrorMessage) StringMarkdown() string {
	var buf strings.Builder
	for _, segment := range msg {
		buf.WriteString(segment.StringMarkdown())
	}
	return buf.String()
}

func (msg ErrorMessage) StringPlain() string {
	var buf strings.Builder
	for _, segment := range msg {
		buf.WriteString(segment.StringPlain())
	}
	return buf.String()
}

func (msg *ErrorMessage) Write(text StyledText) {
	*msg = append(*msg, text)
}

func (msg *ErrorMessage) WriteText(style TextStyle, text string) {
	msg.Write(T(style, text))
}

func (msg *ErrorMessage) WriteInnerText(style TextStyle, text string) {
	msg.Write(T_SPACE)
	msg.WriteText(style, text)
	msg.Write(T_SPACE)
}

func (msg *ErrorMessage) WriteEndText(style TextStyle, text string) {
	msg.Write(T_SPACE)
	msg.WriteText(style, text)
}

func (msg *ErrorMessage) WriteBuffer(style TextStyle, buf *strings.Builder) {
	var text = buf.String()
	buf.Reset()
	if len(text) > 0 {
		msg.WriteText(style, text)
	}
}

func (msg *ErrorMessage) WriteAll(another ErrorMessage) {
	*msg = append(*msg, another...)
}

func (msg *ErrorMessage) WriteAllWithIndent(another ErrorMessage, amount uint) {
	msg.WriteRepeated(T_INDENT, amount)
	for i, item := range another {
		msg.Write(item)
		if item == T_LF && i != len(another)-1 {
			msg.WriteRepeated(T_INDENT, amount)
		}
	}
}

func (msg *ErrorMessage) WriteRepeated(text StyledText, amount uint) {
	for i := uint(0); i < amount; i += 1 {
		msg.Write(text)
	}
}

func JoinErrMsg(messages ([] ErrorMessage), separator StyledText) ErrorMessage {
	var joined = make(ErrorMessage, 0, len(messages))
	for i, item := range messages {
		joined.WriteAll(item)
		if i != len(messages)-1 {
			joined.Write(separator)
		}
	}
	return joined
}

func ListErrMsgItems(items ([] ErrorMessage), prefix string) ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, fmt.Sprintf("%s {", prefix))
	if len(items) > 0 {
		msg.Write(T_LF)
	}
	for i, item := range items {
		msg.WriteAllWithIndent(item, 1)
		if i != len(items)-1 {
			msg.WriteText(TS_NORMAL, ",")
		}
		msg.Write(T_LF)
	}
	msg.WriteText(TS_NORMAL, "}")
	return msg
}
