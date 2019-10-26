package node

import "kumachan/parser/scanner"


type Node struct {
    Point  scanner.Point
    Span   scanner.Span
    Info   interface{}
}
