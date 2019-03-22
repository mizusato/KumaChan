package scanner


import "../syntax"


type Token struct {
    id       syntax.Item
    content  string
}


func Scan (code string) []Token {
    var tokens = make([]Token, 0, 10000)
    return tokens
}
