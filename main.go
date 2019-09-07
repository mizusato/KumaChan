package main

import "os"
import "fmt"
import "io"
import "io/ioutil"
import "./parser/syntax"
import "./parser/scanner"
import "./parser"
import "./transformer"

func check (err error) {
    if (err != nil) {
        panic(err)
    }
}

func parser_debug (file io.Reader, name string, root string) {
    var code_bytes, err = ioutil.ReadAll(file)
    check(err)
    var code_string = string(code_bytes)
    var code = []rune(code_string)
    var tokens, info, semi = scanner.Scan(code)
    fmt.Println("Tokens:")
    for i, token := range tokens {
        fmt.Printf(
            "(%v) at [%v, %v] %v: %v\n",
            i, info[token.Pos].Row, info[token.Pos].Col,
            syntax.Id2Name[token.Id],
            string(token.Content),
        )
    }
    var RootId, exists = syntax.Name2Id[root]
    if !exists {
        panic("invalid root syntax unit " + root)
    }
    var nodes, err_ptr, err_desc = parser.BuildBareTree(RootId, tokens)
    fmt.Println("------------------------------------------------------")
    fmt.Println("AST Nodes:")
    parser.PrintBareTree(nodes)
    var tree = parser.Tree {
        Code: code, Tokens: tokens, Info: info, Semi: semi,
        Nodes: nodes, File: name,
    }
    fmt.Println("------------------------------------------------------")
    fmt.Println("AST:")
    parser.PrintTree(tree)
    if err_ptr != -1 {
        parser.Error(&tree, err_ptr, err_desc)
    }
}

func main () {
    syntax.Init()
    transformer.Init()
    parser_debug(os.Stdin, "<eval>", "eval")
}
