package error

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"kumachan/parser/ast"
	"kumachan/parser/scanner"
	"kumachan/transformer/node"
)


const ERR_FOV = 5

type E interface {
	Message()  ErrorMessage
}

type MaybeErrorPoint interface { MaybeErrorPoint() }
func (impl ErrorPoint) MaybeErrorPoint() {}
type ErrorPoint struct {
	AST   *ast.Tree
	Node  node.Node
}

func GetErrorTypeName(e interface{}) string {
	var T = reflect.TypeOf(e)
	return T.String()
}

func MsgFailedToCompile(cause interface{}, err []ErrorMessage) ErrorMessage {
	var err_type = GetErrorTypeName(cause)
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_ERROR, fmt.Sprintf (
		"*** Failed to Compile (%s)", err_type,
	))
	msg.WriteText(TS_NORMAL, "\n*\n")
	msg.WriteAll(JoinErrMsg(err, T(TS_NORMAL, "\n*\n")))
	return msg
}

func FormatError (
	code         scanner.Code,
	info         scanner.RowColInfo,
	span_map     scanner.RowSpanMap,
	file_name    string,
	coordinate   scanner.Point,
	spot         scanner.Span,
	fov          uint,
	highlight    TextStyle,
	description  ErrorMessage,
) ErrorMessage {
	var nh_rows = make([]scanner.Span, 0)
	var i = coordinate.Row
	var j = coordinate.Row
	for i > 1 && uint(coordinate.Row - i) < (fov/2) {
		i -= 1
	}
	var last_row int
	if len(code) > 0 {
		last_row = info[len(code)-1].Row
	} else {
		last_row = 1
	}
	for j < last_row && uint(j - coordinate.Row) < (fov/2) {
		j += 1
	}
	var start_row = i
	var end_row = j
	for r := start_row; r <= end_row; r += 1 {
		nh_rows = append(nh_rows, span_map[r])
	}
	var expected_width = len(strconv.Itoa(last_row))
	var align = func(num int) string {
		var num_str = strconv.Itoa(num)
		var num_width = len(num_str)
		var buf strings.Builder
		buf.WriteString(num_str)
		for i := num_width; i < expected_width; i += 1 {
			buf.WriteRune(' ')
		}
		return buf.String()
	}
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_INFO, "-----")
	msg.WriteInnerText(TS_INFO, fmt.Sprintf (
		"(row %d, column %d)",
		coordinate.Row, coordinate.Col,
	))
	msg.WriteText(TS_INFO, file_name)
	msg.Write(T_LF)
	var style = TS_NORMAL
	for i, row := range nh_rows {
		var current_row = (start_row + i)
		msg.WriteText(TS_NORMAL, fmt.Sprintf (
			" %s |", align(current_row),
		))
		msg.Write(T_SPACE)
		var buf strings.Builder
		for j, char := range code[row.Start: row.End] {
			var pos = (row.Start + j)
			if pos == spot.Start {
				msg.WriteBuffer(style, &buf)
				style = highlight
			}
			if pos == spot.End {
				msg.WriteBuffer(style, &buf)
				style = TS_NORMAL
			}
			buf.WriteRune(char)
		}
		if row.End == spot.Start {
			msg.WriteBuffer(style, &buf)
			style = highlight
		}
		if row.End == spot.End {
			msg.WriteBuffer(style, &buf)
			style = TS_NORMAL
		}
		msg.WriteBuffer(style, &buf)
		msg.Write(T_LF)
	}
	msg.WriteAll(description)
	return msg
}

func FormatErrorAt (
	point  ErrorPoint,
	desc   ErrorMessage,
) ErrorMessage {
	var AST = point.AST
	var Node = point.Node
	return FormatError (
		AST.Code,  AST.Info,    AST.SpanMap,
		AST.Name,  Node.Point,  Node.Span,
		ERR_FOV,   TS_SPOT,     desc,
	)
}
