package error

import (
	"fmt"
	"kumachan/parser"
	"kumachan/transformer/node"
	"strings"
)

const ERR_MSG_ROW_DELTA = 2
const Bold = "\033[1m"
const Red = "\033[31m"
const Blue = "\033[34m"
const Reset = "\033[0m"


type MaybeErrorPoint interface { MaybeErrorPoint() }
func (impl ErrorPoint) MaybeErrorPoint() {}
type ErrorPoint struct {
	AST   *parser.Tree
	Node  node.Node
}

func (point ErrorPoint) GenErrMsg(description string) string {
	var code = point.AST.Code
	var file = point.AST.Name
	var coor = point.Node.Point
	var delta = ERR_MSG_ROW_DELTA
	var start, end = __GetErrorPointSiblingRange(point, delta)
	var span = point.Node.Span
	var spot_left = string(code[start: span.Start])
	var spot = string(code[span.Start: span.End])
	var spot_right = string(code[span.End: end])
	var whole = string(code[start: end])
	var lines = strings.Split(whole, "\n")
	var t = 0
	var spot_line = 0
	for i, line := range lines {
		t += (len(line) + 1)
		if t > len(spot_left) {
			spot_line = i
			break
		}
	}
	var highlighted = fmt.Sprintf (
		"%s%v%s%v%s",
		spot_left, Bold+Red, spot, Reset, spot_right,
	)
	var highlighted_lines = strings.Split(highlighted, "\n")
	var buf strings.Builder
	fmt.Fprintf (
		&buf, "%vFile:%v %v%s%v\n",
		Bold, Reset, Blue, file, Reset,
	)
	for i, line := range highlighted_lines  {
		var line_number = coor.Row + (i - spot_line)
		fmt.Fprintf(&buf, "%d | %s\n", line_number, line)
	}
	fmt.Fprintf (
		&buf,
		"%s %vat (row %d, column %d) in %s%v\n",
		description, Red, coor.Row, coor.Col, file, Reset,
	)
	return buf.String()
}

func __GetErrorPointSiblingRange(point ErrorPoint, row_delta int) (int, int) {
	var span = point.Node.Span
	var code = point.AST.Code
	var move_cursor = func (
		initial    int,
		step       int,
		milestone  rune,
		limit      int,
	) int {
		if initial == len(code) {
			return initial
		}
		var cur = initial
		var d = 0
		for {
			if code[cur] == milestone {
				d += 1
				if d == limit {
					break
				}
			}
			var next = cur + step
			if next >= 0 && next < len(code) {
				cur = next
			} else {
				break
			}
		}
		return cur
	}
	var l = move_cursor(span.Start, -1, '\n', (row_delta + 1))
	var r = move_cursor(span.End, 1, '\n', row_delta)
	if code[l] == '\n' {
		l = (l + 1)
	}
	return l, r
}
