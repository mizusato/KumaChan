package parser

import "os"
import "fmt"
import "strings"

const SiblingOffset = 2
const Bold = "\033[1m"
const Red = "\033[31m"
const Blue = "\033[34m"
const Reset = "\033[0m"


func InternalError (msg string) {
    panic(fmt.Sprintf("Internal Parser Error: %v", msg))
}


func Error (tree *Tree, ptr int, msg string) {
    var node = &tree.Nodes[ptr]
    var p int
    if node.Pos >= len(tree.Tokens) {
        p = len(tree.Tokens)-1
    } else {
        p = node.Pos
    }
    var token = &tree.Tokens[p]
    var point = tree.Info[token.Pos]
    var file = tree.File
    var l, r = GetSiblingRange(tree, p)
    /*
    fmt.Fprintf(os.Stderr, "[Debug] (%v, %v)\n", l, r)
    for i := range tree.Code {
        fmt.Fprintf(os.Stderr, "[Debug] %v: %v\n", i, string([]rune{tree.Code[i]}))
    }
    */
    var spot = string(tree.Code[l:r])
    var lines = strings.Split(spot, "\n")
    fmt.Fprintf (
        os.Stderr, "%vFile:%v %v%v%v\n",
        Bold, Reset, Blue, file, Reset,
    )
    var cp = l
    for i, line := range lines {
        var row = point.Row - SiblingOffset + i
        if row < 1 {
            row = 1
        }
        fmt.Fprintf(os.Stderr, "%v | ", row)
        for _, char := range []rune(line) {
            if cp == token.Pos {
                fmt.Fprintf(os.Stderr, "%v%v", Red, Bold)
            }
            fmt.Fprintf(os.Stderr, "%v", string([]rune{char}))
            if cp == token.Pos {
                fmt.Fprintf(os.Stderr, "%v", Reset)
            }
            cp += 1
        }
        fmt.Fprintf(os.Stderr, "\n")
        cp += 1
    }
    fmt.Fprintf (
        os.Stderr, "%v%v at (row %v, column %v) in %v%v\n",
        Red, msg, point.Row, point.Col, file, Reset,
    )
    os.Exit(1)
}


func GetSiblingRange (tree *Tree, token_ptr int) (int, int) {
    var token = &tree.Tokens[token_ptr]
    var point = tree.Info[token.Pos]
    var left_bound = point.Row - SiblingOffset
    var right_bound = point.Row + SiblingOffset
    var p = token.Pos
    var l = token_ptr
    var L = p
    for {
        if l-1 >= 0 {
            var left_token = &tree.Tokens[l-1]
            var left_point = tree.Info[left_token.Pos]
            if left_point.Row >= left_bound {
                l -= 1
                L = left_token.Pos
            } else {
                break
            }
        } else {
            break
        }
    }
    var r = token_ptr
    var R = p + len(token.Content)
    for {
        if r+1 < len(tree.Tokens) {
            var right_token = &tree.Tokens[r+1]
            var right_point = tree.Info[right_token.Pos]
            if right_point.Row <= right_bound {
                r += 1
                R = right_token.Pos + len(right_token.Content)
            } else {
                break
            }
        } else {
            break
        }
    }
    return L, R
}
