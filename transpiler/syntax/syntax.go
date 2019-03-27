package syntax


import "regexp"

type Regexp = *regexp.Regexp

func r (pattern string) Regexp {
    return regexp.MustCompile(`^` + pattern)
}


type Id int

type Token struct {
    Name     string
    Pattern  Regexp
}

var Id2Name = [...]string {
    "String", "Raw", "Comment", "Blank", "LF",
    "Exp", "Dec", "Int", "Hex", "Oct", "Bin",
    "(", ")", "[", "]", "{", "}",
    ".[", "..[", "...[", ".{", "..{", "...{",
    ".", ",", "::", ":",
    ">=", "<=", "!=", "==", "=>", "=",
    "->", "<-", "<<", ">>", ">", "<",
    "!", "&&", "||", "~", "&", "|", `\`,
    "+", "-", "**", "*", "/", "%",
    "Name",
}

var Name2Id map[string]Id   // initialize in Init()

const Blanks = ` \t\r　`
const Symbols = `\{\}\[\]\(\)\.\,\:\<\>\=\!~\&\|\\\+\-\*\/%`

var Tokens = [...]Token {
    Token { Name: "String",  Pattern: r(`'[^']*'`), },
    Token { Name: "String",  Pattern: r(`"[^"]*"`), },
    Token { Name: "Raw",     Pattern: r(`/~([^~]|[^/]|~[^/]|[^~]/)*~/`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^\n]*`) },
    Token { Name: "Comment", Pattern: r(`\.\.[^\[\{][^`+Blanks+`]*`) },
    Token { Name: "Blank",   Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",      Pattern: r(`\n+`) },
    Token { Name: "Exp",     Pattern: r(`\d+(\.\d+)?[Ee][\+\-]?\d+`) },
    Token { Name: "Dec",     Pattern: r(`\d+\.\d+`) },
    Token { Name: "Int",     Pattern: r(`\d+`) },
    Token { Name: "Hex",     Pattern: r(`0x[0-9A-Fa-f]+`) },
    Token { Name: "Oct",     Pattern: r(`\\[0-7]+`) },
    Token { Name: "Bin",     Pattern: r(`\\\([01]+\)`) },
    Token { Name: "(",       Pattern: r(`\(`) },
    Token { Name: ")",       Pattern: r(`\)`) },
    Token { Name: "[",       Pattern: r(`\[`) },
    Token { Name: "]",       Pattern: r(`\]`) },
    Token { Name: "{",       Pattern: r(`\{`) },
    Token { Name: "}",       Pattern: r(`\}`) },
    Token { Name: ".[",      Pattern: r(`\.\[`) },
    Token { Name: "..[",     Pattern: r(`\.\.\[`) },
    Token { Name: "...[",    Pattern: r(`\.\.\.\[`) },
    Token { Name: ".{",      Pattern: r(`\.\{`) },
    Token { Name: "..{",     Pattern: r(`\.\.\{`) },
    Token { Name: "...{",    Pattern: r(`\.\.\.\{`) },
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: ">=",      Pattern: r(`\>\=`) },
    Token { Name: "<=",      Pattern: r(`\<\=`) },
    Token { Name: "!=",      Pattern: r(`\!\=`) },
    Token { Name: "==",      Pattern: r(`\=\=`) },
    Token { Name: "=>",      Pattern: r(`\=\>`) },
    Token { Name: "->",      Pattern: r(`\-\>`) },
    Token { Name: "<-",      Pattern: r(`\<\-`) },
    Token { Name: "<<",      Pattern: r(`\<\<`) },
    Token { Name: ">>",      Pattern: r(`\>\>`) },
    Token { Name: "<",       Pattern: r(`\<`) },
    Token { Name: ">",       Pattern: r(`\>`) },
    Token { Name: "!",       Pattern: r(`\!`) },
    Token { Name: "&&",      Pattern: r(`\&\&`) },
    Token { Name: "||",      Pattern: r(`\|\|`) },
    Token { Name: "~",       Pattern: r(`~`) },
    Token { Name: "&",       Pattern: r(`\&`) },
    Token { Name: "|",       Pattern: r(`\|`) },
    Token { Name: `\`,       Pattern: r(`\\`) },
    Token { Name: "+",       Pattern: r(`\+`) },
    Token { Name: "-",       Pattern: r(`\-`) },
    Token { Name: "**",       Pattern: r(`\*\*`) },
    Token { Name: "*",       Pattern: r(`\*`) },
    Token { Name: "/",       Pattern: r(`\/`) },
    Token { Name: "%",       Pattern: r(`%`) },
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+`]+`) },
}


func Init () {
    Name2Id = make(map[string]Id)
    for id, name := range Id2Name {
        Name2Id[name] = Id(id)
    }
}
