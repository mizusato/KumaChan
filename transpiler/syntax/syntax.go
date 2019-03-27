package syntax


import "regexp"
import "strings"

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
    // Token
    "String", "Raw", "Comment", "Blank", "LF",
    "Hex", "Oct", "Bin", "Exp", "Dec", "Int",
    "(", ")", "[", "]", "{", "}",
    ".[", "..[", "...[", ".{", "..{", "...{",
    ".", ",", "::", ":",
    ">=", "<=", "!=", "==", "=>", "=",
    "-->", "<--", "->", "<-", "<<", ">>", ">", "<",
    "!", "&&", "||", "~", "&", "|", `\`,
    "+", "-", "*", "/", "%", "^",
    "Name",
    // Inserted
    "#call", "#get",
    // Keyword
    "@true", "@false",
    // Item
    "program", "expr",  // TODO
}

var Name2Id map[string]Id   // initialize in Init()

var Keywords map[Id][]rune // initialize in Init()


const LF = `\n`
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
    Token { Name: "LF",      Pattern: r(LF+`+`) },
    Token { Name: "Hex",     Pattern: r(`0x[0-9A-Fa-f]+`) },
    Token { Name: "Oct",     Pattern: r(`\\[0-7]+`) },
    Token { Name: "Bin",     Pattern: r(`\\\([01]+\)`) },
    Token { Name: "Exp",     Pattern: r(`\d+(\.\d+)?[Ee][\+\-]?\d+`) },
    Token { Name: "Dec",     Pattern: r(`\d+\.\d+`) },
    Token { Name: "Int",     Pattern: r(`\d+`) },
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
    Token { Name: "-->",      Pattern: r(`\-\-\>`) },
    Token { Name: "<--",      Pattern: r(`\<\-\-`) },
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
    Token { Name: "*",       Pattern: r(`\*`) },
    Token { Name: "/",       Pattern: r(`\/`) },
    Token { Name: "%",       Pattern: r(`%`) },
    Token { Name: "^",       Pattern: r(`\^`) },
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+LF+`]+`) },
}


var SyntaxDefinition = [...]string {

    "program = expr", // TODO: commands

    "name = Name | String",

    "expr = operand expr_tail",
    "expr_tail? = operator oprand! expr_tail",
    "operator = ",  // TODO

    "operand = operand_base operand_ext",
    "operand_base = ( expr! )! | function | literal | dot_para | identifier",
    "operand_ext? = get | call",
    "get = #get [ expr ]",
    "call = #call arglist | -> name #call arglist",
    "arglist = ( expr_list ) extra_arg",
    "extra_arg? = --> expr!",
    "expr_list? = expr expr_list_tail",
    "expr_list_tail? = , expr! expr_list_tail",

    "function = fun_expr | lambda | bool_lambda",
    "fun_expr = paralist ->! {! body! }!",
    "lambda = .{ paralist ->! expr! }! | .{ expr! }!",
    "bool_lambda = ..{ paralist ->! expr! }! | ..{ expr! }!",

    "paralist = ( ) | ( name_list ) | ( typed_list )",
    "name_list = name name_list_tail",
    "name_list_tail? = , name! name_list_tail",
    "typed_list = type name! typed_list_tail",
    "typed_list_tail? = , type! name! typed_list_tail",

    "type = type_base type_ext",
    "type_ext? = < type_args >",
    "type_base = name type_base_tail",
    "type_base_tail? = . name! type_base_tail",
    "type_args = type type_args_tail",
    "type_args_tail? = , type! type_args_tail",

    "identifier = Name",
    "dot_para = . Name",
    "literal = hash | list | primitive",

    "hash = { } | { hash_item hash_tail }",
    "hash_tail? = , hash_item! hash_tail",
    "hash_item = name : expr",

    "list = [ ] | [ list_item list_tail ]",
    "list_tail? = , list_item! list_tail",
    "list_item = expr",

    "primitive = string | number | bool",
    "string = String",
    "number = Hex | Exp | Dec | Int",
    "bool = @true | @false",

}


func Init () {
    Name2Id = make(map[string]Id)
    Keywords = make(map[Id][]rune)
    for id, name := range Id2Name {
        Name2Id[name] = Id(id)
        if strings.HasPrefix(name, "@") {
            var rlist = []rune(name)
            Keywords[Id(id)] = rlist[1:]
        }
    }
}
