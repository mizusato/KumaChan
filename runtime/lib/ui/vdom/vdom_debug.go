package vdom

import (
	"io"
	"fmt"
	"strconv"
	"strings"
	"reflect"
)


type InspectContext struct {
	GetHandlerId  func(*EventHandler) (string, bool)
}

func Inspect(node *Node, ctx InspectContext) string {
	var buf strings.Builder
	writeNode(&buf, node, ctx, 0, 2)
	return buf.String()
}

func writeNode (
	buf io.Writer, node *Node, ctx InspectContext,
	depth uint, indent uint,
) {
	writeBlank(buf, depth, indent)
	fmt.Fprintf(buf, "<!-- %X -->", reflect.ValueOf(node).Pointer())
	writeLineFeed(buf)
	writeBlank(buf, depth, indent)
	writeStatic(buf, "<")
	writeString(buf, node.Tag)
	writeAttrs(buf, node.Attrs)
	writeEvents(buf, node.Events, ctx)
	writeStyles(buf, node.Styles)
	writeStatic(buf, ">")
	writeLineFeed(buf)
	switch content := node.Content.(type) {
	case *Text:
		writeQuotedString(buf, *content)
	case *Children:
		for _, child := range *content {
			writeNode(buf, child, ctx, (depth + 1), indent)
		}
	}
	writeBlank(buf, depth, indent)
	writeStatic(buf, "</")
	writeString(buf, node.Tag)
	writeStatic(buf, ">")
	writeLineFeed(buf)
}

func writeAttrs(buf io.Writer, attrs *Attrs) {
	if attrs == EmptyAttrs { return }
	attrs.Data.ForEach(func(key String, value interface{}) {
		writeStatic(buf, " ")
		writeString(buf, key)
		writeStatic(buf, "=")
		writeQuotedString(buf, value.(String))
	})
}

func writeStyles(buf io.Writer, styles *Styles) {
	if styles == EmptyStyles { return }
	writeStatic(buf, " style=\"")
	styles.Data.ForEach(func(name String, value interface{}) {
		writeString(buf, name)
		writeStatic(buf, ":")
		writeString(buf, value.(String))
		writeStatic(buf, ";")
	})
	writeStatic(buf, "\"")
}

func writeEvents(buf io.Writer, events *Events, ctx InspectContext) {
	if events == EmptyEvents { return }
	events.Data.ForEach(func(name String, opts_ interface{}) {
		var opts = opts_.(*EventOptions)
		writeStatic(buf, " ")
		writeStatic(buf, "@")
		writeString(buf, name)
		if opts.Capture {
			writeStatic(buf, ".capture")
		}
		if opts.Prevent {
			writeStatic(buf, ".prevent")
		}
		if opts.Stop {
			writeStatic(buf, ".stop")
		}
		writeStatic(buf, "=\"")
		var id, exists = ctx.GetHandlerId(opts.Handler)
		if exists {
			fmt.Fprintf(buf, "%s", id)
		} else {
			fmt.Fprintf(buf, "(invalid)")
		}
		writeStatic(buf, "\"")
	})
}

func writeStatic(buf io.Writer, content string) {
	fmt.Fprintf(buf, "%s", content)
}

func writeString(buf io.Writer, str String) {
	fmt.Fprintf(buf, "%s", string(str))
}

func writeQuotedString(buf io.Writer, str String) {
	fmt.Fprintf(buf, "%s", strconv.Quote(string(str)))
}

func writeBlank(buf io.Writer, n uint, chunk_size uint) {
	for i := uint(0); i < n; i += 1 {
		for j := uint(0); j < chunk_size; j += 1 {
			fmt.Fprintf(buf, " ")
		}
	}
}

func writeLineFeed(buf io.Writer) {
	fmt.Fprintf(buf, "\n")
}

