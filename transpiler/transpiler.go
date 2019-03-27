package main;


import "os"
import "fmt"
import "io/ioutil"
import "./syntax"
import "./scanner"


func check (err error) {
    if (err != nil) {
        panic(err)
    }
}


func main () {
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
    fmt.Println("change the world")
}
