package syntax


import "regexp"


type Regexp = *regexp.Regexp


func r (pattern string) Regexp {
    return regexp.MustCompile(pattern)
}


type Item int


const (
    Invalid Item = iota
    String
    Raw
    Comment
    Space
    Linefeed    
)


type Token struct {
    Id       Item
    Name     string
    Pattern  Regexp
}


var Tokens = [...]Token {
    Token { Id: String,   Name: "String",   Pattern: r(`^'[^']*'`), },
    Token { Id: String,   Name: "String",   Pattern: r(`^"[^"]*"`), },
    Token { Id: Raw,      Name: "Raw",      Pattern: r(`^/~([^~]|[^/]|~[^/]|[^~]/)*~/`) },
    Token { Id: Comment,  Name: "Comment",  Pattern: r(`^/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Id: Space,    Name: "Space",    Pattern: r(`^[ \t　]+`) },
    Token { Id: Linefeed, Name: "Linefeed", Pattern: r(`^(\r\n|\n|\r)+`) },    
}


