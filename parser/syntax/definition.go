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
    "_tree":  "|->",
}

var Extra = [...] string { "Call", "Get", "Void" }

var Tokens = [...] Token {
    Token { Name: "MulStr",  Pattern: r(`'''([^']|'[^']|''[^'])+'''`) },
    Token { Name: "String",  Pattern: r(`'[^'\n]*'`) },
    Token { Name: "String",  Pattern: r(`"[^"\n]*"`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^\n]*`) },
    Token { Name: "Pragma",  Pattern: r(`#[^\n]*`) },
    Token { Name: "<<",      Pattern: r(`\<\<`) },
    Token { Name: ">>",      Pattern: r(`\>\>`) },
    Token { Name: "Blank",   Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",      Pattern: r(LF+`+`) },
    Token { Name: "LF",      Pattern: r(`;+`) },
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
    Token { Name: "...",     Pattern: r(`\.\.\.`) },
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: "@",       Pattern: r(`@`) },
    Token { Name: "??",      Pattern: r(`\?\?`) },
    Token { Name: "?",       Pattern: r(`\?`) },
    Token { Name: ">=",      Pattern: r(`\>\=`) },
    Token { Name: "<=",      Pattern: r(`\<\=`) },
    Token { Name: "==",      Pattern: r(`\=\=`) },
    Token { Name: "!=",      Pattern: r(`\!\=`) },
    Token { Name: "~~",      Pattern: r(`\~\~`) },
    Token { Name: "!~",      Pattern: r(`\!\~`) },
    Token { Name: "=>",      Pattern: r(`\=\>`) },
    Token { Name: "=",       Pattern: r(`\=`) },
    Token { Name: "->",      Pattern: r(`\-\>`) },
    Token { Name: "<-",      Pattern: r(`\<\-`) },
    Token { Name: "<",       Pattern: r(`\<`) },
    Token { Name: ">",       Pattern: r(`\>`) },
    Token { Name: "!",       Pattern: r(`\!`) },
    Token { Name: "&&",      Pattern: r(`\&\&`) },
    Token { Name: "|->",     Pattern: r(`\|\-\>`) },
    Token { Name: "||",      Pattern: r(`\|\|`) },
    Token { Name: "~",       Pattern: r(`~`) },
    Token { Name: "&",       Pattern: r(`\&`) },
    Token { Name: "|",       Pattern: r(`\|`) },
    Token { Name: `\`,       Pattern: r(`\\`) },
    Token { Name: "+",       Pattern: r(`\+`) },
    Token { Name: "-",       Pattern: r(`\-`) },
    Token { Name: "**",      Pattern: r(`\*\*`) },
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
    "@if", "@else", "@elif", "@switch", "@otherwise",
    "@while", "@for", "@in", "@break", "@continue", "@return",
    "@yield", "@await",
    "@let", "@type", "@new", "@singleton", "@initial",
    "@set", "@do", "@nothing",
    "@function", "@async", "@lambda", "@$",
    "@invoke", "@iterator", "@promise", "@observer",
    "@static", "@mock", "@handle",
    "@throw", "@assert", "@ensure", "@try", "@to",
    "@unless", "@failed", "@finally", "@panic",
    "@where", "@when", "@match", "@tree",
    "@struct", "@immutable", "@overload", "@operator",
    "@enum",
    "@mount", "@push",
    "@class", "@init", "@create", "@private", "@data", "@interface",
    "@str", "@len", "@copy", "@negate",
    "@prms", "@iter", "@obsv", "@async_iter",
    "@true", "@false",
    "@is", "@or", "@not",
}


/* Infix Operators */
var Operators = [...] Operator {
    /* Fallback */
    Operator { Match: "??",   Priority: 60,  Assoc: Left,   Lazy: true   },
    /* Comparison */
    Operator { Match: "<",    Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">",    Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "<=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "==",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "~~",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!~",   Priority: 50,  Assoc: Left,   Lazy: false  },
    /* Logic */
    Operator { Match: "&&",   Priority: 40,  Assoc: Left,   Lazy: true   },
    Operator { Match: "||",   Priority: 30,  Assoc: Left,   Lazy: true   },
    /* Arithmetic */
    Operator { Match: "+",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "-",    Priority: 70,  Assoc: Left,   Lazy: false  },
    Operator { Match: "*",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "/",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "%",    Priority: 80,  Assoc: Left,   Lazy: false  },
    Operator { Match: "^",    Priority: 90,  Assoc: Right,  Lazy: false  },
}


var RedefinableOperators = []string {
    "==", "<",
    "@negate", "+", "-", "*", "/", "%", "^",
}


var SyntaxDefinition = [...] string {

    /*** Root ***/
    "module = @module name! export imports decls commands",
    "name = Name",
    "export? = @export { namelist! }! | @export namelist!",
    "imports? = import imports",
    "import = import_all | import_names | import_module",
    "import_names = @import as_list @from name",
    "import_module = @import as_item",
    "import_all = @import * @from name",
    "as_list = as_item as_list_tail",
    "as_list_tail? = , as_item! as_list_tail",
    "as_item = name @as name! | name",
    "namelist = name namelist_tail",
    "namelist_tail? = , name! namelist_tail",
    /* Eval */
    "eval = decls commands",

    /*** Declaration ***/
    "decls? = decl decls",
    "decl_type = decl_singleton | decl_type_alias",
    "decl_singleton = @type name = @new @singleton",
    "decl_type_alias = @type name type_params = type",
    "decl_define = function | schema | enum | class | interface",

    /*** Type Parameters ***/
    "type_params? = [ weak_list ]!",
    "weak_list = weak_list_item weak_list_tail",
    "weak_list_tail? = , weak_list_item! weak_list_tail",
    "weak_list_item = name : type! | name",

    /*** Type Expression ***/
    "type_expr = @type { type }",
    "type = union",
    "union = intersection union_tail",
    "union_tail? = | intersection! union_tail",
    "intersection = type_operand intersection_tail",
    "intersection_tail? = & type_operand! intersection_tail",
    "type_operand = name field_gets type_args | fun_sig | ( type )!",
    "field_gets? = field_get field_gets",
    "field_get = Get . name!",
    "type_args? = [ type_arglist ]!",
    "type_arglist = type type_arglist_tail",
    "type_arglist_tail? = , type! type_arglist_tail",
    "fun_sig = [ -> type! ]! | [ typelist! ->! type! ]!",
    "typelist = type typelist_tail",
    "typelist_tail? = , type typelist_tail",

    /*** Function ***/
    "function = f_overload | f_single",
    "f_single = @function name type_params paralist! ->! type! body!",
    "f_overload = @function name type_params { f_item_list }",
    "f_item_list = f_item f_item_list_tail",
    "f_item_list_tail? = , f_item! f_item_list_tail",
    "f_item = paralist! ->! type! body!",
    "paralist = ( ) | ( typed_list! )!",
    "typed_list = typed_list_item typed_list_tail",
    "typed_list_tail? = , typed_list_item! typed_list_tail",
    "typed_list_item = name :! type!",
    "body = { static_commands commands mock_hook handle_hook }!",
    "static_commands? = @static { commands }",
    "mock_hook? = ... @mock name! { commands }",
    "handle_hook? = ... @handle name { handle_cmds }! finally",
    "handle_cmds? = handle_cmd handle_cmds",
    "handle_cmd = unless | failed",
    "unless = @unless name unless_para { commands }",
    "unless_para? = Call ( typed_list! )!",
    "failed = @failed opt_to name { commands }",
    "finally? = @finally { commands }",
    /* Lambda */
    "lambda = quick_lambda | @lambda paralist_weak ret_weak body_flex",
    "paralist_weak? = name | ( ) | ( weak_list! )!",
    "ret_weak? = -> type",
    "body_flex = => expr! | => body! | body!",
    "quick_lambda = [ namelist :! expr! ]! | [ expr! ]!",
    /* IIFE */
    "iife = @invoke body",
    /* Operator Overloading */
    "overload = @overload { operators_defs }!",
    "operator_defs? = operator_def operator_defs",
    "operator_def = @operator general_op operator_def_fun",
    "general_op = operator | unary",
    "operator_def_fun = (! namelist! )! body!",

    /*** Command ***/
    "commands? = command commands",
    "command = cmd_group1 | cmd_group2 | cmd_group3",
    "cmd_group1 = cmd_branch | cmd_loop | cmd_loop_ctrl",
    "cmd_group2 = cmd_pause | cmd_abrupt | cmd_return | cmd_guard",
    "cmd_group3 = cmd_scope | cmd_set | cmd_side_effect | cmd_pass",
    /* Conditional Brach @ Group 1 */
    "cmd_branch = cmd_if | cmd_switch",
    "block = { commands }!",
    "cmd_if = @if expr! block! elifs else",
    "elifs? = elif elifs",
    "elif = @else @if expr! block! | @elif expr! block!",
    "else? = @else block!",
    "cmd_switch = switch_when | switch_match",
    "switch_when = @switch @when {! cases! }!",
    "switch_match = @switch @match expr {! cases! }!",
    "cases = case more_cases",
    "more_cases? = case more_cases",
    "case = @otherwise block! | expr block!",
    /* Loop @ Group 1 */
    "cmd_loop = cmd_while | cmd_for",
    "cmd_while = @while expr! block!",
    "cmd_for = @for for_params! @in expr! block!",
    "for_params = for_params_list | for_params_hash | for_params_value",
    "for_params_list = for_value [ for_index! ]!",
    "for_params_hash = { for_key :! for_value! }!",
    "for_params_value = for_value",
    "for_value = name",
    "for_index = name",
    "for_key = name",
    /* Loop Control @ Group 1 */
    "cmd_loop_ctrl = @break | @continue",
    /* Execution Pause @ Group 2 */
    "cmd_pause = cmd_yield | cmd_async_for | cmd_await",
    "cmd_yield = @yield pattern = expr! | @yield expr!",
    "cmd_await = @await pattern = expr! | @await expr!",
    "cmd_async_for = @await name @in expr! block!",
    /* Execution Abrupt @ Group 2 */
    "cmd_abrupt = cmd_return | cmd_throw | cmd_panic",
    "cmd_throw = @throw expr!",
    "cmd_panic = @panic expr!",
    /* Return @ Group 2 */
    "cmd_return = @return Void | @return expr",
    /* Guard @ Group 2 */
    "cmd_guard = cmd_ensure | cmd_try | cmd_assert",
    "cmd_ensure = @ensure name! ensure_args { expr! }!",
    "ensure_args? = Call ( exprlist )",
    "exprlist = expr exprlist_tail",
    "exprlist_tail? = , expr! exprlist_tail",
    "cmd_try = @try opt_to name { commands }!",
    "opt_to? = @to",
    "cmd_assert = @assert expr!",
    /* Scope Operation @ Group 3 */
    "cmd_scope = cmd_let | cmd_initial | cmd_rebind",
    "cmd_let = @let pattern = expr!",
    "cmd_initial = @initial name var_type = expr!",
    "var_type? = : type",
    "cmd_rebind = @set name set_op = expr",
    "set_op? = op_arith",
    "pattern = name | { namelist }!",
    /* Setter Emitter @ Group 3 */
    "cmd_set = @set left_val set_op = expr",
    "left_val = operand",
    /* Side-Effect Expression @ Group 3 */
    "cmd_side_effect = expr",
    /* Pass @ Group 3 */
    "cmd_pass = @do @nothing",

    /* Schema */
    "schema = schema_kind name type_params bases {! field_list! overload }!",
    "schema_kind = @struct | @immutable",
    "bases? = : typelist!",
    "field_list = field field_list_tail",
    "field_list_tail? = , field! field_list_tail",
    "field = name : type! field_default",
    "field_default? = = expr",
    /* Enum */
    "enum = @enum name {! enum_item_list }!",
    "enum_item_list = enum_item enum_item_list_tail",
    "enum_item_list_tail? = , enum_item enum_item_list_tail",
    "enum_item = schema | name",

    /*** Object-Oriented ***/
    /* Class */
    "class = @class name type_params bases { init pfs methods overload }",
    "init = @init paralist_strict! body! creators",
    "creators? = creator creators",
    "creator = @create paralist_strict! body!",
    "pfs? = pf pfs",
    "pf = @private name paralist_strict! ->! type! body!",
    "methods? = method methods",
    "method = name paralist_strict! ->! type! body!",
    /* Interface */
    "interface = @interface name type_params { method_protos }!",
    "method_protos? = method_proto method_protos",
    "method_proto = name paralist_strict! ->! type!",

    /*** Expression ***/
    "expr = lower_unary operand expr_tail | operand expr_tail",
    "lower_unary = @mount | @push",
    "expr_tail? = operator operand! expr_tail",
    /* Operators (Infix) */
    "operator = op_fallback | op_compare | op_logic | op_arith",
    "op_fallback = ?? ",
    "op_compare = < | > | <= | >= | == | != | ~~ | !~ ",
    `op_logic = && | _bar2 `,
    "op_arith = + | - | * | / | % | ^ ",
    /* Operators (Prefix) */
    "unary? = @not | _exc | - ",
    /* Operand */
    "operand = operand_body calls",
    "operand_body = unary operand_base operand_tail",
    "operand_base = wrapped | primitive | new | brace_expr | identifier",
    "wrapped = ( expr! )!",
    "identifier = Name",
    "new = @new type Call ( arglist )! | @new type literal!",
    "literal = : list | map | struct | body",
    "operand_tail? = tail_operation operand_tail",
    "tail_operation = field_get | slice | get | inflate",
    "inflate = :: [ typelist! ]!",
    "slice = Get [ low_bound : high_bound ]!",
    "low_bound? = expr",
    "high_bound? = expr",
    "get = Get [ exprlist! ]! nil_flag",
    "nil_flag? = ?",
    "calls = plain_calls pipelines",
    "plain_calls? = plain_call plain_calls",
    "plain_call = Call ( arglist )!",
    "arglist? = exprlist",
    "pipelines? = pipeline pipelines",
    "pipeline = => operand_body ( p_arglist )!",
    "p_arglist? = p_arg_list",
    "p_arg_list = p_arg p_arg_list_tail",
    "p_arg_list_tail? = , p_arg p_arg_list_tail",
    "p_arg = @$ | expr",

    /*** Literal ***/
    "brace_expr = when | match | iife | tree | lambda",

    "when = @when {! branch_list }!",
    "branch_list = branch! branch_list_tail",
    "branch_list_tail? = , branch! branch_list_tail",
    "branch = @otherwise :! expr! | expr! :! expr!",
    "match = @match expr {! branch_list }!",
    "observer = @observer body",
    "tree = _tree tree_node! | @tree { tree_node! }!",
    "tree_node = type { node_props node_children }!",
    "node_props? = node_prop node_props",
    "node_prop = name = expr! | string = expr!",
    "node_children? = node_child node_children",
    "node_child = tree_node | = expr",

    "struct = { } | { struct_item struct_item_tail }!",
    "struct_item_tail? = , struct_item struct_item_tail",
    "struct_item = name : expr! | :: name!",
    /* Hash Table */
    "map = { } | { map_item! map_tail }!",
    "map_tail? = , map_item! map_tail",
    "map_item = expr =>! expr",
    /* Linear List */
    "list = [ ] | [ list_item! list_tail ]!",
    "list_tail? = , list_item list_tail",
    "list_item = expr",
    /* Primitive */
    "primitive = string | number | bool",
    "string = String | MulStr",
    "number = Hex | Exp | Dec | Int",
    "bool = @true | @false",

}
