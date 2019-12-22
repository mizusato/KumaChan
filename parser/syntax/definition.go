package syntax

import "regexp"
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `;\{\}\[\]\(\)\.\,\:\$#@\?\<\>\=\!~\&\|\\\+\-\*\/%\^'"` + "`"

var EscapeMap = map [string] string {
    "_exc1":  "!",
    "_exc2":  "!!",
    "_bar1":  "|",
    "_bar2":  "||",
    "_at":    "@",
}

var Tokens = [...] Token {
    Token { Name: "String",  Pattern: r(`'[^']*'`) },
    Token { Name: "Text",    Pattern: r(`"[^"#\[\]]*"`) },
    Token { Name: "TxBegin", Pattern: r(`"[^"#\[\]]*#\[`) },
    Token { Name: "TxInner", Pattern: r(`\][^"#\[\]]*#\[`) },
    Token { Name: "TxEnd",   Pattern: r(`\][^"#\[\]]*"`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^;`+LF+`]*`) },
    Token { Name: "Pragma",  Pattern: r(`#[^;`+LF+`]*`) },
    Token { Name: "Blank",   Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",      Pattern: r(`[;`+LF+`]+`) },
    Token { Name: "Hex",     Pattern: r(`0x[0-9A-Fa-f]+`) },
    Token { Name: "Oct",     Pattern: r(`\\[0-7]+`) },
    Token { Name: "Bin",     Pattern: r(`\\\([01]+\)`) },
    Token { Name: "Exp",     Pattern: r(`\d+(\.\d+)?[Ee][\+\-]?\d+`) },
    Token { Name: "Float",   Pattern: r(`\d+\.\d+`) },
    Token { Name: "Dec",     Pattern: r(`\d+`) },
    Token { Name: "(",       Pattern: r(`\(`) },
    Token { Name: ")",       Pattern: r(`\)`) },
    Token { Name: "[",       Pattern: r(`\[`) },
    Token { Name: "]",       Pattern: r(`\]`) },
    Token { Name: "{",       Pattern: r(`\{`) },
    Token { Name: "}",       Pattern: r(`\}`) },
    Token { Name: "...",     Pattern: r(`\.\.\.`) }, // (Reserved)
    Token { Name: "..",      Pattern: r(`\.\.`) },   // (Reserved)
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },   // Module Namespace
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: "$",       Pattern: r(`\$`) },     // Iterator Literal
    Token { Name: "$$",      Pattern: r(`\$$`) },    // (Reserved)
    Token { Name: "@",       Pattern: r(`@`) },
    Token { Name: "@@",      Pattern: r(`@@`) },     // (Reserved)
    Token { Name: "??",      Pattern: r(`\?\?`) },   // Nil Coalescing
    Token { Name: "?",       Pattern: r(`\?`) },     // Optional Chaining
    Token { Name: ">=",      Pattern: r(`\>\=`) },
    Token { Name: "<=",      Pattern: r(`\<\=`) },
    Token { Name: "==",      Pattern: r(`\=\=`) },
    Token { Name: "!=",      Pattern: r(`\!\=`) },
    Token { Name: "=>",      Pattern: r(`\=\>`) },
    Token { Name: "=",       Pattern: r(`\=`) },
    Token { Name: "->",      Pattern: r(`\-\>`) },
    Token { Name: "<-",      Pattern: r(`\<\-`) },
    Token { Name: "<<",      Pattern: r(`\<\<`) },   // Bitwise SHL
    Token { Name: ">>",      Pattern: r(`\>\>`) },   // Bitwise SHR
    Token { Name: "<",       Pattern: r(`\<`) },
    Token { Name: ">",       Pattern: r(`\>`) },
    Token { Name: "!!",      Pattern: r(`\!\!`) },   // Bitwise NOT
    Token { Name: "!",       Pattern: r(`\!`) },     // Type Assertion
    Token { Name: "~",       Pattern: r(`~`) },      // Continuation
    Token { Name: "&&",      Pattern: r(`\&\&`) },   // Bitwise AND
    Token { Name: "&",       Pattern: r(`\&`) },     // Procedure Flag
    Token { Name: "||",      Pattern: r(`\|\|`) },   // Bitwise OR
    Token { Name: "|",       Pattern: r(`\|`) },     // Pipeline
    Token { Name: `\`,       Pattern: r(`\\`) },     // (Reserved)
    Token { Name: "++",      Pattern: r(`\+\+`) },   // Pull from Iterator
    Token { Name: "+",       Pattern: r(`\+`) },
    Token { Name: "--",      Pattern: r(`\--`) },    // (Reserved)
    Token { Name: "-",       Pattern: r(`\-`) },
    Token { Name: "**",      Pattern: r(`\*\*`) },   // (Reserved)
    Token { Name: "*",       Pattern: r(`\*`) },
    Token { Name: "/",       Pattern: r(`\/`) },
    Token { Name: "%%",      Pattern: r(`%%`) },     // (Reserved)
    Token { Name: "%",       Pattern: r(`%`) },
    Token { Name: "^^",      Pattern: r(`\^\^`) },   // Bitwise XOR
    Token { Name: "^",       Pattern: r(`\^`) },     // Tuple
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+LF+`]+`) },
    //    { Name: "NoLF",    [ Inserted by Scanner ] },
    //    { Name: "Void",    [ Inserted by Scanner ] },
}

var ExtraTokens = [...] string { "NoLF", "Void" }

var ConditionalKeywords = [...] string {

    "@external", "@import", "@as",

    "@section",
    "@procedure", "@function",
    "@singleton",
    "@struct", "@class", "@mutable",
    "@extends", "@static", "@init", "@data", "@private", "@native",
    "@type",
    "@trait", "@is", "@union",
    "@const",

    "@if", "@else",
    "@while", "@for", "@in", "@break", "@continue",
    "@yield",
    "@return",
    "@assert", "@panic", "@finally",
    "@let", "@initial", "@reset",
    "@do", "@nothing",

    "@not", "@and", "@or", "@try",
    "@callcc",
    "@new",
    "@when", "@switch", "@otherwise",
    "@with",
    "@Yes", "@No",

}


/* Infix Operators */
var Operators = [...] Operator {
    /* Nil Coalescing */
    Operator { Match: "??",   Priority: 60,  Assoc: Left,   Lazy: true   },
    /* Comparison */
    Operator { Match: "<",    Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">",    Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "<=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "==",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    /* Bitwise */
    Operator { Match: "<<",   Priority: 40,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">>",   Priority: 40,  Assoc: Left,   Lazy: false  },
    Operator { Match: "&&",   Priority: 35,  Assoc: Left,   Lazy: false  },
    Operator { Match: "^^",   Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: "||",   Priority: 25,  Assoc: Left,   Lazy: false  },
    /* Logic */
    Operator { Match: "@and", Priority: 20,  Assoc: Left,   Lazy: true   },
    Operator { Match: "@or",  Priority: 10,  Assoc: Left,   Lazy: true   },
    /* Arithmetic */
    Operator { Match: "+",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "-",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "*",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "/",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "%",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "^",    Priority: 90,  Assoc: Right,  Lazy: false  },
}


var SyntaxDefinition = [...] string {
    /* Group: Module */
    "module = module_header commands",
      "module_header = shebang module_name_decl!",
        "shebang? = Pragma",
        "module_name_decl = @module name",
          "name = Name | operator",
            "operator = ?? | ++ | -- | < | > | <= | >= | == | != | + | - | * | / | % | ^ ",
    "commands?  = command commands",
      "command = import | decl | do",
        "import = @import name! @from! imported_path!",
          "imported_path = String",
        "decl = decl_type | decl_fun | decl_val",
        "do = @do expr!",
    "ref = module_prefix name type_args",
      "type_args? = [ type! more_types ]",
        "more_types? = , type! more_types",
    "type = type_ref | type_literal",
      "type_ref = ref",
      "type_literal = repr",
    "decl_type = visibility opaque_opt type_body",
      "visibility? = @global | @local",
      "opaque_opt? = @opaque",
      "type_body = type_leaf | type_union",
        "type_union = @union @type name! type_params { type_decl more_type_decls }",
          "more_type_decls? = type_decl more_type_decls",
        "type_leaf = @type name! type_params repr!",
          "type_params? = [ type_param! more_type_params ]!",
            "more_type_params? = , type_param! more_type_params",
            "type_param = name",
          "repr = repr_tuple | repr_bundle | repr_func | repr_native",
            "repr_tuple = ( ) | ( type! more_types )!",
              "more_types? = , type! more_types",
            "repr_bundle = { } | { field! more_fields }!",
              "field = name :! type!",
              "more_fields? = , field! more_fields",
            "repr_func = [ signature! ]!",
              "signature = input_type ->! output_type!",
                "input_type = type",
                "output_type = type",
            "repr_native = @native native_id",
              "native_id = String",
    "decl_fun = visibility @fun name! type_params =! [! signature! ]! lambda!",
      "lambda = pattern => expr!",
        "pattern = pattern_none | pattern_tuple | pattern_bundle",
          "pattern_none = name",
          "pattern_tuple = ( ) | ( name more_names )",
            "more_names? = , name more_names",
          "pattern_bundle = { } | { name more_names }",
    "decl_val = visibility @val name! =! expr!",

    "expr = operand expr_tail",
    "expr_tail? = operator operand! expr_tail",
    // operand -> Group: Operand

    "expr = casts call pipes",
      "casts? = cast casts",
        "cast? = opt [ cast_type! ]!",
          "opt? = ?",
          "cast_type = * | type",
      "call = term more_terms",
        "more_terms? = term more_terms",
      "pipes? = pipe pipes",
        "pipe = opt -> target_fun! opt_call",
          "target_fun = name | ( expr )!",
          "opt_call? = call",
    "term = if | match | block | access | bundle | tuple | list | text | literal | ref",
    "block = @let { binding_list return! }!",
      "binding_list? = binding binding_list_tail",
        "binding_list_tail? = , binding! binding_list_tail",
        "binding = pattern = expr!",
      "return = @return expr!",
    "bundle = { } | { pair more_pairs }!",
      "more_pairs? = , pair! more_pairs",
      "pair = name : expr! | name",

    "literal = string | int | float | bool",
    "string = String",
    "int = Dec | Hex | Oct | Bin",
    "float = Float | Exp",
    "bool = @Yes | @No",

    "text = Text | TxBegin first_segment more_segments TxEnd!",
    "first_segment = segment",
    "segment = segment_tag expr!",
    "segment_tag? = name : ",
    "more_segments? = next_segment more_segments",
    "next_segment = TxInner segment",

}
