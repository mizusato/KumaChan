package syntax

import "regexp"
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `;\{\}\[\]\(\)\.\,\:\<\>\=\!~\&\|\\\+\-\*\/%`

var EscapeMap = map[string]string {
    "_exc":   "!",
    "_bar1":  "|",
    "_bar2":  "||",
}

var Extra = [...]string { "Call", "Get" }

var Tokens = [...]Token {
    Token { Name: "String",  Pattern: r(`'[^']*'`), },
    Token { Name: "String",  Pattern: r(`"[^"]*"`), },
    Token { Name: "Raw",     Pattern: r(`/~([^~]|[^/]|~[^/]|[^~]/)*~/`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^\n]*`) },
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


var Keywords = [...]string {
    "@module", "@export", "@as", "@use", "@import",
    "@let", "@var", "@reset", "@unset", "@set",
    "@local", "@upper", "@global", "@static",
    "@category", "@struct", "@require", "@enum", "@concept",
    "@class", "@init", "@interface",
    "@true", "@false",
}


var SyntaxDefinition = [...]string {

    "program = module | expr",

    "module = module_declare exports commands",
    "module_declare = @module name!",
    "name = Name | String",
    "exports? = export exports",
    "export = @export as_list!",
    "as_list = as_item as_list_tail",
    "as_list_tail? = , as_item! as_list_tail",
    "as_item = name @as name! | name",

    "commands? = command commands",
    "command = cmd_module cmd_scope cmd_set cmd_def cmd_expr",
    "cmd_expr = expr",

    "cmd_module = cmd_use | cmd_import",
    "cmd_use = @use as_list",
    "cmd_import = @import name_list",
    "name_list = name name_list_tail",
    "name_list_tail? = , name! name_list_tail",

    "cmd_scope = cmd_let | cmd_var | cmd_reset | cmd_unset",
    "cmd_let = @let name = expr",
    "cmd_var = @var name = expr",
    "cmd_reset = @reset name = expr",
    "cmd_unset = @unset name",

    "cmd_set = @set left_val = expr",
    "left_val = operand_base get_list",
    "get_list = get get_list_tail",
    "get_list_tail? = get get_list_tail",

    "cmd_def = function | abs_def",
    "abs_def = category | struct | enum | concept | class | interface",

    "function = proto {! body }!",
    "proto = affect name paralist_strict! ret",
    "affect = @local | @upper | @global",
    "ret? = -> type",
    "body = static_block commands",
    "static_block? = @static { commands }",

    "category = @category name { branches! }!",
    "branches? = branch branches",
    "branch = abs_def",

    "struct = @struct name { field_list condition }",
    "field_list = field field_list_tail",
    "field_list_tail? = field! field_list_tail",
    "field = type name! field_default",
    "field_default? = = expr",
    "condition = @require expr",

    "enum = @enum name = enum_literal",

    "concept = @concept name = expr",

    "class = @class name { initializers! methods }",
    "initializers? = initializer initializers",
    "initializer = @init paralist_strict! {! body! }!",
    "methods? = method methods",
    "method = method_proto {! body }!",
    "method_proto = method_type name paralist_strict! ret",
    "method_type? = &",

    "interface = @interface name { members }",
    "members? = member members",
    "member = method_proto | method",

    "expr = operand expr_tail",
    "expr_tail? = operator operand! expr_tail",

    "operator = op_group1 | op_group2 | op_group3 | op_group4",
    "op_group1 = >= | <= | != | == | => | = ",
    "op_group2 = -> | << | >> | > | < ",
    `op_group3 = _exc | && | _bar2 | ~ | & | _bar1 | \ `,
    "op_group4 = + | - | * | / | % | ^ ",

    "operand = operand_base operand_tail",
    "operand_base = ( expr! )! | lambda | literal | dot_para | identifier",
    "operand_tail? = get operand_tail | call operand_tail",
    "get = get_expr | get_name",
    "call = call_self | call_method",
    "get_expr = Get [ expr! ]!",
    "get_name = Get . name!",
    "call_self = Call arglist!",
    "call_method = -> name Call args! | -> name extra_arg!",
    "args = ( arglist )! extra_arg",
    "extra_arg? = -> lambda",
    "arglist? = exprlist",
    "exprlist = expr exprlist_tail",
    "exprlist_tail? = , expr! exprlist_tail",

    "lambda = lambda_full | lambda_simple | lambda_bool",
    "lambda_full = paralist ->! ret_lambda {! body! }!",
    "lambda_simple = .{ paralist ->! expr! }! | .{ expr! }!",
    "lambda_bool = ..{ paralist ->! expr! }! | ..{ expr! }!",
    "ret_lambda? = type",

    "paralist = name | ( ) | ( name_list! )! | ( typed_list! )!",
    "paralist_strict = ( ) | ( typed_list! )!",
    "typed_list = type policy name! typed_list_tail",
    "typed_list_tail? = , type! name! typed_list_tail",

    "type = type_base type_ext",
    "type_ext? = < type_args! >!",
    "type_base = name type_base_tail",
    "type_base_tail? = . name! type_base_tail",
    "type_args = type type_args_tail",
    "type_args_tail? = , type! type_args_tail",
    "policy? = & | *",

    "identifier = Name",
    "dot_para = . Name",
    "literal = hash | list | concept_literal | enum_literal | primitive",

    "hash = { } | { hash_item! hash_tail }!",
    "hash_tail? = , hash_item! hash_tail",
    "hash_item = name : expr",

    "list = [ ] | [ list_item! list_tail ]!",
    "list_tail? = , list_item! list_tail",
    "list_item = expr",

    "concept_literal = { name _bar1 filters! }!",
    "filters = exprlist",

    "enum_literal = { enum_items }!",
    "enum_items = exprlist",

    "primitive = string | number | bool",
    "string = String",
    "number = Hex | Exp | Dec | Int",
    "bool = @true | @false",

}
