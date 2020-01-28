package scanner


import "io"
import "fmt"
import "kumachan/parser/syntax"


type Code = []rune

type Span struct {
    Start  int
    End    int
}

func (span Span) Merged(another Span) Span {
    var merged = Span { Start: span.Start, End: another.End }
    if !(merged.Start <= merged.End) {
        panic(fmt.Sprintf("invalid span merge: %+v and %+v", span, another))
    }
    return merged
}

type Token struct {
    Id       syntax.Id
    Span     Span
    Content  []rune
}

type Tokens = []Token


type Point struct {
    Row  int
    Col  int
}

type RowColInfo = []Point

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

func MatchToken (code Code, pos int) (amount int, id syntax.Id) {
    for _, token := range syntax.Tokens {
        var reader = RuneListReader { src: code, pos: pos }
        var loc = token.Pattern.FindReaderIndex(&reader)
        // fmt.Printf("Try %v\n", token.Name)
        if loc != nil {
            if (loc[0] != 0) { panic("invalid token definition " + token.Name) }
            return loc[1], syntax.Name2Id[token.Name]
        }
    }
    return 0, 0
}


func __IsLineFeed (token *Token) bool {
    if token != nil {
        return (token.Id == syntax.Name2Id["LF"])
    } else {
        return false
    }
}

func __IsReturnKeyword (token *Token) bool {
    if token != nil {
        return (token.Id == syntax.Name2Id["Name"] &&
            string(token.Content) == "return")
    } else {
        return false
    }
}

func __IsLeftCurlyBrace (token *Token) bool {
    if token != nil {
        return (token.Id == syntax.Name2Id["{"])
    } else {
        return false
    }
}


func Scan (code Code) (Tokens, RowColInfo) {
    var Comment = syntax.Name2Id["Comment"]
    var Blank = syntax.Name2Id["Blank"]
    var LF = syntax.Name2Id["LF"]
    var LeftParentheses = syntax.Name2Id["("]
    var LeftBracket = syntax.Name2Id["["]
    var RightCurlyBrace = syntax.Name2Id["}"]
    var tokens = make(Tokens, 0, 10000)
    var info = GetInfo(code)
    var length = len(code)
    var previous *Token
    var pos = 0
    for pos < length {
        var amount, id = MatchToken(code, pos)
        if amount == 0 { break }
        if id == Comment { pos += amount; continue }
        if id == Blank { pos += amount; continue }
        if id == LF && __IsLineFeed(previous) { pos += amount; continue }
        var span = Span { Start: pos, End: pos + amount }
        var zero_span = Span { Start: pos, End: pos }
        var current = Token {
            Id:      id,
            Span:    span,
            Content: code[span.Start : span.End],
        }
        // TODO: remove these obsoleted checks
        if current.Id == LF || current.Id == RightCurlyBrace {
            /* tell from "return [LF] expr" and "return expr" */
            if __IsReturnKeyword(previous) {
                tokens = append(tokens, Token {
                    Id:      syntax.Name2Id["Void"],
                    Span:    zero_span,
                    Content: []rune(""),
                })
            }
        }
        /* inject a LF if there is no LF after { */
        var LF_injected = false
        if current.Id != LF {
            if __IsLeftCurlyBrace(previous) || previous == nil {
                tokens = append(tokens, Token {
                    Id:      syntax.Name2Id["LF"],
                    Span:    zero_span,
                    Content: []rune(""),
                })
                LF_injected = true
            }
        }
        /* tell from "a [LF] (b+c).d" and "a(b+c).d" */
        if current.Id == LeftParentheses || current.Id == LeftBracket {
            if !(LF_injected || __IsLineFeed(previous)) {
                tokens = append(tokens, Token {
                    Id:      syntax.Name2Id["NoLF"],
                    Span:    zero_span,
                    Content: []rune(""),
                })
            }
        }
        tokens = append(tokens, current)
        previous = &current
        pos += amount
    }
    if (pos < length) {
        panic(fmt.Sprintf("invalid token at %+v", info[pos]))
    }
    if len(tokens) > 0 && tokens[len(tokens)-1].Id == LF {
        tokens = tokens[:len(tokens)-1]
    }
    return tokens, info
}
