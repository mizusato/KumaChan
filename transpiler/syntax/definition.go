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

var Extra = [...] string { "Call", "Get" }

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
    Token { Name: "!=",      Pattern: r(`\!\=`) },
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
}


var Keywords = [...] string {
    "@module", "@export", "@import", "@use", "@as",
    "@if", "@elif", "@else", "@switch", "@case", "@while", "@for", "@in",
    "@break", "@continue", "@return",
    "@let", "@var", "@reset",
    "@exec", "@set",
    "@lazy", "@async", "@lambda",
    "@local", "@upper", "@global", "@static",
    "@throw", "@assert", "@require", "@try",
    "@handle", "@unless", "@failed", "@finally",
    "@category", "@struct", "@require", "@enum", "@concept",
    "@class", "@init", "@interface", "@expose",
    "@true", "@false",
    "@is", "@not", "@yield", "@await",
}


/**
 *  Operator Info
 *
 *  If the 'Name' field starts with "_", the operator will be non-overloadable.
 *  If the 'Name' field starts with "*", the operator will be lazily-evaluated.
 *  The value of the 'Priority' field must be non-negative.
 */
var Operators = [...] Operator {
    /* Oriented */
    Operator { Match: "<<",   Name: "pull",    Priority: 20,  Assoc: Right },
    Operator { Match: ">>",   Name: "push",    Priority: 20,  Assoc: Left  },
    Operator { Match: "=>",   Name: "derive",  Priority: 20,  Assoc: Left  },
    /* Comparison */
    Operator { Match: "<",    Name: "lt",      Priority: 30,  Assoc: Left  },
    Operator { Match: ">",    Name: "gt",      Priority: 30,  Assoc: Left  },
    Operator { Match: "<=",   Name: "lte",     Priority: 30,  Assoc: Left  },
    Operator { Match: ">=",   Name: "gte",     Priority: 30,  Assoc: Left  },
    Operator { Match: "==",   Name: "eq",      Priority: 30,  Assoc: Left  },
    Operator { Match: "!=",   Name: "neq",     Priority: 30,  Assoc: Left  },
    /* Logic */
    Operator { Match: "@is",  Name: "_is",     Priority: 10,  Assoc: Left  },
    Operator { Match: "&&",   Name: "*and",    Priority: 60,  Assoc: Left  },
    Operator { Match: "||",   Name: "*or",     Priority: 50,  Assoc: Left  },
    Operator { Match: "&",    Name: "ins",     Priority: 60,  Assoc: Left  },
    Operator { Match: "|",    Name: "uni",     Priority: 50,  Assoc: Left  },
    Operator { Match: `\`,    Name: "diff",    Priority: 40,  Assoc: Left  },
    /* Arithmetic */
    Operator { Match: "+",    Name: "plus",    Priority: 70,  Assoc: Left  },
    Operator { Match: "-",    Name: "minus",   Priority: 70,  Assoc: Left  },
    Operator { Match: "*",    Name: "times",   Priority: 80,  Assoc: Left  },
    Operator { Match: "/",    Name: "divide",  Priority: 80,  Assoc: Left  },
    Operator { Match: "%",    Name: "modulo",  Priority: 80,  Assoc: Left  },
    Operator { Match: "^",    Name: "power",   Priority: 90,  Assoc: Right },
}


var SyntaxDefinition = [...] string {

    /* Root */
    "program = module | command",

    /* Module */
    "module = module_declare exports commands",
    "module_declare = @module name!",
    "name = Name | String",
    "exports? = export exports",
    "export = @export as_list!",
    "as_list = as_item as_list_tail",
    "as_list_tail? = , as_item! as_list_tail",
    "as_item = name @as name! | name",

    /* Commands */
    "commands? = command commands",
    "command = cmd_group1 | cmd_group2 | cmd_group3",
    "cmd_group1 = cmd_flow | cmd_yield | cmd_await | cmd_return | cmd_err",
    "cmd_group2 = cmd_module | cmd_scope | cmd_def",
    "cmd_group3 = cmd_set | cmd_exec",
    /* Flow Control Commands @ Group 1 */
    "cmd_flow = cmd_if | cmd_switch | cmd_while | cmd_for",
    "cmd_if = @if expr! block! elifs else",
    "elifs? = elif elifs",
    "elif = @elif expr! block!",
    "else? = @else expr! block!",
    "cmd_switch = @switch expr { switch_branches! }!",
    "switch_branches? = switch_branch switch_branches",
    "switch_branch = @case expr! block!",
    "cmd_while = @while expr! loop_block!",
    "cmd_for = @for for_para! @in expr! loop_block!",
    "for_para = for_value | for_key , for_value",
    "for_key = name",
    "for_value = name",
    "block = { commands }!",
    "loop_block = { loop_cmds }!",
    "loop_cmds? = loop_cmd loop_cmds",
    "loop_cmd = loop_control | command",
    "loop_control = @break | @continue",
    /* Yield Command @ Group 1 */
    "cmd_yield = @yield name = expr! | @yield expr!",
    /* Await Command @ Group 1 */
    "cmd_await = @await name = expr! | @await expr!",
    /* Return Command @ Group 1 */
    "cmd_return = @return expr",
    /* Error Related Commands @ Group 1 */
    "cmd_err = cmd_throw | cmd_assert | cmd_require | cmd_try",
    "cmd_throw = @throw expr!",
    "cmd_assert = @assert expr!",
    "cmd_require = @require name! require_args { expr! }!",
    "require_args? = Call ( exprlist )",
    "cmd_try = @try name : command!",
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
    "cmd_reset = @reset name = expr",
    /* Definition Commands @ Group 2 */
    "cmd_def = function | abs_def",
    "abs_def = category | struct | enum | concept | class | interface",
    /* Set Command @ Group 3 */
    "cmd_set = @set left_val = expr",
    "left_val = operand_base gets!",
    "gets? = get gets",
    /* Exec Command @ Group 3 */
    "cmd_exec = @exec expr | expr",

    /* Function Definition */
    "function = proto {! body }!",
    "proto = fun_header name Call paralist_strict! ret",
    "fun_header = affect | fun_type | fun_type affect",
    "fun_type = @lazy | @async",
    "affect = @local | @upper | @global",
    "ret? = -> type",
    "body = static_head commands handle_tail",
    "static_head? = @static block",
    "handle_tail? = _at @handle name handle_block finally",
    "handle_block = { handle_cmds }!",
    "handle_cmds? = handle_cmd handle_cmds",
    "handle_cmd = unless | failed | command",
    "unless = @unless name unless_para { commands }",
    "unless_para? = Call ( namelist )",
    "failed = @failed name { commands }",
    "finally = _at @finally block",

    /* Abstraction Object Definition */
    /* Category */
    "category = @category name { branches! }!",
    "branches? = branch branches",
    "branch = abs_def | function",
    /* Schema */
    "struct = @struct name { field_list condition }",
    "field_list = field field_list_tail",
    "field_list_tail? = field! field_list_tail",
    "field = type name! field_default",
    "field_default? = = expr",
    "condition = @require expr",
    /* Enum */
    "enum = @enum name enum_literal",
    "enum_literal = { enum_items }",
    "enum_items = exprlist",
    /* Concept */
    "concept = @concept name expr",
    "concept_literal = { name _bar1 filters! }!",
    "filters = exprlist",
    /* Class */
    "class = @class name supers { initializers! methods }",
    "supers? = @is exprlist",
    "initializers? = initializer initializers",
    "initializer = @init paralist_strict! {! body! }!",
    "methods? = method methods",
    "method = method_proto {! body }!",
    "method_proto = method_type name paralist_strict! ret",
    "method_type? = &",
    /* Interface */
    "interface = @interface name { members }",
    "members? = member members",
    "member = method_proto | method",

    /* Expression */
    "expr = operand expr_tail",
    "expr_tail? = operator operand! expr_tail",
    /* Operators (Infix) */
    "operator = op_oriented | op_compare | op_logic | op_arith",
    "op_oriented = << | >> | => ",
    "op_compare = < | > | <= | >= | == | != ",
    `op_logic = @is | && | _bar2 | & | _bar1 | \ `,
    "op_arith = + | - | * | / | % | ^ ",
    /* Operators (Prefix) */
    "unary? = @not | - | _exc | ~ | @expose",
    /* Operand */
    "operand = unary operand_base operand_tail",
    "operand_base = ( expr! )! | lambda | literal | dot_para | identifier",
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
    "args = ( arglist )! extra_arg",
    "extra_arg? = -> lambda | -> adv_literal | = expr!",
    "arglist? = exprlist",
    "exprlist = expr exprlist_tail",
    "exprlist_tail? = , expr! exprlist_tail",

    /* Lambda */
    "lambda = lambda_full | lambda_simple | lambda_bool | lambda_nopl",
    "lambda_full = header_lambda paralist_lambda -> ret_lambda {! body! }!",
    "header_lambda = fun_type_lambda @lambda",
    "fun_type_lambda? = fun_type",
    "ret_lambda? = type",
    "lambda_simple = .{ paralist ->! expr! }! | .{ expr! }!",
    "lambda_bool = ..{ paralist ->! expr! }! | ..{ expr! }!",
    "lambda_nopl = ...{ body! }!",
    /* Parameter List */
    "paralist = ( ) | ( namelist ) | ( typed_namelist! )!",
    "paralist_lambda = name | Call paralist",
    "paralist_strict = ( ) | ( typed_namelist! )!",
    "typed_namelist = type policy name! typed_namelist_tail",
    "typed_namelist_tail? = , type! name! typed_namelist_tail",
    "policy? = & | *",
    /* Type Expression */
    "type = name | ( expr! )!",

    /* Literals */
    "literal = primitive | adv_literal",
    "adv_literal = comprehension | abs_literal | list | hash",
    "abs_literal = concept_literal | enum_literal",
    /* Hash Table */
    "hash = { } | { hash_item! hash_tail }!",
    "hash_tail? = , hash_item! hash_tail",
    "hash_item = name : expr | :: name",
    /* Linear List */
    "list = [ ] | [ list_item! list_tail ]!",
    "list_tail? = , list_item! list_tail",
    "list_item = expr list_item_extra",
    "list_item_extra? = : expr",
    /* List/Iterator Comprehension */
    "comprehension = .[ comp_rule! ]! | [ comp_rule ]!",
    "comp_rule = expr _bar1 in_list! opt_filters",
    "opt_filters? = exprlist",
    "in_list = in_item in_list_tail",
    "in_list_tail? = , in_item! in_list_tail",
    "in_item = name @in expr",
    /* Primitive Values */
    "primitive = string | number | bool",
    "string = String",
    "number = Hex | Exp | Dec | Int",
    "bool = @true | @false",

}