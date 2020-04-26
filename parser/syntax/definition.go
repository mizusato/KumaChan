package syntax

import "regexp"
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }


const IdentifierPartName = "Name"
var IgnoreTokens = [] string { "Comment", "Blank", "LF" }
var EscapeMap = map[string] string {
    "_bang1":  "!",
    "_bang2":  "!!",
    "_bar1":   "|",
    "_bar2":   "||",
    "_at":     "@",
}

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `\{\}\[\]\(\)\.\,\:\;\~@#$\^\&\|\\'"` + "`"

var Tokens = [...] Token {
    Token { Name: "String",  Pattern: r(`'[^']*'`) },
    Token { Name: "Text",    Pattern: r(`"[^"]*"`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Doc",     Pattern: r(`///[^`+LF+`]*`) },
    Token { Name: "Comment", Pattern: r(`//[^`+LF+`]*`) },
    Token { Name: "Pragma",  Pattern: r(`#[^`+LF+`]*`) },
    Token { Name: "Blank",   Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",      Pattern: r(`[`+LF+`]+`) },
    Token { Name: "Int",     Pattern: r(`\-?0[xX][0-9A-Fa-f]+`) },
    Token { Name: "Int",     Pattern: r(`\-?0[oO][0-7]+`) },
    Token { Name: "Int",     Pattern: r(`\-?0[bB][01]+`) },
    Token { Name: "Float",   Pattern: r(`\-?\d+(\.\d+)?[Ee][\+\-]?\d+`) },
    Token { Name: "Float",   Pattern: r(`\-?\d+\.\d+`) },
    Token { Name: "Int",     Pattern: r(`\-?\d[\d_]*`) },
    Token { Name: "Char",    Pattern: r(`\^.`) },
    Token { Name: "Char",    Pattern: r(`\\u[0-9A-Fa-f]+`) },
    Token { Name: "Char",    Pattern: r(`\\[a-z]`) },
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
    Token { Name: "::",      Pattern: r(`\:\:`) },
    Token { Name: ":=",      Pattern: r(`\:\=`) },
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: ";",       Pattern: r(`\;`) },
    Token { Name: "~",       Pattern: r(`\~`) },
    Token { Name: "`",       Pattern: r("`") },
    Token { Name: "@",       Pattern: r(`\@`) },
    Token { Name: "$",       Pattern: r(`\$`) },
    Token { Name: "&",       Pattern: r(`\&`) },
    Token { Name: "|",       Pattern: r(`\|`) },
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+LF+`]+`) },
}

var ConditionalKeywords = [...] string {
    "@module", "@import", "@from",
    "@type", "@union", "@native", "@protected", "@opaque",
    "@private", "@public", "@function", "@const", "@macro", "@do",
    "@if", "@elif", "@else", "@switch", "@case", "@default", "@switch*",
    "@lambda", "@let", "@rec", "@return",
}

var SyntaxDefinition = [...] string {
    "module = shebang module_name! commands",
      "shebang? = Pragma",
      "module_name = @module name! ;!",
        "name = Name",
      "commands? = command commands",
        "command = import | do | decl_type | decl_func | decl_const | decl_macro",
          "import = @import name! @from! string! ;!",
          "do = @do expr! ;!",
    "ref = module_prefix name type_args",
      "module_prefix? = name :: | ::",
      "type_args? = [ type! more_types ]!",
        "more_types? = , type! more_types",
    "type = type_literal | type_ref",
      "type_ref = ref",
      "type_literal = repr",
        "repr = repr_tuple | repr_bundle | repr_func",
          "repr_tuple = ( ) | ( type_list )!",
            "type_list = type! more_types",
          "repr_bundle = { } | { field_list }!",
            "field_list = field! more_fields",
              "field = name :! type!",
              "more_fields? = , field! more_fields",
          "repr_func = lambda_header input_type! output_type!",
            "lambda_header = @lambda | & ",
            "input_type = type",
            "output_type = type",
    "decl_type = @type name! type_params docs_brief type_value ;!",
      "docs_brief = docs",
        "docs? = doc docs",
          "doc = Doc",
      "type_value = native_type | union_type | boxed_type",
        "native_type = @native",
        "union_type = @union {! decl_type! more_decl_types }!",
          "type_params? = [ namelist ]!",
            "namelist = name! more_names",
              "more_names? = , name! more_names",
          "more_decl_types? = decl_type more_decl_types",
        "boxed_type = box_option inner_type",
          "box_option? = @protected | @opaque",
          "inner_type? = type",
    "decl_func = scope @function name! type_params {! docs_brief signature! :! body docs_detail }! ;!",
      "scope = @public | @private",
      "signature = repr_func",
      "body = native | lambda!",
        "native = @native string!",
        "lambda = lambda_header pattern! term!",
          "pattern = pattern_trivial | pattern_tuple | pattern_bundle",
            "pattern_trivial = name",
            "pattern_tuple = ( ) | ( namelist )!",
            "pattern_bundle = { } | { namelist }!",
      "docs_detail = docs",
    "decl_const = scope @const name! :=! [! type! ]! docs_brief const_value ;!",
      "const_value = native | expr!",
    "decl_macro = scope @macro name! macro_params :=! expr! ;!",
      "macro_params = ( ) | ( namelist )!",
    "expr = call pipeline",
      "call = func arg",
        "func = term",
        "arg? = call",
      "pipeline? = pipe_op pipe_func pipe_arg pipeline",
        "pipe_op = _bar1 | . ",
        "pipe_func = term!",
        "pipe_arg? = call",
    "term = cast | lambda | multi_switch | switch | if | block | cps | bundle | get | tuple | infix | array | text | literal | ref",
      "cast = [ type expr ]!",
      "switch = @switch term {! branch_list }!",
        "branch_list = branch! more_branches",
          "more_branches? = , branch more_branches",
          "branch = branch_key :! expr!",
            "branch_key = @default | @case type_ref! opt_pattern",
              "opt_pattern? = pattern",
      "multi_switch = @switch* ( exprlist )! {! multi_branch_list }!",
        "exprlist = expr! more_exprs",
          "more_exprs? = , expr! more_exprs",
        "multi_branch_list = multi_branch! more_multi_branches",
          "more_multi_branches? = , multi_branch more_multi_branches",
          "multi_branch = multi_branch_key :! expr!",
            "multi_branch_key = @default | @case multi_type_ref! multi_pattern",
              "multi_type_ref = [! ref_list ]!",
                "ref_list = ref! more_refs",
                  "more_refs? = , ref! more_refs",
              "multi_pattern? = pattern_tuple",
      "if = @if term :! if_yes ,! elifs @else! :! if_no",
        "elifs? = elif elifs",
          "elif = @elif term! :! expr! ,!",
        "if_yes = expr!",
        "if_no = expr!",
      "block = @let binding! more_bindings return!",
        "more_bindings? = , binding! more_bindings",
        "binding = rec_opt pattern := binding_type expr!",
          "rec_opt? = @rec :!",
          "binding_type? = [ type! ]!",
        "return = , @return expr!",
      "cps = ~ ref! cps_binding cps_input :! cps_output",
        "cps_binding? = lambda_header pattern := binding_type",
        "cps_input = expr!",
        "cps_output = expr!",
      "bundle = { } | { update pairlist }!",
        "pairlist = pair! more_pairs",
          "more_pairs? = , pair! more_pairs",
          "pair = name : expr! | name",
        "update? = ... expr! ,!",
      "get = $ { expr! member! more_members }!",
        "more_members? = member more_members",
        "member = . name!",
      "tuple = ( ) | ( exprlist )!",
      "infix = $ (! operand1 operator operand2 )!",
        "operand1 = term!",
        "operator = term!",
        "operand2 = term!",
      "array = _at ( ) | _at (! exprlist )!",
      "text = Text",
      "literal = string | int | float | char",
        "string = String",
        "int = Int",
        "float = Float",
        "char = Char",
    // TODO: compile-time concatenated string & text (with chars)
    //       e.g.
    //         'Hello' \n 'World' ^!
    //         ("Number " ^# "#") (str! n)   // explicit ^# as raw #
    // TODO: l10n message constants (type: lambda {...} String)
    //       message NumRecords { name: String, count: Int }:
    //           number:   count,
    //           plural:   'The table ' .name. ' has ' .count. '  records' \n,
    //           singular: 'The table ' .name. ' has ' .count. ' record' \n
    // TODO: bytes literal and compact array literal
}
