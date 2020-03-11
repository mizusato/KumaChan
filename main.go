package main

import (
    "os"
    "io"
    "fmt"
    "reflect"
    "io/ioutil"
    "kumachan/checker"
    "kumachan/loader"
    "kumachan/transformer"
    "kumachan/parser/syntax"
    "kumachan/parser/scanner"
    "kumachan/parser"
)


func check (err error) {
    if (err != nil) {
        panic(err)
    }
}

func debug_parser(file io.Reader, name string, root string) {
    var code_bytes, e = ioutil.ReadAll(file)
    check(e)
    var code_string = string(code_bytes)
    var code = []rune(code_string)
    var tokens, info, _ = scanner.Scan(code)
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
    var _, exists = syntax.Name2Id[root]
    if !exists {
        panic("invalid root syntax unit " + root)
    }
    var tree, err = parser.Parse(code, root, name)
    fmt.Println("------------------------------------------------------")
    fmt.Println("AST Nodes:")
    parser.PrintBareTree(tree.Nodes)
    fmt.Println("------------------------------------------------------")
    fmt.Println("AST:")
    parser.PrintTree(tree)
    if err != nil {
        fmt.Println(err.Message())
    } else {
        fmt.Println("------------------------------------------------------")
        fmt.Println("Transformed:")
        transformer.PrintNode(reflect.ValueOf(transformer.Transform(tree)))
        // fmt.Printf("%+v\n", transformer.Transform(tree))
    }
}

func debug_loader() (*loader.Module, loader.Index) {
    if len(os.Args) != 2 {
        panic("invalid arguments")
    }
    var path = os.Args[1]
    var mod, idx, err = loader.LoadEntry(path)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
        os.Exit(127)
    } else {
        fmt.Println("Modules loaded, no errors occurred.")
    }
    return mod, idx
}

func debug_checker(mod *loader.Module, idx loader.Index) {
    var _, err = checker.RegisterTypes(mod, idx)
    if err != nil {
        fmt.Fprintf(os.Stderr, "%s\n", err.Error())
        os.Exit(126)
    } else {
        fmt.Println("Types registered, no errors occurred.")
    }
}

func main () {
    // debug_parser(os.Stdin, "[eval]", "module")
    var mod, idx = debug_loader()
    debug_checker(mod, idx)
}
