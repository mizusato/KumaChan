package main


import "os"
import "fmt"
import "io/ioutil"
import "./syntax"
import "./scanner"
import "./parser"
import "./transpiler"


func check (err error) {
    if (err != nil) {
        panic(err)
    }
}


func test () {
    syntax.Init()
    transpiler.Init()
    var code_bytes, err = ioutil.ReadAll(os.Stdin)
    check(err)
    var code_string = string(code_bytes)
    var code = []rune(code_string)
    var tokens, info = scanner.Scan(code)
    fmt.Printf("\n")
    for i, token := range tokens {
        fmt.Printf(
            "#%v %+v [%v:%v]: %v\n",
            i, info[token.Pos], token.Id, syntax.Id2Name[token.Id],
            string(token.Content),
        )
    }
    fmt.Printf("\n")
    var tree = parser.BuildTree(code, "<eval>")
    fmt.Printf("\n")
    parser.PrintBareTree(tree.Nodes)
    fmt.Printf("\n")
    parser.PrintTree(tree)
    var js = transpiler.Transpile(&tree, 0)
    fmt.Println(js)
}


func main () {
    test()
}
