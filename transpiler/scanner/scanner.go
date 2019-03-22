package scanner


import "io"
import "fmt"
import "../syntax"


type Token struct {
    Id       syntax.Item
    Name     string
    Content  []rune
    Pos      int
}


type RuneListReader struct {
	src []rune
	pos int
}


func (r *RuneListReader) ReadRune() (rune, int, error) {
	if r.pos >= len(r.src) {
		return -1, 0, io.EOF
	}
	next := r.src[r.pos]
	r.pos += 1
	return next, 1, nil
}


func MatchToken (code []rune, pos int) (int, syntax.Item, string) {
    for _, token := range syntax.Tokens {
        reader := &RuneListReader { src: code, pos: pos }
        loc := token.Pattern.FindReaderIndex(reader)
        // fmt.Printf("Try %v\n", token.Name)
        if loc != nil {
            if (loc[0] != 0) { panic("invalid token pattern") }
            return loc[1], token.Id, token.Name
        }
    }
    return 0, 0, ""
}


func Scan (code []rune) []Token {
    var tokens = make([]Token, 0, 10000)
    var length = len(code)
    var pos = 0
    for pos < length {
        // fmt.Printf("pos %v\n", pos)
        amount, id, name := MatchToken(code, pos)
        if amount > 0 {
            tokens = append(tokens, Token {
                Id: id,
                Name: name,
                Pos: pos,
                Content: code[pos:pos+amount],
            })
            pos += amount
        } else {
            break
        }
    }
    if (pos < length) { panic(fmt.Sprintf("invalid token at %v", pos)) }
    return tokens
}
