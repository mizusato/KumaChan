package parser

import (
    "fmt"
    "strconv"
)
import "strings"
import "kumachan/parser/syntax"

const SiblingOffset = 2
const Bold = "\033[1m"
const Red = "\033[31m"
const Blue = "\033[34m"
const Reset = "\033[0m"


type Error struct {
    HasExpectedPart  bool
    ExpectedPart     syntax.Id
    NodeIndex        int  // may be bigger than the index of last token
}

func (err *Error) Message() string {
    if err.HasExpectedPart {
        return fmt.Sprintf (
            "Syntax unit '%v' expected",
            syntax.Id2Name[err.ExpectedPart],
        )
    } else {
        return "Parser stuck"
    }
}

func (err *Error) DetailedMessage(tree *Tree) string {
    var node = &tree.Nodes[err.NodeIndex]
    var token_index int
    var got string
    if node.Pos >= len(tree.Tokens) {
        token_index = len(tree.Tokens)-1
        got = "EOF"
    } else {
        token_index = node.Pos
        got = syntax.Id2Name[tree.Tokens[token_index].Id]
    }
    var token = &tree.Tokens[token_index]
    var token_span_size = token.Span.End - token.Span.Start
    for (token_index + 1) < len(tree.Tokens) && token_span_size == 0 {
        token_index += 1
        token = &tree.Tokens[token_index]
        token_span_size = token.Span.End - token.Span.Start
    }
    var point = tree.Info[token.Span.Start]
    var file = tree.Name
    var l, r = GetSiblingRange(tree, token_index)
    if tree.Code[l] == '\n' {
        l = (l + 1)
    }
    var spot = string(tree.Code[l:r])
    var lines = strings.Split(spot, "\n")
    var buf strings.Builder
    fmt.Fprintf(
        &buf, "%vFile:%v %v%v%v\n",
        Bold, Reset, Blue, file, Reset,
    )
    var base_row = point.Row - SiblingOffset
    if base_row < 1 {
        base_row = 1
    }
    var char_ptr = l
    var last_row = base_row + len(lines) - 1
    var expected_width = len(strconv.Itoa(last_row))
    var fill_zeros = func(num int) string {
        var num_str = strconv.Itoa(num)
        var num_width = len(num_str)
        var buf strings.Builder
        buf.WriteString(num_str)
        for i := num_width; i < expected_width; i += 1 {
            buf.WriteRune(' ')
        }
        return buf.String()
    }
    for i, line := range lines {
        var row = base_row + i
        fmt.Fprintf(&buf, "%s | ", fill_zeros(row))
        for _, char := range []rune(line) {
            if char_ptr == token.Span.Start && token_span_size > 0 {
                fmt.Fprintf(&buf, "%v%v", Red, Bold)
            }
            buf.WriteRune(char)
            if char_ptr == (token.Span.End - 1) && token_span_size > 0 {
                fmt.Fprintf(&buf, "%v", Reset)
            }
            char_ptr += 1
        }
        fmt.Fprintf(&buf, "\n")
        char_ptr += 1
    }
    fmt.Fprintf (
        &buf, "%v%v (got '%v') at (row %v, column %v) in %v%v%v",
        Red, err.Message(), got, point.Row, point.Col, Bold, file, Reset,
    )
    return buf.String()
}

func GetSiblingRange (tree *Tree, token_index int) (int, int) {
    var token = &tree.Tokens[token_index]
    var point = tree.Info[token.Span.Start]
    var left_bound = point.Row - SiblingOffset
    var right_bound = point.Row + SiblingOffset
    var p = token.Span.Start
    var l = token_index
    var L = p
    for l-1 >= 0 {
        var left_token = &tree.Tokens[l-1]
        var left_point = tree.Info[left_token.Span.Start]
        if left_point.Row >= left_bound {
            l -= 1
            L = left_token.Span.Start
        } else {
            break
        }
    }
    var r = token_index
    var R = token.Span.End
    for r+1 < len(tree.Tokens) {
        var right_token = &tree.Tokens[r+1]
        var right_point = tree.Info[right_token.Span.Start]
        if right_point.Row <= right_bound {
            r += 1
            R = right_token.Span.End
        } else {
            break
        }
    }
    return L, R
}

func InternalError (msg string) {
    panic(fmt.Sprintf("Internal Parser Error: %v", msg))
}