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

func Scan (code Code) (Tokens, RowColInfo) {
    var ignore = make(map[syntax.Id] bool)
    for _, ignore_name := range syntax.IgnoreTokens {
        ignore[syntax.Name2Id[ignore_name]] = true
    }
    var tokens = make(Tokens, 0, 10000)
    var info = GetInfo(code)
    var length = len(code)
    var pos = 0
    for pos < length {
        var amount, id = MatchToken(code, pos)
        if amount == 0 { break }
        if ignore[id] { pos += amount; continue }
        var span = Span { Start: pos, End: pos + amount }
        var current = Token {
            Id:      id,
            Span:    span,
            Content: code[span.Start : span.End],
        }
        tokens = append(tokens, current)
        pos += amount
    }
    if (pos < length) {
        panic(fmt.Sprintf("invalid token at %+v", info[pos]))
    }
    return tokens, info
}
