package parser

import "os"
import "fmt"


func InternalError (msg string) {
    panic(fmt.Sprintf("Internal Parser Error: %v", msg))
}


func Error (tree *Tree, ptr int, msg string) {
    var node = &tree.Nodes[ptr]
    var point = tree.Info[node.Pos]
    var file = tree.File
    fmt.Fprintf (
        os.Stderr, "%v at %v (row %v, column %v)\n",
        msg, file, point.Row, point.Col,
    )
    os.Exit(1)
}
