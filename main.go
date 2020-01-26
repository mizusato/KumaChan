package main

import (
    "kumachan/loader"
    "kumachan/transformer"
    "os"
    "reflect"
)
import "fmt"
import "io"
import "io/ioutil"
import "kumachan/parser/syntax"
import "kumachan/parser/scanner"
import "kumachan/parser"

func check (err error) {
    if (err != nil) {
        panic(err)
    }
}

func parser_debug (file io.Reader, name string, root string) {
    var code_bytes, e = ioutil.ReadAll(file)
    check(e)
    var code_string = string(code_bytes)
    var code = []rune(code_string)
    var tokens, info = scanner.Scan(code)
    fmt.Println("Tokens:")
    for i, token := range tokens {
        fmt.Printf(
            "(%v) at [%v, %v](%v, %v) %v: %v\n",
            i,
            token.Span.Start,
            token.Span.End,
            info[token.Span.Start].Row,
            info[token.Span.Start].Col,
            syntax.Id2Name[token.Id],
            string(token.Content),
        )
    }
    var RootId, exists = syntax.Name2Id[root]
    if !exists {
        panic("invalid root syntax unit " + root)
    }
    var nodes, err = parser.BuildTree(RootId, tokens)
    fmt.Println("------------------------------------------------------")
    fmt.Println("AST Nodes:")
    parser.PrintBareTree(nodes)
    var tree = parser.Tree {
        Nodes: nodes, Name: name,
        Code: code, Tokens: tokens, Info: info,
    }
    fmt.Println("------------------------------------------------------")
    fmt.Println("AST:")
    parser.PrintTree(&tree)
    if err != nil {
        fmt.Println(err.DetailedMessage(&tree))
    } else {
        fmt.Println("------------------------------------------------------")
        fmt.Println("Transformed:")
        transformer.PrintNode(reflect.ValueOf(transformer.Transform(&tree)))
        // fmt.Printf("%+v\n", transformer.Transform(&tree))
    }
}

func loader_debug() {
    if len(os.Args) != 2 {
        panic("invalid arguments")
    }
    var path = os.Args[1]
    var _, err = loader.LoadModule(path, loader.ErrorContext { ImportPoint: nil })
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
    } else {
        fmt.Println("Modules loaded, no errors.")
    }
}

func main () {
    syntax.Init()
    // parser_debug(os.Stdin, "[eval]", "module")
    loader_debug()
}
