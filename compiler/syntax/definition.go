package syntax

import "regexp"
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `;\{\}\[\]\(\)\.\,\:@\?\<\>\=\!~\&\|\\\+\-\*\/%\^`

var EscapeMap = map [string] string {
    "_exc":   "!",
    "_bar1":  "|",
    "_bar2":  "||",
    "_at":    "@",
}

var Extra = [...] string { "Call", "Get", "Void" }

var Tokens = [...] Token {
    Token { Name: "String",  Pattern: r(`'[^']*'`), },
    Token { Name: "String",  Pattern: r(`"[^"]*"`), },
    Token { Name: "Raw",     Pattern: r(`/~([^~]|[^/]|~[^/]|[^~]/)*~/`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^\n]*`) },
    Token { Name: "..[",     Pattern: r(`\.\.\[`) },
    Token { Name: "...[",    Pattern: r(`\.\.\.\[`) },
    Token { Name: "..{",     Pattern: r(`\.\.\{`) },
    Token { Name: "...{",    Pattern: r(`\.\.\.\{`) },
    Token { Name: "Comment", Pattern: r(`\.\.[^\[\{][^`+Blanks+LF+`]*`) },
    Token { Name: "Blank",   Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",      Pattern: r(LF+`+`) },
    Token { Name: "LF",      Pattern: r(`;`) },
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
    Token { Name: ".{",      Pattern: r(`\.\{`) },
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: "@",       Pattern: r(`@`) },
    Token { Name: "?",       Pattern: r(`\?`) },
    Token { Name: ">=",      Pattern: r(`\>\=`) },
    Token { Name: "<=",      Pattern: r(`\<\=`) },
    Token { Name: "!==",     Pattern: r(`\!\=\=`) },
    Token { Name: "!=",      Pattern: r(`\!\=`) },
    Token { Name: "===",     Pattern: r(`\=\=\=`) },
    Token { Name: "==",      Pattern: r(`\=\=`) },
    Token { Name: "=>",      Pattern: r(`\=\>`) },
    Token { Name: "=",       Pattern: r(`\=`) },
    Token { Name: "-->",     Pattern: r(`\-\-\>`) },
    Token { Name: "<--",     Pattern: r(`\<\-\-`) },
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
    //    { Name: "Call",    [ Inserted by Scanner ] },
    //    { Name: "Get",     [ Inserted by Scanner ] },
    //    { Name: "Void",    [ Inserted by Scanner ] },
}


var Keywords = [...] string {
    "@module", "@export", "@import", "@use", "@as",
    "@if", "@else", "@elif", "@switch", "@case", "@default",
    "@while", "@for", "@in", "@break", "@continue", "@return",
    "@let", "@var", "@reset",
    "@set", "@do", "@nothing",
    "@function", "@generator", "@async", "@lambda",
    "@static", "@mock", "@handle",
    "@throw", "@assert", "@ensure", "@try", "@to",
    "@unless", "@failed", "@finally", "@panic",
    "@xml", "@map",
    "@struct", "@require", "@one", "@of",
    "@class", "@init", "@data", "@interface", "@expose",
    "@true", "@false",
    "@is", "@or", "@not", "@yield", "@await",
}


var Operators = [...] Operator {
    /* Pull, Push, Derive, Otherwise */
    Operator { Match: "<<",   Priority: 20,  Assoc: Right,  Lazy: false },
    Operator { Match: ">>",   Priority: 20,  Assoc: Left,   Lazy: false  },
    Operator { Match: "=>",   Priority: 20,  Assoc: Left,   Lazy: true  },
    Operator { Match: "@or",  Priority: 20,  Assoc: Left,   Lazy: true  },
    /* Comparison */
    Operator { Match: "<",    Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">",    Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: "<=",   Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">=",   Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: "==",   Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!=",   Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: "===",  Priority: 30,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!==",  Priority: 30,  Assoc: Left,   Lazy: false  },
    /* Logic */
    Operator { Match: "@is",  Priority: 10,  Assoc: Left,   Lazy: false  },
    Operator { Match: "&&",   Priority: 60,  Assoc: Left,   Lazy: true  },
    Operator { Match: "||",   Priority: 50,  Assoc: Left,   Lazy: true  },
    Operator { Match: "&",    Priority: 60,  Assoc: Left,   Lazy: false  },
    Operator { Match: "|",    Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: `\`,    Priority: 40,  Assoc: Left,   Lazy: false  },
    /* Arithmetic */
    Operator { Match: "+",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "-",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "*",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "/",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "%",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "^",    Priority: 90,  Assoc: Right,  Lazy: false },
}


var SyntaxDefinition = [...] string {

    /* Root */
    "program = module | eval",

    /* Module */
    "module = module_declare exports commands",
    "module_declare = @module name!",
    "name = Name",
    "exports? = export exports",
    "export = @export as_list!",
    "as_list = as_item as_list_tail",
    "as_list_tail? = , as_item! as_list_tail",
    "as_item = name @as name! | name",

    /* Eval */
    "eval = commands",

    /* Commands */
    "commands? = command commands",
    "command = cmd_group1 | cmd_group2 | cmd_group3",
    "cmd_group1 = cmd_flow | cmd_yield | cmd_await | cmd_return | cmd_err",
    "cmd_group2 = cmd_module | cmd_scope | cmd_def",
    "cmd_group3 = cmd_pass | cmd_set | cmd_exec",
    /* Flow Control Commands @ Group 1 */
    "cmd_flow = cmd_if | cmd_switch | cmd_while | cmd_for | cmd_loop_ctrl",
    "block = { commands }!",
    "cmd_loop_ctrl = @break | @continue",
    "cmd_if = @if expr! block! elifs else",
    "elifs? = elif elifs",
    "elif = @else @if expr! block! | @elif expr! block!",
    "else? = @else block!",
    "cmd_switch = @switch { cases default }!",
    "cases? = case cases",
    "case = @case expr! block!",
    "default? = @default block!",
    "cmd_while = @while expr! block!",
    "cmd_for = @for for_params! @in expr! block!",
    "for_params = for_params_list | for_params_hash | for_params_value",
    "for_params_list = for_value [ for_index! ]!",
    "for_params_hash = { for_key :! for_value! }!",
    "for_params_value = for_value",
    "for_value = name",
    "for_index = name",
    "for_key = name",
    /* Yield Command @ Group 1 */
    "cmd_yield = @yield name var_type = expr! | @yield expr!",
    /* Await Command @ Group 1 */
    "cmd_await = @await name var_type = expr! | @await expr!",
    /* Return Command @ Group 1 */
    "cmd_return = @return Void | @return expr",
    /* Error Related Commands @ Group 1 */
    "cmd_err = cmd_throw | cmd_assert | cmd_panic | cmd_ensure | cmd_try",
    "cmd_throw = @throw expr!",
    "cmd_assert = @assert expr!",
    "cmd_panic = @panic expr!",
    "cmd_ensure = @ensure name! ensure_args { expr! }!",
    "ensure_args? = Call ( exprlist )",
    "cmd_try = @try opt_to name { commands }!",
    "opt_to? = @to",
    /* Module Related Commands @ Group 2 */
    "cmd_module = cmd_use | cmd_import",
    "cmd_use = @use as_list",
    "cmd_import = @import namelist",
    "namelist = name namelist_tail",
    "namelist_tail? = , name! namelist_tail",
    /* Scope Related Commands @ Group 2 */
    "cmd_scope = cmd_let | cmd_var | cmd_reset",
    "cmd_let = @let name var_type = expr",
    "cmd_var = @var name var_type = expr",
    "var_type? = : type",
    "cmd_reset = @reset name set_op = expr",
    "set_op? = op_arith",
    /* Definition Commands @ Group 2 */
    "cmd_def = function | abs_def",
    "abs_def = struct | class | interface",
    /* Pass Command @ Group 3 */
    "cmd_pass = @do @nothing",
    /* Set Command @ Group 3 */
    "cmd_set = @set left_val set_op = expr",
    "left_val = operand_base gets!",
    "gets? = get gets",
    /* Exec Command @ Group 3 */
    "cmd_exec = expr",

    /* Function Definition */
    "function = f_sync | f_async | generator",
    "f_sync = @function name Call paralist_strict! -> type body",
    "f_async = @async name Call paralist_strict! -> type body",
    "generator = @generator name Call paralist_strict! body",
    "paralist_strict = ( ) | ( typed_list! )!",
    "typed_list = typed_list_item typed_list_tail",
    "typed_list_tail? = , typed_list_item! typed_list_tail",
    "typed_list_item = name :! type!",
    "type = identifier type_gets type_args | ( expr )",
    "type_gets? = type_get type_gets",
    "type_get = Get . name",
    "type_args? = Call < type_arglist! >!",
    "type_arglist = type_arg type_arglist_tail",
    "type_arglist_tail? = , type_arg! type_arglist_tail",
    "type_arg = type | primitive",
    // note: the rule name "body" has sth to do with Block()
    "body = { static_commands commands mock_hook handle_hook }!",
    "static_commands? = @static { commands }",
    "mock_hook? = _at @mock name! { commands }",
    "handle_hook? = _at @handle name { handle_cmds }! finally",
    "handle_cmds? = handle_cmd handle_cmds",
    "handle_cmd = unless | failed",
    "unless = @unless name unless_para { commands }",
    "unless_para? = Call ( namelist )",
    "failed = @failed opt_to name { commands }",
    "finally? = _at @finally { commands }",

    /* Type Object Definition */
    /* Generics */
    "generic_params = < typed_list! >!",
    /* Schema */
    "struct = @struct name generic_params { field_list condition }",
    "field_list = field field_list_tail",
    "field_list_tail? = , field! field_list_tail",
    "field = name :! type! field_default",
    "field_default? = = expr",
    "condition = , @require expr",
    /* Finite */
    "finite_literal = @one @of { exprlist_opt }! | { exprlist }",
    "exprlist_opt? = exprlist",
    /* SimpleType */
    "type_literal = { name _bar1 filters! }!",
    "filters = exprlist",
    /* Class */
    "class = @class name generic_params supers { initializer methods } data",
    "supers? = @is exprlist",
    "initializer = @init Call paralist_strict! body!",
    "methods? = method methods",
    "method = method_proto body!",
    "method_proto = name Call paralist_strict! -> type",
    "data? = @data hash",
    /* Interface */
    "interface = @interface name { members }",
    "members? = member members",
    "member = method_proto | method",

    /* Expression */
    "expr = operand expr_tail",
    "expr_tail? = operator operand! expr_tail",
    /* Operators (Infix) */
    "operator = op_group1 | op_compare | op_logic | op_arith",
    "op_group1 = << | >> | => | @or",
    "op_compare = < | > | <= | >= | == | != | === | !== ",
    `op_logic = @is | && | _bar2 | & | _bar1 | \ `,
    "op_arith = + | - | * | / | % | ^ ",
    /* Operators (Prefix) */
    "unary? = @not | - | _exc | ~ | @expose",
    /* Operand */
    "operand = unary operand_base operand_tail",
    "operand_base = wrapped | lambda | literal | dot_para | identifier",
    "wrapped = ( expr! )!",
    "operand_tail? = get operand_tail | call operand_tail",
    "dot_para = . Name",
    "identifier = Name",
    "get = get_expr | get_name",
    "call = call_self | call_method",
    "get_expr = Get [ expr! ]! nil_flag",
    "get_name = Get . name! nil_flag",
    "nil_flag? = ?",
    "call_self = Call args",
    "call_method = -> name method_args",
    "method_args = Call args | extra_arg",
    "args = ( arglist )! extra_arg | < type_arglist >",
    "extra_arg? = -> lambda | -> adv_literal",
    "arglist? = exprlist",
    "exprlist = expr exprlist_tail",
    "exprlist_tail? = , expr! exprlist_tail",

    /* Lambda */
    "lambda = lambda_block | lambda_inline",
    "lambda_block = header_lambda paralist_block ret_lambda body!",
    "header_lambda = @lambda | @async | @generator",
    "paralist_block? = name | Call paralist",
    "paralist = ( ) | ( namelist ) | ( typed_list! )!",
    "ret_lambda? = -> type | ->",
    "lambda_inline = .{ paralist_inline expr! }!",
    "paralist_inline? = namelist -->",

    /* Literals */
    "literal = primitive | adv_literal",
    "adv_literal = xml | comprehension | abs_literal | map | list | hash",
    "abs_literal = type_literal | finite_literal",
    // TODO: generics and array
    "map = @map { } | @map { map_item! map_tail }",
    "map_tail? = , map_item! map_tail",
    "map_item = map_key :! expr!",
    "map_key = expr",
    // TODO: when expression: when { cond1: val1, cond2: val2 }
    /* Hash Table */
    "hash = { } | { hash_item! hash_tail }!",
    "hash_tail? = , hash_item! hash_tail",
    "hash_item = name :! expr! | :: name!",
    /* Linear List */
    "list = [ ] | [ list_item! list_tail ]!",
    "list_tail? = , list_item! list_tail",
    "list_item = expr",
    /* List/Iterator Comprehension */
    "comprehension = .[ comp_rule! ]! | [ comp_rule ]!",
    "comp_rule = expr _bar1 in_list! opt_filters",
    "opt_filters? = exprlist",
    "in_list = in_item in_list_tail",
    "in_list_tail? = , in_item! in_list_tail",
    "in_item = name @in expr",
    /* XML Expression */
    "xml = @xml { xml_elements }!",
    "xml_elements? = xml_element xml_elements",
    "xml_element = xml_tag | xml_inner_expr",
    "xml_inner_expr = { expr! }! | identifier | primitive | adv_literal",
    "xml_tag = < xml_name xml_props xml_tag_tail",
    "xml_name = Name xml_name_tail",
    "xml_name_tail? = - Name xml_name_tail",
    "xml_props? = xml_prop xml_props",
    "xml_prop = xml_name =! xml_inner_expr!",
    "xml_tag_tail = xml_tag_tail_empty | xml_tag_tail_full",
    "xml_tag_tail_empty = / >!",
    "xml_tag_tail_full = >! xml_elements <! /! name! >!",
    /* Primitive Values */
    "primitive = string | number | bool",
    "string = String",
    "number = Hex | Exp | Dec | Int",
    "bool = @true | @false",

}
