package main;


import "os"
import "fmt"
import "io/ioutil"
import "./syntax"
import "./parser"


func check (err error) {
    if (err != nil) {
        panic(err)
    }
}


func test () {
    syntax.Init()
    var code_bytes, err = ioutil.ReadAll(os.Stdin)
    check(err)
    var code_string = string(code_bytes)
    var code = []rune(code_string)
    var tree = parser.BuildTree(code)
    fmt.Printf("\n")
    for _, token := range tree.Tokens {
        fmt.Printf(
            "#%+v [%v:%v]: %v\n",
            tree.Info[token.Pos], token.Id, syntax.Id2Name[token.Id],
            string(token.Content),
        )
    }
    fmt.Printf("\n")
    parser.PrintRawTree(tree.Nodes)
    fmt.Printf("\n")
    parser.PrintTree(tree)
}


func main () {
    test()
}
