package main;


import "fmt"
import "./scanner"


func main () {
    var tokens = scanner.Scan([]rune(`'123' /* abc */`))
    for _, token := range tokens {
        fmt.Printf(
            "#%v [%v:%v]: %v\n",
            token.Pos, token.Id, token.Name, string(token.Content),
        )
    }
    fmt.Println("change the world")
}
