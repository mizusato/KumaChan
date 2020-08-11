package scanner

import "io"
import "fmt"
import "kumachan/parser/syntax"


const ZERO_COLUMN = 0
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
func GetRowColInfo(code Code) RowColInfo {
    var info = make(RowColInfo, 0)
    var row = 1
    var col = 0
    for _, char := range code {
        if char != '\n' {
            col += 1
        } else {
            row += 1
            col = ZERO_COLUMN
        }
        info = append(info, Point { Row: row, Col: col })
    }
    return info
}
type RowSpanMap = []Span
func GetRowSpanMap(code Code) RowSpanMap {
    var span_map = make([] Span, 0)
    span_map = append(span_map, Span {})
    var row = 1
    var col = 0
    for i, char := range code {
        if char != '\n' {
            col += 1
        } else {
            span_map = append(span_map, Span {
                Start: (i - col),
                End:   i,
            })
            row += 1
            col = ZERO_COLUMN
        }
    }
    span_map = append(span_map, Span {
        Start: (len(code) - col),
        End:   len(code),
    })
    return span_map
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


func MatchToken(code Code, pos int, skip_kw bool) (amount int, id syntax.Id) {
    for _, token_def := range syntax.GetTokens() {
        if skip_kw && token_def.Keyword {
            continue
        }
        var reader = RuneListReader { src: code, pos: pos }
        var loc = token_def.Pattern.FindReaderIndex(&reader)
        // fmt.Printf("Try %v\n", token.Name)
        if loc != nil {
            if (loc[0] != 0) {
                panic("invalid token definition " + token_def.Name)
            }
            return loc[1], syntax.Name2Id[token_def.Name]
        }
    }
    return 0, 0
}

func Scan(code Code) (Tokens, RowColInfo, RowSpanMap) {
    var idenitifer = syntax.Name2Id[syntax.IdentifierPartName]
    var ignore = make(map[syntax.Id] bool)
    var keyword = make(map[syntax.Id] bool)
    for _, ignore_name := range syntax.GetIgnoreTokens() {
        ignore[syntax.Name2Id[ignore_name]] = true
    }
    for _, token_def := range syntax.GetTokens() {
        if token_def.Keyword {
            keyword[syntax.Name2Id[token_def.Name]] = true
        }
    }
    var tokens = make(Tokens, 0)
    var info = GetRowColInfo(code)
    var span_map = GetRowSpanMap(code)
    var length = len(code)
    var pos = 0
    for pos < length {
        var amount, id = MatchToken(code, pos, false)
        if amount == 0 { break }
        if ignore[id] { pos += amount; continue }
        if keyword[id] {
            var non_kw_amount, non_kw_id = MatchToken(code, pos, true)
            if non_kw_amount > amount && non_kw_id == idenitifer {
                amount = non_kw_amount
                id = non_kw_id
            }
        }
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
    return tokens, info, span_map
}
