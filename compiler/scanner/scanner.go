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
    Row int
    Col int
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
        info = append(info, Point { Row: row, Col: col })
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


func IsRightParOrName (token *Token) bool {
    if token != nil {
        var name = syntax.Id2Name[token.Id]
        if name == ")" || name == "]" || name == ">" || name == "Name" {
            return true
        } else {
            return false
        }
    } else {
        return false
    }
}


func IsReturnKeyword (token *Token) bool {
    var NameId = syntax.Name2Id["Name"]
    if token != nil {
        return (token.Id == NameId && string(token.Content) == "return")
    } else {
        return false
    }
}

func Try2InsertExtra (tokens TokenSequence, current Token) TokenSequence {
    /*
     *  A blank magic to distinguish
     *      let t = f(g*h)(x)
     *  between
     *      let t = f
     *      (g*h)(x)
     */
    var current_name = syntax.Id2Name[current.Id]
    if current_name == "(" || current_name == "<" {
        return append(tokens, Token {
            Id:       syntax.Name2Id["Call"],
            Pos:      current.Pos,
            Content:  []rune(""),
        })
    } else if current_name == "[" || current_name == "." {
        return append(tokens, Token {
            Id:       syntax.Name2Id["Get"],
            Pos:      current.Pos,
            Content:  []rune(""),
        })
    }
    return tokens
}


func Scan (code Code) (TokenSequence, RowColInfo) {
    var BlankId = syntax.Name2Id["Blank"]
    var CommentId = syntax.Name2Id["Comment"]
    var LFId = syntax.Name2Id["LF"]
    var RCBId = syntax.Name2Id["}"]
    var tokens = make(TokenSequence, 0, 10000)
    var info = GetInfo(code)
    var length = len(code)
    var previous_ptr *Token
    var pos = 0
    for pos < length {
        // fmt.Printf("pos %v\n", pos)
        amount, id := MatchToken(code, pos)
        if amount == 0 { break }
        if id == BlankId { pos += amount; continue }
        if id == CommentId { pos += amount; continue }
        var current = Token {
            Id: id,
            Pos: pos,
            Content: code[pos : pos+amount],
        }
        if IsRightParOrName(previous_ptr) {
            tokens = Try2InsertExtra(tokens, current)
        }
        if current.Id == LFId || current.Id == RCBId {
            /* tell from "return [LF] expr" and "return expr" */
            if IsReturnKeyword(previous_ptr) {
                tokens = append(tokens, Token {
                    Id:       syntax.Name2Id["Void"],
                    Pos:      current.Pos,
                    Content:  []rune(""),
                })
            }
        }
        tokens = append(tokens, current)
        previous_ptr = &current
        pos += amount
    }
    var clear = make(TokenSequence, 0, 10000)
    for _, token := range tokens {
        if token.Id != LFId {
            clear = append(clear, token)
        }
    }
    if (pos < length) {
        panic(fmt.Sprintf("invalid token at %+v", info[pos]))
    }
    return clear, info
}
