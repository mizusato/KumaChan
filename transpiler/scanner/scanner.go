package scanner


import "io"
import "fmt"
import "../syntax"


type Code = []rune


type Token struct {
    Id       syntax.Id
    Content  []rune
    Pos      int
}

type TokenSequence = []Token


type Point struct {
    row int
    col int
}

type RowColInfo = []Point


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


func GetInfo (code Code) RowColInfo {
    var info = make(RowColInfo, 0, 10000)
    var row = 1
    var col = 0
    for _, char := range code {
        if char != '\n' {
            col += 1
        } else {
            row += 1
            col = 0
        }
        info = append(info, Point { row: row, col: col })
    }
    return info
}


func MatchToken (code Code, pos int) (amount int, id syntax.Id) {
    for _, token := range syntax.Tokens {
        reader := &RuneListReader { src: code, pos: pos }
        loc := token.Pattern.FindReaderIndex(reader)
        // fmt.Printf("Try %v\n", token.Name)
        if loc != nil {
            if (loc[0] != 0) { panic("invalid token pattern") }
            return loc[1], syntax.Name2Id[token.Name]
        }
    }
    return 0, 0
}


func Scan (code Code) (TokenSequence, RowColInfo) {
    var BlankId = syntax.Name2Id["Blank"]
    var tokens = make(TokenSequence, 0, 10000)
    var info = GetInfo(code)
    var length = len(code)
    var pos = 0
    for pos < length {
        // fmt.Printf("pos %v\n", pos)
        amount, id := MatchToken(code, pos)
        if amount == 0 { break }
        if id == BlankId { pos += amount; continue }
        tokens = append(tokens, Token {
            Id: id,
            Pos: pos,
            Content: code[pos:pos+amount],
        })
        pos += amount
    }
    if (pos < length) {
        panic(fmt.Sprintf("invalid token at %+v", info[pos]))
    }
    return tokens, info
}
