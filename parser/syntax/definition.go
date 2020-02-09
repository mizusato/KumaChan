package syntax

import "regexp"
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `\{\}\[\]\(\)\.\,\:\;\$#@~\&\|\?\\'"` + "`"

var EscapeMap = map [string] string {
    "_bang1":  "!",
    "_bang2":  "!!",
    "_bar1":   "|",
    "_bar2":   "||",
    "_at":     "@",
}

var Tokens = [...] Token {
    Token { Name: "String",  Pattern: r(`'[^']*'`) },
    Token { Name: "Text",    Pattern: r(`"[^"]*"`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/|\*/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^`+LF+`]*`) },
    Token { Name: "Pragma",  Pattern: r(`#[^`+LF+`]*`) },
    Token { Name: "Blank",   Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",      Pattern: r(`[`+LF+`]+`) },
    Token { Name: "Int",     Pattern: r(`0x[0-9A-Fa-f]+`) },   // hexadecimal
    Token { Name: "Int",     Pattern: r(`\\[0-7]+`) },         // octal
    Token { Name: "Int",     Pattern: r(`\\\([01]+\)`) },      // binary
    Token { Name: "Float",   Pattern: r(`\d+(\.\d+)?[Ee][\+\-]?\d+`) },
    Token { Name: "Float",   Pattern: r(`\d+\.\d+`) },
    Token { Name: "Int",     Pattern: r(`\d+`) },              // decimal
    Token { Name: "(",       Pattern: r(`\(`) },
    Token { Name: ")",       Pattern: r(`\)`) },
    Token { Name: "[",       Pattern: r(`\[`) },
    Token { Name: "]",       Pattern: r(`\]`) },
    Token { Name: "{",       Pattern: r(`\{`) },
    Token { Name: "}",       Pattern: r(`\}`) },
    Token { Name: "...",     Pattern: r(`\.\.\.`) },
    Token { Name: "..",      Pattern: r(`\.\.`) },
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "?",       Pattern: r(`\?`) },
    Token { Name: "??",      Pattern: r(`\?\?`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },
    Token { Name: ":=",      Pattern: r(`\:\=`) },
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: ";",       Pattern: r(`\;`) },
    Token { Name: "$$",      Pattern: r(`\$$`) },
    Token { Name: "$",       Pattern: r(`\$`) },
    Token { Name: "@@",      Pattern: r(`@@`) },
    Token { Name: "@",       Pattern: r(`@`) },
    Token { Name: "~~",      Pattern: r(`~~`) },
    Token { Name: "~",       Pattern: r(`~`) },
    Token { Name: "&&",      Pattern: r(`\&\&`) },
    Token { Name: "&",       Pattern: r(`\&`) },
    Token { Name: "||",      Pattern: r(`\|\|`) },
    Token { Name: "|",       Pattern: r(`\|`) },
    Token { Name: `\`,       Pattern: r(`\\`) },
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+LF+`]+`) },
    //    { Name: "NoLF",    [ Inserted by Scanner ] },
    //    { Name: "Void",    [ Inserted by Scanner ] },
}

var ExtraTokens = [...] string { "NoLF", "Void" }

var ConditionalKeywords = [...] string {
    "@module", "@import", "@from",
    "@type", "@opaque", "@union", "@native",
    "@global", "@local",
    "@const", "@export",
    "@do",
    "@match", "@else", "@if",
    "@let", "@return",
}

var Operators = [...] Operator {}

var SyntaxDefinition = [...] string {
    "eval = commands",
    "module = shebang module_name! commands",
      "shebang? = Pragma",
      "module_name = @module name! ;!",
        "name = Name",
      "commands? = command commands",
        "command = import | decl_type | decl_func | decl_const | do",
          "import = @import name! @from! string! ;!",
          "do = @do expr! ;!",
    "ref = module_prefix name type_args",
      "module_prefix? = name :: | ::",
      "type_args? = ( type! more_types )",
        "more_types? = , type! more_types",
    "type = type_ref | type_literal",
      "type_ref = ref",
      "type_literal = repr",
    "decl_type = opaque_opt @type name! type_params type_value ;!",
      "opaque_opt? = @opaque",
      "type_value = native_type | union_type | compound_type",
        "native_type = @native",
        "union_type = @union {! decl_type! more_decl_types }!",
          "type_params? = ( name! more_names )!",
            "more_names? = , name! more_names",
          "more_decl_types? = decl_type more_decl_types",
        "compound_type = repr!",
          "repr = repr_tuple | repr_bundle | repr_func",
            "repr_tuple = [ ] | [ type_list ]!",
              "type_list = type! more_types",
            "repr_bundle = { } | { field_list }!",
              "field_list = field! more_fields",
                "field = name :! type!",
                "more_fields? = , field! more_fields",
            "repr_func = & input_type! .! output_type!",
              "input_type = type",
              "output_type = type",
    "decl_func = scope name! type_params :=! (! repr_func! )! body ;!",
      "scope = @global | @local",
      "body = native | lambda!",
        "native = @native string!",
        "lambda = & pattern! .! pipe!",
          "pattern = pattern_none | pattern_tuple | pattern_bundle",
            "pattern_none = name",
            "pattern_tuple = [ ] | [ namelist ]!",
              "namelist = name more_names",
            "pattern_bundle = { } | { namelist }!",
    "decl_const = export_opt @const name! type_params :=! const_value ;!",
      "export_opt? = @export",
      "const_value = native_const | expr!",
        "native_const = casts native",
    "expr = casts pipe more_pipes",
      "casts? = cast casts",
        "cast = ( cast_target! )!",
          "cast_target = omit | type",
            "omit = ...",
      "pipe = term more_terms",
        "more_terms? = term more_terms",
      "more_pipes? = _bar1 pipe! more_pipes",
    "term = lambda | match | if | block | get | bundle | tuple | list | text | literal | ref",
      "match = @match tuple {! branch_list else }!",
        "branch_list = branch! more_branches",
          "more_branches? = , branch more_branches",
          "branch = repr_tuple opt_pattern :! branch_value",
            "opt_pattern? = pattern",
            "branch_value = ... | expr!",
        "else? = , @else! :! expr!",
      "if = @if { if_cond! ?! if_yes! :! if_no! }!",
        "if_cond = expr!",
        "if_yes = expr!",
        "if_no = expr!",
      "block = @let { binding! more_bindings return! }!",
        "more_bindings? = , binding! more_bindings",
        "binding = pattern := expr!",
        "return = , @return expr!",
      "get = _at {! expr! }! members",
        "members? = member members",
          "member = opt . name!",
            "opt? = ?",
      "bundle = { } | { update pairlist }!",
        "pairlist = pair! more_pairs",
          "more_pairs? = , pair! more_pairs",
          "pair = name : expr! | name",
        "update? = get ,!",
      "tuple = [ ] | [ exprlist ]!",
        "exprlist = expr! more_exprs",
          "more_exprs? = , expr! more_exprs",
      "list = $ { } | $ {! exprlist }!",
      "text = Text",
      "literal = string | int | float",
        "string = String",
        "int = Int",
        "float = Float",
}
