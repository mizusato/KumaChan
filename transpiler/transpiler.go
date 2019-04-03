package main;


import "os"
import "fmt"
import "strings"
import "strconv"
import "io/ioutil"
import "./syntax"
import "./scanner"
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
    var tokens, info = scanner.Scan(code)
    for _, token := range tokens {
        fmt.Printf(
            "#%+v [%v:%v]: %v\n",
            info[token.Pos], token.Id, syntax.Id2Name[token.Id],
            string(token.Content),
        )
    }
    fmt.Printf("\n")
    var raw_tree = parser.BuildRawTree(tokens)
    fmt.Printf("\n")
    for i, node := range raw_tree {
        var children = make([]string, 0, 20)
        for i := 0; i < node.Length; i++ {
            children = append(children, strconv.Itoa(node.Children[i]))
        }
        var children_str = strings.Join(children, ",")
        fmt.Printf(
            "#%v %v [%v] pos=%+v, amount=%v\n",
            i, syntax.Id2Name[node.Part.Id], children_str,
            info[node.Pos], node.Amount,
        )
    }
}


func main () {
    test()
}
