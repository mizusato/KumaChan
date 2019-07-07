package syntax

import "regexp"
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `;\{\}\[\]\(\)\.\,\:@\?\<\>\=\!~\&\|\\\+\-\*\/%\^'"` + "`"

var EscapeMap = map [string] string {
    "_exc":   "!",
    "_bar1":  "|",
    "_bar2":  "||",
    "_at":    "@",
}

var Extra = [...] string { "Call", "Get", "Void" }

var Tokens = [...] Token {
    Token { Name: "MulStr",  Pattern: r(`'''([^']|'[^']|''[^'])+'''`) },
    Token { Name: "String",  Pattern: r(`'[^'\n]*'`) },
    Token { Name: "String",  Pattern: r(`"[^"\n]*"`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^\n]*`) },
    Token { Name: "<<",      Pattern: r(`\<\<[`+Blanks+`]`) },
    Token { Name: ">>",      Pattern: r(`[`+Blanks+`]\>\>`) },
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
    Token { Name: "<=>",     Pattern: r(`\<\=\>`) },
    Token { Name: "<!=>",    Pattern: r(`\<!\=\>`) },
    Token { Name: ">=",      Pattern: r(`\>\=`) },
    Token { Name: "<=",      Pattern: r(`\<\=`) },
    Token { Name: "==",      Pattern: r(`\=\=`) },
    Token { Name: "!=",      Pattern: r(`\!\=`) },
    Token { Name: "~~",      Pattern: r(`\~\~`) },
    Token { Name: "!~",      Pattern: r(`\!\~`) },
    Token { Name: "=>",      Pattern: r(`\=\>`) },
    Token { Name: "=",       Pattern: r(`\=`) },
    Token { Name: "-->",     Pattern: r(`\-\-\>`) },
    Token { Name: "<--",     Pattern: r(`\<\-\-`) },
    Token { Name: "->",      Pattern: r(`\-\>`) },
    Token { Name: "<-",      Pattern: r(`\<\-`) },
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


/* Conditional Keywords */
var Keywords = [...] string {
    "@module", "@export", "@include", "@import", "@from", "@as",
    "@if", "@else", "@elif", "@switch", "@case", "@default",
    "@while", "@for", "@in", "@break", "@continue", "@return",
    "@yield", "@await",
    "@let", "@type", "@singleton", "@var", "@reset",
    "@set", "@do", "@nothing",
    "@function", "@async", "@lambda",
    "@invoke", "@iterator", "@promise",
    "@static", "@mock", "@handle",
    "@throw", "@assert", "@ensure", "@try", "@to",
    "@unless", "@failed", "@finally", "@panic",
    "@where", "@when", "@otherwise",
    "@struct", "@fields", "@config", "@operator", "@guard",
    "@one", "@of", "@enum", "@$",
    "@class", "@init", "@mount", "@create", "@private", "@data", "@interface",
    "@str", "@len", "@prms", "@iter", "@async_iter", "@negate",
    "@true", "@false",
    "@is", "@or", "@not",
}


/* Infix Operators */
var Operators = [...] Operator {
    /* Pull, Push, Derive, Otherwise */
    Operator { Match: "<<",   Priority: 20,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">>",   Priority: 20,  Assoc: Left,   Lazy: false  },
    Operator { Match: "=>",   Priority: 20,  Assoc: Left,   Lazy: true   },
    Operator { Match: "@or",  Priority: 20,  Assoc: Left,   Lazy: true   },
    /* Comparison */
    Operator { Match: "<",    Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">",    Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "<=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "==",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "~~",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!~",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "<=>",  Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "<!=>", Priority: 50,  Assoc: Left,   Lazy: false  },
    /* Logic */
    Operator { Match: "@is",  Priority: 60,  Assoc: Left,   Lazy: false  },
    Operator { Match: "@as",  Priority: 60,  Assoc: Left,   Lazy: false  },
    Operator { Match: "&&",   Priority: 40,  Assoc: Left,   Lazy: true   },
    Operator { Match: "||",   Priority: 30,  Assoc: Left,   Lazy: true   },
    Operator { Match: "&",    Priority: 90,  Assoc: Left,   Lazy: false  },
    Operator { Match: "|",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: `\`,    Priority: 70,  Assoc: Left,   Lazy: false  },
    /* Arithmetic */
    Operator { Match: "+",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "-",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "*",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "/",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "%",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "^",    Priority: 90,  Assoc: Right,  Lazy: false  },
}


var RedefinableOperators = []string {
    "@as", "@str", "@len", "@prms", "@iter", "@async_iter", "@enum",
    "==", "<",
    "<<", ">>",
    "@negate", "+", "-", "*", "/", "%", "^",
}


var SyntaxDefinition = [...] string {

    /*** Root ***/
    "module = @module name! export includes commands",
    "name = Name",
    "export? = @export { namelist! }! | @export namelist!",
    "includes? = include includes",
    "include = @include string",
    "namelist = name namelist_tail",
    "namelist_tail? = , name! namelist_tail",
    /* Included */
    "included = includes commands",
    /* Eval */
    "eval = commands",

    /*** Command ***/
    "commands? = command commands",
    "command = cmd_group1 | cmd_group2 | cmd_group3",
    "cmd_group1 = cmd_flow | cmd_pause | cmd_err | cmd_return",
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
    /* Pause Commands @ Group 1 */
    "cmd_pause = cmd_yield | cmd_async_for | cmd_await",
    "cmd_yield = @yield name var_type = expr! | @yield expr!",
    "cmd_await = @await name var_type = expr! | @await expr!",
    "cmd_async_for = @await name @in expr! block!",
    /* Error Related Commands @ Group 1 */
    "cmd_err = cmd_throw | cmd_assert | cmd_panic | cmd_ensure | cmd_try",
    "cmd_throw = @throw expr!",
    "cmd_assert = @assert expr!",
    "cmd_panic = @panic expr!",
    "cmd_ensure = @ensure name! ensure_args { expr! }!",
    "ensure_args? = Call ( exprlist )",
    "cmd_try = @try opt_to name { commands }!",
    "opt_to? = @to",
    /* Return Command @ Group 1 */
    "cmd_return = @return Void | @return expr",
    /* Module Related Commands @ Group 2 */
    "cmd_module = cmd_import",
    "cmd_import = import_all | import_names | import_module",
    "import_names = @import as_list @from name",
    "import_module = @import as_item",
    "import_all = @import * @from name",
    "as_list = as_item as_list_tail",
    "as_list_tail? = , as_item! as_list_tail",
    "as_item = name @as name! | name",
    /* Scope Related Commands @ Group 2 */
    "cmd_scope = cmd_let | cmd_type | cmd_var | cmd_reset",
    "cmd_let = @let name var_type = expr",
    "cmd_type = @type name = @singleton | @type name generic_params = expr",
    "cmd_var = @var name var_type = expr",
    "var_type? = : type",
    "cmd_reset = @reset name set_op = expr",
    "set_op? = op_arith",
    /* Definition Commands @ Group 2 */
    "cmd_def = function | schema | enum | class | interface",
    /* Pass Command @ Group 3 */
    "cmd_pass = @do @nothing",
    /* Set Command @ Group 3 */
    "cmd_set = @set left_val set_op = expr",
    "left_val = operand_base gets!",
    "gets? = get gets",
    /* Exec Command @ Group 3 */
    "cmd_exec = expr",

    /*** Function ***/
    "function = f_sync",
    "f_sync = @function name Call paralist_strict! ->! type! body!",
    "paralist_strict = ( ) | ( typed_list! )!",
    "typed_list = typed_list_item typed_list_tail",
    "typed_list_tail? = , typed_list_item! typed_list_tail",
    "typed_list_item = name :! type!",
    "type = fun_sig | type_expr | ( expr )!",
    "fun_sig = @$ Call < opt_typelist >! <! type! >!",
    "opt_typelist? = typelist",
    "typelist = type typelist_tail",
    "typelist_tail? = , type typelist_tail",
    "type_expr = identifier type_gets type_args",
    "type_gets? = type_get type_gets",
    "type_get = Get . name",
    "type_args? = Call < type_arglist! >!",
    "type_arglist = type_arg type_arglist_tail",
    "type_arglist_tail? = , type_arg! type_arglist_tail",
    "type_arg = type | primitive",
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
    /* Lambda */
    "lambda = lambda_block | lambda_inline",
    "lambda_block = @lambda paralist_block ret_lambda body!",
    "paralist_block? = name | Call paralist",
    "paralist = ( ) | ( namelist ) | ( typed_list! )!",
    "ret_lambda? = -> type | ->",
    "opt_arrow? = ->",
    "lambda_inline = .{ paralist_inline expr! }!",
    "paralist_inline? = namelist --> | ( namelist ) -->",
    /* IIFE */
    "iife = invoke | iterator | promise | async_iterator",
    "invoke = @invoke body",
    "iterator = @iterator body",
    "promise = @promise body",
    "async_iterator = @async @iterator body!",
    /* Operator Overloading */
    "operator_defs? = operator_def operator_defs",
    "operator_def = @operator general_op operator_def_fun",
    "general_op = operator | unary",
    "operator_def_fun = (! namelist! )! opt_arrow body!",

    /*** Type ***/
    "generic_params? = < namelist > | < typed_list! >!",
    "type_literal = simple_type_literal | finite_literal",
    /* Schema */
    "schema = @struct name generic_params { field_list }! schema_config",
    "field_list = field field_list_tail",
    "field_list_tail? = , field! field_list_tail",
    "field = name : type! field_default | @fields @of type!",
    "field_default? = = expr",
    "schema_config? = @config { struct_guard operator_defs }!",
    "struct_guard? = @guard body!",
    /* Enum */
    "enum = @enum name {! namelist! }!",
    /* Finite Set */
    "finite_literal = @one @of { exprlist }! | { exprlist }",
    /* SimpleType */
    "simple_type_literal = { name _bar1 filters! }!",
    "filters = exprlist",

    /*** Object-Oriented ***/
    /* Class */
    "class = @class name generic_params supers { init pfs methods class_opt }",
    "supers? = @is typelist",
    "init = @init Call paralist_strict! body! creators",
    "creators? = creator creators",
    "creator = @create Call paralist_strict! body!",
    "pfs? = pf pfs",
    "pf = @private name Call paralist_strict! ->! type! body!",
    "methods? = method methods",
    "method = name Call paralist_strict! ->! type! body!",
    "class_opt = operator_defs data",
    "data? = @data hash",
    /* Interface */
    "interface = @interface name generic_params { members }",
    "members? = member members",
    "member = method_implemented | method_blank",
    "method_implemented = name Call paralist_strict! ->! type! body",
    "method_blank = name Call paralist_strict! -> type",

    /*** Expression ***/
    "expr = operand expr_tail",
    "expr_tail? = operator operand! expr_tail",
    /* Operators (Infix) */
    "operator = op_group1 | op_compare | op_logic | op_arith",
    "op_group1 = << | >> | => | @or",
    "op_compare = < | > | <= | >= | == | != | ~~ | !~ | <=> | <!=> ",
    `op_logic = @is | @as | && | _bar2 | & | _bar1 | \ `,
    "op_arith = + | - | * | / | % | ^ ",
    /* Operators (Prefix) */
    "unary? = unary_group1 | unary_group2 | unary_group3 | unary_group4",
    "unary_group1 = @not | - | @negate | _exc | ~ ",
    "unary_group2 = @str | @len",
    "unary_group3 = @prms | @iter | @async_iter | @enum",
    "unary_group4 = @mount",
    /* Operand */
    "operand = unary operand_base operand_tail",
    "operand_base = wrapped | literal | dot_para | identifier",
    "wrapped = ( expr! )!",
    "operand_tail? = slice operand_tail | get operand_tail | call operand_tail",
    "dot_para = . Name",
    "identifier = Name",
    "slice = Get [ low_bound : high_bound ]!",
    "low_bound? = expr",
    "high_bound? = expr",
    "get = get_expr | get_name",
    "call = call_self | call_method",
    "get_expr = Get [ expr! ]! nil_flag",
    "get_name = Get . name! nil_flag",
    "nil_flag? = ?",
    "call_self = Call args",
    "call_method = -> name method_args",
    "method_args = Call args | extra_arg",
    "args = ( arglist )! extra_arg | < type_arglist >",
    "extra_arg? = -> lambda",
    "arglist? = exprlist",
    "exprlist = expr exprlist_tail",
    "exprlist_tail? = , expr! exprlist_tail",

    /*** Literal ***/
    "literal = primitive | adv_literal",
    "adv_literal = fun_sig | comp | type_literal | list | hash | brace_literal",
    "brace_literal = when | iife | lambda | struct",
    "when = @when { when_list }!",
    "when_list = when_item when_list_tail",
    "when_item = @otherwise : expr | expr : expr",
    "when_list_tail? = , when_item when_list_tail",
    "struct = type struct_hash",
    "struct_hash = { } | { struct_hash_item struct_hash_tail }!",
    "struct_hash_tail? = , struct_hash_item struct_hash_tail",
    "struct_hash_item = name : expr! | :: name!",
    /* Hash Table */
    "hash = { } | { hash_item! hash_tail }!",
    "hash_tail? = , hash_item! hash_tail",
    "hash_item = name :! expr! | string :! expr! | :: name!",
    /* Linear List */
    "list = [ ] | [ list_item! list_tail ]!",
    "list_tail? = , list_item list_tail",
    "list_item = expr",
    /* List/Iterator Comprehension */
    "comp = .[ comp_rule! ]! | [ comp_rule ]!",
    "comp_rule = expr , @for in_list! opt_filters",
    "opt_filters? = , @where exprlist",
    "in_list = in_item in_list_tail",
    "in_list_tail? = , in_item! in_list_tail",
    "in_item = name @in expr",
    /* Primitive */
    "primitive = string | number | bool",
    "string = String | MulStr",
    "number = Hex | Exp | Dec | Int",
    "bool = @true | @false",

}
