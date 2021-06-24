package syntax

import (
    "regexp"
    "strings"
)
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }


const RootPartName = "root"
const ReplRootPartName = "repl_root"
const IdentifierPartName = "Name"
var __EscapeMap = map[string] string {
    "bar":  "|",
}
var __IgnoreTokens = [...] string {
    "Comment",
    "Blank",
    "LF",
}

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `\{\}\[\]\(\)\.,:;#\&\|\\'"` + "`"
const IdentifierRegexp = `[^`+Symbols+Blanks+LF+`]+`
const IdentifierFullRegexp = "^" + IdentifierRegexp + "$"

var __Tokens = [...] Token {
    // pragma and comment
    Token { Name: "Shebang",  Pattern: r(`#![^`+LF+`]*`) },
    Token { Name: "Title",    Pattern: r(`##[^`+LF+`]*`) },
    Token { Name: "Meta",     Pattern: r(`#[^`+LF+`]*`) },
    Token { Name: "Comment",  Pattern: r(`/\*([^\*/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Doc",      Pattern: r(`///[^`+LF+`]*`) },
    Token { Name: "Comment",  Pattern: r(`//[^`+LF+`]*`) },
    // blank
    Token { Name: "Blank",  Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",     Pattern: r(`[`+LF+`]+`) },
    // literals
    Token { Name: "SqStr",  Pattern: r(`'[^']*'`) },
    Token { Name: "DqStr",  Pattern: r(`"[^"]*"`) },
    Token { Name: "Int",    Pattern: r(`\-?0[xX][0-9A-Fa-f]+`) },
    Token { Name: "Int",    Pattern: r(`\-?0[oO][0-7]+`) },
    Token { Name: "Int",    Pattern: r(`\-?0[bB][01]+`) },
    Token { Name: "Float",  Pattern: r(`\-?\d+(\.\d+)?[Ee][\+\-]?\d+`) },
    Token { Name: "Float",  Pattern: r(`\-?\d+\.\d+`) },
    Token { Name: "Int",    Pattern: r(`\-?\d[\d_]*`) },
    Token { Name: "Char",   Pattern: r("`.`") },
    Token { Name: "Char",   Pattern: r(`\\u[0-9A-Fa-f]+`) },
    Token { Name: "Char",   Pattern: r(`\\[a-z]`) },
    // symbols
    Token { Name: "(",    Pattern: r(`\(`) },
    Token { Name: ")",    Pattern: r(`\)`) },
    Token { Name: "[",    Pattern: r(`\[`) },
    Token { Name: "]",    Pattern: r(`\]`) },
    Token { Name: "{",    Pattern: r(`\{`) },
    Token { Name: "}",    Pattern: r(`\}`) },
    Token { Name: "...",  Pattern: r(`\.\.\.`) },
    Token { Name: "..",   Pattern: r(`\.\.`) },
    Token { Name: ".",    Pattern: r(`\.`) },
    Token { Name: ",",    Pattern: r(`\,`) },
    Token { Name: "::[",  Pattern: r(`\:\:[`+Blanks+`]*\[`) },
    Token { Name: "::",   Pattern: r(`\:\:`) },
    Token { Name: ":=",   Pattern: r(`\:\=`) },
    Token { Name: ":",    Pattern: r(`\:`) },
    Token { Name: ";",    Pattern: r(`\;`) },
    Token { Name: "&",    Pattern: r(`\&`) },
    Token { Name: "|",    Pattern: r(`\|`) },
    Token { Name: `\`,    Pattern: r(`\\`) },
    // keywords
    Token { Name: "If",         Pattern: r(`if`),         Keyword: true },
    Token { Name: "Elif",       Pattern: r(`elif`),       Keyword: true },
    Token { Name: "Else",       Pattern: r(`else`),       Keyword: true },
    Token { Name: "Switch",     Pattern: r(`switch`),     Keyword: true },
    Token { Name: "Select",     Pattern: r(`select`),     Keyword: true },
    Token { Name: "Case",       Pattern: r(`case`),       Keyword: true },
    Token { Name: "Let",        Pattern: r(`let`),        Keyword: true },
    // identifier
    Token { Name: "Name", Pattern: r(IdentifierRegexp) },
}
func GetTokens() ([] Token) { return __Tokens[:] }
func GetIgnoreTokens() ([] string) { return __IgnoreTokens[:] }
func GetIdentifierRegexp() *regexp.Regexp {
    return regexp.MustCompile(IdentifierRegexp)
}
func GetIdentifierFullRegexp() *regexp.Regexp {
    return regexp.MustCompile(IdentifierFullRegexp)
}

var __ConditionalKeywords = [...] string {
    "@import", "@from",
    "@type", "@enum", "@interface", "@native",
    "@weak", "@protected", "@opaque",
    "@export", "@function", "@const", "@do", "@alias",
    "@<", "@>", "@=>", "@exact",
    "@default", "@end", "@rec",
}
func GetKeywordList() ([] string) {
    var list = make([] string, 0)
    for _, v := range __ConditionalKeywords {
        var kw = strings.TrimPrefix(v, "@")
        list = append(list, kw)
    }
    for _, t := range __Tokens {
        if t.Keyword {
            var kw = strings.ToLower(t.Name)
            list = append(list, kw)
        }
    }
    return list
}

var __SyntaxDefinition = [...] string {
    "root = shebang stmts",
      "shebang? = Shebang",
      "stmts? = stmt stmts",
        "stmt = title | import | do | alias | decl_type | decl_const | decl_func",
          "title = Title",
          "import = @import name! @from! string_text! ;!",
            "name = Name",
          "do = @do expr! ;!",
          "alias = @alias name! :=! alias_target ;!",
            "alias_target = module_prefix name!",
    "repl_root = repl_assign | repl_do | repl_eval",
        "repl_assign = name := expr!",
        "repl_do = @do expr!",
        "repl_eval = expr!",
    "type = type_literal | type_ref",
      "type_ref = module_prefix name type_args",
        "module_prefix? = name :: | ::",
        "type_args? = [ type! more_types ]!",
          "more_types? = , type! more_types",
      "type_literal = repr",
        "repr = repr_func | repr_tuple | repr_record",
          "repr_tuple = ( ) | ( type_list )!",
            "type_list = type! more_types",
          "repr_record = { } | { field_list }!",
            "field_list = field! more_fields",
              "field = docs meta name :! type!",
              "more_fields? = , field! more_fields",
          "repr_func = & input_type! @=>! output_type!",
            "input_type = type",
            "output_type = type",
    "decl_type = docs meta @type name! type_params impl type_def ;!",
      "docs? = doc docs",
        "doc = Doc",
      "meta = meta_items",
        "meta_items? = meta_item meta_items",
        "meta_item = Meta",
      "impl? = ( type_decl_ref! more_type_decl_refs )!",
        "more_type_decl_refs? = , type_decl_ref! more_type_decl_refs",
        "type_decl_ref = module_prefix name",
      "type_def = t_native | t_enum | t_interface | t_boxed",
        "t_native = @native",
        "t_enum = @enum {! decl_type! more_decl_types }!",
          "more_decl_types? = decl_type more_decl_types",
        "t_interface = @interface repr_record",
        "t_boxed = box_option match_option inner_type",
          "box_option? = @protected | @opaque",
          "match_option? = @weak",
          "inner_type? = type",
      "type_params? = [ type_param! more_type_params ]!",
        "more_type_params? = , type_param more_type_params",
        "type_param = type_param_default name type_bound",
          "type_bound? = type_hi_bound | type_lo_bound",
            "type_hi_bound = @< type!",
            "type_lo_bound = @> type!",
          "type_param_default? = [ type! ]!",
    "decl_func = docs meta scope @function name! :! type_params sig! body ;!",
      "scope? = @export",
      "sig = implicit repr_func",
        "implicit? = repr_record",
      "body? = native | lambda",
        "native = @native string_text!",
        "lambda = & pattern! @=> expr!",
          "pattern = pattern_trivial | pattern_tuple | pattern_record",
            "pattern_trivial = name",
            "pattern_tuple = ( ) | ( namelist )!",
              "namelist = name! more_names",
              "more_names? = , name! more_names",
            "pattern_record = { } | { field_map_list }!",
              "field_map_list = field_map more_field_maps",
                "more_field_maps? = , field_map more_field_maps",
                "field_map = name field_map_to",
                  "field_map_to? = : name",
    "decl_const = docs meta scope @const name! :! type! const_def ;!",
      "const_def? = := const_value",
      "const_value = native | expr!",
    "expr = term pipes",
      "pipes? = . pipe! pipes",
        "pipe = pipe_func | pipe_cast " +
            "| pipe_get | pipe_ref_field " +
            "| pipe_switch | pipe_ref_branch",
          "pipe_func = { callee! pipe_func_arg }!",
            "callee = expr",
            "pipe_func_arg? = expr",
          "pipe_cast = [ type! ]!",
          "pipe_get = name",
          "pipe_ref_field = & name",
          "pipe_switch = ( type_ref! )!",
          "pipe_ref_branch = & ( type_ref! )!",
    "term = call | ctor_lambda | pipeline_lambda | lambda " +
        "| switch | select | if " +
        "| block | cps | record | tuple | inline_ref " +
        "| array | int | float | formatter | string | char",
      "call = call_prefix | call_infix",
        "call_prefix = { callee expr }!",
        "call_infix = ( infix_left operator infix_right! )!",
          "infix_left = expr",
          "operator = expr",
          "infix_right = expr",
      "ctor_lambda = bar ctor_modifier type_ref bar!",
        "ctor_modifier? = @exact : ",
      "pipeline_lambda = bar pipes bar!",
      "switch = Switch expr :! sw_branch_list ,! @end!",
        "sw_branch_list = sw_branch! more_sw_branches",
          "more_sw_branches? = , sw_branch more_sw_branches",
          "sw_branch = sw_key :! expr!",
            "sw_key = @default | Case namelist opt_pattern",
              "opt_pattern? = pattern",
      "select = Select ( exprlist )! :! sl_branch_list ,! @end!",
        "exprlist = expr! more_exprs",
          "more_exprs? = , expr! more_exprs",
        "sl_branch_list = sl_branch! more_sl_branches",
          "more_sl_branches? = , sl_branch more_sl_branches",
          "sl_branch = sl_key :! expr!",
            "sl_key = @default | Case sl_case_list sl_pattern",
              "sl_pattern? = pattern_tuple",
              "sl_case_list = sl_case! more_sl_cases",
                "more_sl_cases? = , sl_case! more_sl_cases",
                "sl_case = ( namelist )!",
      "if = If cond :! if_yes ,! elifs Else! :! if_no",
        "cond = expr",
        "elifs? = elif elifs",
          "elif = Elif cond! :! expr! ,!",
        "if_yes = expr!",
        "if_no = expr!",
      "block = binding more_bindings block_value",
        "more_bindings? = , binding more_bindings",
        "binding = Let pattern! binding_type :=! expr!",
          "binding_type? = : rec_opt type!",
            "rec_opt? = @rec",
        "block_value = ,! expr!",
      "cps = \\ cps_binding callee! cps_input ,! cps_output",
        "cps_binding? = pattern cps_binding_type := ",
          "cps_binding_type? = : type!",
        "cps_input = expr!",
        "cps_output = expr!",
      "record = { } | { update pairlist }!",
        "pairlist = pair! more_pairs",
          "more_pairs? = , pair! more_pairs",
          "pair = name : expr! | name",
        "update? = ... expr! ,!",
      "tuple = ( ) | ( exprlist )!",
      "inline_ref = module_prefix name inline_type_args",
        "inline_type_args? = ::[ type! more_types ]!",
      "array = [ ] | [ exprlist ]!",
      "int = Int",
      "float = Float",
      "formatter = formatter_text formatter_parts",
        "formatter_parts? = .. formatter_part! formatter_parts",
        "formatter_part = formatter_text | char",
        "formatter_text = DqStr",
      "string = string_text string_parts",
        "string_parts? = .. string_part! string_parts",
        "string_part = string_text | char",
        "string_text = SqStr",
      "char = Char",
}
