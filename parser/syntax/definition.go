package syntax

import "regexp"
type Regexp = *regexp.Regexp
func r (pattern string) Regexp { return regexp.MustCompile(`^` + pattern) }

const LF = `\n`
const Blanks = ` \t\r　`
const Symbols = `;\{\}\[\]\(\)\.\,\:#@\?\<\>\=\!~\&\|\\\+\-\*\/%\^'"` + "`"

var EscapeMap = map [string] string {
    "_exc1":  "!",
    "_exc2":  "!!",
    "_bar1":  "|",
    "_bar2":  "||",
    "_at":    "@",
}


var Extra = [...] string { "Call", "Get", "Void" }

var Tokens = [...] Token {
    Token { Name: "String",  Pattern: r(`'[^']*'`) },
    Token { Name: "Text",    Pattern: r(`"[^"@\[\]]*"`) },
    Token { Name: "TxBegin", Pattern: r(`"[^"@\[\]]*@\[`) },
    Token { Name: "TxInner", Pattern: r(`\][^"@\[\]]*@\[`) },
    Token { Name: "TxEnd",   Pattern: r(`\][^"@\[\]]*"`) },
    Token { Name: "Comment", Pattern: r(`/\*([^\*]|[^/]|\*[^/]|[^\*]/)*\*/`) },
    Token { Name: "Comment", Pattern: r(`//[^\n]*`) },
    Token { Name: "Pragma",  Pattern: r(`#[^\n]*`) },
    Token { Name: "Blank",   Pattern: r(`[`+Blanks+`]+`) },
    Token { Name: "LF",      Pattern: r(LF+`+|;+`) },
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
    Token { Name: "...",     Pattern: r(`\.\.\.`) }, // Delimiter
    Token { Name: "..",      Pattern: r(`\.\.`) },   // Sequence
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },   // Module Namespace
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: "@",       Pattern: r(`@`) },      // Class/Schema Attribute
    Token { Name: "??",      Pattern: r(`\?\?`) },   // Nil Coalescing
    Token { Name: "?",       Pattern: r(`\?`) },     // Flag: optional
    Token { Name: ">=",      Pattern: r(`\>\=`) },
    Token { Name: "<=",      Pattern: r(`\<\=`) },
    Token { Name: "==",      Pattern: r(`\=\=`) },
    Token { Name: "!=",      Pattern: r(`\!\=`) },
    Token { Name: "~~",      Pattern: r(`\~\~`) },   // Shallow Equal
    Token { Name: "!~",      Pattern: r(`\!\~`) },   // Shallow Unequal
    Token { Name: "=>",      Pattern: r(`\=\>`) },
    Token { Name: "=",       Pattern: r(`\=`) },
    Token { Name: "->",      Pattern: r(`\-\>`) },
    Token { Name: "<-",      Pattern: r(`\<\-`) },   // Element of (a Type)
    Token { Name: "<<",      Pattern: r(`\<\<`) },   // Bitwise SHL
    Token { Name: ">>",      Pattern: r(`\>\>`) },   // Bitwise SHR
    Token { Name: "<",       Pattern: r(`\<`) },
    Token { Name: ">",       Pattern: r(`\>`) },
    Token { Name: "!!",      Pattern: r(`\!\!`) },   // Bitwise NOT
    Token { Name: "!",       Pattern: r(`\!`) },     // Flag: force
    Token { Name: "~",       Pattern: r(`~`) },      // Continuation
    Token { Name: "&&",      Pattern: r(`\&\&`) },   // Bitwise AND
    Token { Name: "&",       Pattern: r(`\&`) },     // Type Intersection
    Token { Name: "||",      Pattern: r(`\|\|`) },   // Bitwise OR
    Token { Name: "|",       Pattern: r(`\|`) },     // Type Union
    Token { Name: `\`,       Pattern: r(`\\`) },     // (Reserved)
    Token { Name: "++",      Pattern: r(`\+\+`) },   // (Reserved)
    Token { Name: "+",       Pattern: r(`\+`) },
    Token { Name: "--",      Pattern: r(`\--`) },    // (Reserved)
    Token { Name: "-",       Pattern: r(`\-`) },
    Token { Name: "**",      Pattern: r(`\*\*`) },   // (Reserved)
    Token { Name: "*",       Pattern: r(`\*`) },
    Token { Name: "/",       Pattern: r(`\/`) },
    Token { Name: "%%",      Pattern: r(`%%`) },     // (Reserved)
    Token { Name: "%",       Pattern: r(`%`) },
    Token { Name: "^^",      Pattern: r(`\^\^`) },   // Bitwise XOR
    Token { Name: "^",       Pattern: r(`\^`) },
    Token { Name: "Callcc",  Pattern: r(`call\/cc`) },
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+LF+`]+`) },
    //    { Name: "Call",    [ Inserted by Scanner ] },
    //    { Name: "Get",     [ Inserted by Scanner ] },
    //    { Name: "Void",    [ Inserted by Scanner ] },
}


/* Conditional Keywords */
var Keywords = [...] string {

    "@resolve", "@export", "@from", "@import", "@as",

    "@section",
    "@singleton", "@union", "@trait",
    "@schema",
    "@class", "@is", "@implements", "@init", "@private",
    "@interface",
    "@native",

    "@function", "@static", "@mock",
    "@handle", "@unless", "@failed", "@finally",

    "@if", "@else", "@switch", "@otherwise",
    "@while", "@for", "@in", "@break", "@continue",
    "@yield",
    "@return", "@tailing",
    "@assert", "@panic", "@ensure", "@unable", "@to",
    "@let",  "@initial", "@reset",
    "@set",
    "@do", "@nothing",

    "@not", "@and", "@or", "@mount",  // no await anymore. use call/cc instead
    "@type", "@new", "@struct",
    "@when", "@match", "@invoke",
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
    Operator { Match: "<-",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "<=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: ">=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "==",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!=",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "~~",   Priority: 50,  Assoc: Left,   Lazy: false  },
    Operator { Match: "!~",   Priority: 50,  Assoc: Left,   Lazy: false  },
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


var SugarableOperators = [] string {
    "==", "<",
    "negate", "+", "-", "*", "/", "%", "^",
}


var SyntaxDefinition = [...] string {
    /* Group: Root */
    "eval = resolve imports decls commands handle_hook",
    "module_header = shebang export resolve",
    "module = shebang export resolve imports decls commands handle_hook",
      "shebang? = Pragma",
      "export? = @export { namelist! }! | @export namelist!",
        "namelist = name namelist_tail",
        "namelist_tail? = , name! namelist_tail",
        "name = Name",
      "resolve? = @resolve { resolve_item more_resolve_items }!",
		"more_resolve_items? = resolve_item more_resolve_items",
        "resolve_item = name =! resolve_alias String",
          "resolve_alias? = name resolve_version @in!",
            "resolve_version? = ( name! )!",
      "imports? = import imports",
        "import = @import imported_module imported_names",
          "imported_module = alias",
            "alias = name @as name! | name",
          "imported_names? = { * }! | { alias_list! }!",
            "alias_list = alias alias_list_tail",
            "alias_list_tail? = , alias! alias_list_tail",
      // decls -> Group: Declaration
      "commands? = command commands",
        // command -> Group: Command
    /* Group: Type & Generics */
    "type = type_ordinary | type_function | type_continuation",
      "type_ordinary = module_prefix name type_args",
        "module_prefix? = name :: ",
        "type_args? = [ typelist! ]!",
            "typelist = type typelist_tail",
            "typelist_tail? = , type! typelist_tail",
      "type_function = type_generator | [ signature more_signature ]!",
        "more_signature? = _bar1 signature! more_signature",
        "signature = -> type! | typelist ->! type!",
        "type_generator = * [! type! ]!",
      "type_continuation = ~ [! type! ]!",
    "type_params? = [ type_param! more_type_param ]!",
      "more_type_param? = , type_param! more_type_param",
        "type_param = name bounds | name",
          "bounds = single_bound | : [ lower_bound! ,! upper_bound! ]!",
            "single_bound = < type! | > type!",
            "lower_bound = type",
            "upper_bound = type",
    /* Group: Declaration */
    "decls? = decl decls",
      "decl = decl_section | decl_function | decl_type | decl_native_type",
        "decl_section = @section name { decls }!",
        "decl_function = opt_pragma f_overload | opt_pragma f_single",
          "opt_pragma? = Pragma",
          "f_single = @function name type_params paralist! ret body!",
            "paralist = ( ) | ( typed_list! )!",
              "typed_list = typed_list_item typed_list_tail",
              "typed_list_tail? = , typed_list_item! typed_list_tail",
              "typed_list_item = name :! type!",
                // type -> Group: Type & Generics
            "ret = ->! type!",
            "body = { static_commands commands mock_hook handle_hook }!",
              "static_commands? = @static { commands }",
              "mock_hook? = ... @mock name! { commands }",
              "handle_hook? = ... handle_cmds",
                "handle_cmds? = handle_cmd handle_cmds",
                  "handle_cmd = unless | failed",
                    "unless = @unless name! handle_params reaction",
                      "handle_params? = ( typed_list! )!",
                      "reaction = {! commands }!",
                    "failed = @failed opt_to name! handle_params reaction",
                      "opt_to? = @to",
          "f_overload = @function name type_params { f_item_list }",
            "f_item_list = f_item f_item_list_tail",
              "f_item_list_tail? = , f_item! f_item_list_tail",
              "f_item = paralist! ->! type! body!",
        "decl_type = singleton | union | trait | schema | class | interface",
          "singleton = @singleton name! | @singleton {! namelist! }!",
          "union = @union name type_params = type union_tail",
            "union_tail? = _bar1 type! union_tail",
          "trait = @trait name type_params = type trait_tail",
            "trait_tail? = & type! trait_tail",
          "schema = schema_head name type_params bases {! field_list! }!",
            "schema_head = attributes @struct",
              "attributes? = attribute attributes",
                "attribute = _at name!",
            "bases? = < typelist!",
            "field_list = field field_list_tail",
              "field_list_tail? = , field! field_list_tail",
              "field = name : type! field_default",
                "field_default? = = expr",
          "class = class_head name type_params is impls { init pfs methods }",
            "class_head = attributes @class",
            "is? = @is base_list",
              "base_list = ( typelist! )! | typelist!",
            "impls? = @implements base_list",
            "init = @init paralist! body!",
            "pfs? = pf pfs",
              "pf = @private name paralist! ret body!",
            "methods? = method methods",
              "method = name paralist! ret body!",
          "interface = @interface name type_params { method_proto_list! }!",
            "method_proto_list = method_proto more_method_protos",
              "more_method_protos? = method_proto more_method_protos",
              "method_proto = name paralist! ret",
        "decl_native_type = decl_native_class",
          "decl_native_class = @native name type_params { native_methods }!",
            "native_methods? = native_method native_methods",
              "native_method = method_proto",
    /* Group: Command */
    "command = cmd_group1 | cmd_group2 | cmd_group3",
      "cmd_group1 = cmd_branch | cmd_loop | cmd_loop_ctrl",
        "cmd_branch = cmd_if | cmd_switch",
          "cmd_if = @if expr! block! elifs else",
            "block = { commands }!",
            "elifs? = elif elifs",
              "elif = @else @if expr! block!",
            "else? = @else block!",
          "cmd_switch = switch_when | switch_match",
            "switch_when = @switch @when {! cases! }!",
              "cases = case more_cases",
                "more_cases? = case more_cases",
                "case = @otherwise :! block! | expr :! block!",
            "switch_match = @switch @match expr {! type_cases! }!",
              "type_cases = type_case more_type_cases",
                "more_type_cases? = type_case more_type_cases",
                "type_case = @otherwise =>! block! | type =>! block!",
        "cmd_loop = cmd_while | cmd_for",
          "cmd_while = @while expr! block!",
          "cmd_for = @for for_params! @in expr! block!",
            "for_params = value_with_key | value_with_index | value",
              "value_with_index = name [ for_index! ]!",
                "for_index = name",
              "value_with_key = { for_key :! name! }!",
                "for_key = name",
              "value = name",
        "cmd_loop_ctrl = @break | @continue",
      "cmd_group2 = cmd_return | cmd_yield | cmd_panic | cmd_guard",
        "cmd_return = return_flags @return return_content",
          "return_flags? = _at @tailing",
          "return_content = Void | expr",
        "cmd_yield = @yield expr!",
          "pattern = name | { namelist }! | [ namelist ]!",
        "cmd_panic = @panic expr!",
        "cmd_guard = cmd_assert | cmd_ensure | cmd_unable | cmd_finally",
          "cmd_assert = @assert expr!",
          "cmd_ensure = @ensure name! ensure_args { expr! }!",
            "ensure_args? = Call ( exprlist )!",
              "exprlist = expr exprlist_tail",
                "exprlist_tail? = , expr! exprlist_tail",
                // expr -> Group: Expression
          "cmd_unable = @unable opt_to name! unable_args",
            "unable_args? = Call ( exprlist )!",
          "cmd_finally = @finally { commands }!",  // block-level defer
      "cmd_group3 = cmd_scope | cmd_set | cmd_pass | cmd_side_effect",
        "cmd_scope = cmd_let | cmd_initial | cmd_reset",
          "cmd_let = @let pattern = expr!",
          "cmd_initial = @initial name = expr!",
          "cmd_reset = @reset name set_op = expr",
            "set_op? = op_arith",
        "cmd_set = @set left_val set_op = expr",
          "left_val = operand_body",
          // operand_body -> Group: Operand
        "cmd_pass = @do @nothing",
        "cmd_side_effect = expr",
    /* Group: Expression */
    "expr = operand expr_tail",
      "expr_tail? = operator operand! expr_tail",
        // operand -> Group: Operand
        "operator = op_nil | op_compare | op_bitwise | op_logic | op_arith",
          "op_nil = ?? ",
          "op_compare = < | > | <- | <= | >= | == | != | ~~ | !~ ",
          "op_bitwise = << | >> | && | ^^ | _bar2 ",
          "op_logic = @and | @or ",
          "op_arith = + | - | * | / | % | ^ ",
    /* Group: Operand */
    "operand = unary operand_body plain_calls with pipelines",
      "unary? = @not | _exc2 | - | @mount",
      "operand_body = operand_base operand_tail",
        "operand_base = lambda | wrapped | const | cast | callcc | misc | var",
          "lambda = generator | paralist_weak ret_weak body_flex",
            "generator = * yield_type => body!",
              "yield_type? = : type!",
            "paralist_weak? = name | ( ) | ( weak_param more_weak_params )!",
              "more_weak_params? = , weak_param! more_weak_params",
              "weak_param = name : type! | name",
            "ret_weak? = : type",
            "body_flex = => body | => expr!",
          "wrapped = ( expr! )!",
          "const = string | int | float | bool",
            "string = String",
            "int = Dec | Hex | Oct | Bin",
            "float = Float | Exp",
            "bool = @Yes | @No",
          "cast = [ type ]! cast_flag (! expr! )!",
            "cast_flag? = _exc1",
          "callcc = Callcc [! type! ]! (! expr! )!",
          "misc = type_expr | text | new | struct | seq | when | match | iife",
            "type_expr = @type { type }",
            "text = Text | TxBegin first_segment more_segments TxEnd!",
              "first_segment = segment_tag expr!",
                "segment_tag? = name : ",
              "more_segments? = next_segment more_segments",
                "next_segment = TxInner segment_tag expr!",
			"new = @new type!",
			"struct = @struct type! {! struct_items }!",
			  "struct_items? = struct_item struct_items",
			    "struct_item = name :! expr!",
			"seq = brace_seq | dot_seq",
			  "brace_seq = * yield_type { seq_items }!",
                "seq_items? = exprlist",
			  "dot_seq = .. dot_seq_type seq_items ..!",
                "dot_seq_type? = type_generator :",
	    	"when = @when {! branch_list }!",
              "branch_list = branch branch_list_tail",
              "branch_list_tail? = , branch branch_list_tail",
              "branch = @otherwise :! expr! | expr! :! expr!",
            "match = @match expr {! type_branch_list }!",
              "type_branch_list = type_branch type_branch_list_tail",
              "type_branch_list_tail? = , type_branch type_branch_list_tail",
              "type_branch = @otherwise =>! expr! | type! =>! expr!",
            "iife = @invoke body",
          "var = module_prefix name",
        "operand_tail? = tail_operation operand_tail",
          "tail_operation = field_get | slice | get | inflate",
            "field_get = Get . name!",
            "slice = Get [ low_index : high_index ]!",
              "low_index? = expr",
              "high_index? = expr",
            "get = Get [ exprlist! ]! get_flag",
              "get_flag? = ?",
            "inflate = :: [ typelist! ]!",
      "plain_calls? = plain_call plain_calls",
        "plain_call = Call ( arglist )! | { arglist }!",
          "arglist? = exprlist",
      "with? = @with struct",
      "pipelines? = pipeline pipelines",
        "pipeline = _bar1 operand_body pipeline_args",
        "pipeline_args? = Call ( arglist )! | { arglist }!",
}
