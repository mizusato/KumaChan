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
    Token { Name: "...",     Pattern: r(`\.\.\.`) }, // Delimiter
    Token { Name: "..",      Pattern: r(`\.\.`) },   // (Reserved)
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },   // Module Namespace
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: "$",       Pattern: r(`\$`) },     // Sequence
    Token { Name: "@",       Pattern: r(`@`) },      // Attached Member
    Token { Name: "??",      Pattern: r(`\?\?`) },   // Nil Coalescing
    Token { Name: "?",       Pattern: r(`\?`) },     // Optional Chaining
    Token { Name: ">=",      Pattern: r(`\>\=`) },
    Token { Name: "<=",      Pattern: r(`\<\=`) },
    Token { Name: "==",      Pattern: r(`\=\=`) },
    Token { Name: "!=",      Pattern: r(`\!\=`) },
    Token { Name: "=>",      Pattern: r(`\=\>`) },
    Token { Name: "=",       Pattern: r(`\=`) },
    Token { Name: "->",      Pattern: r(`\-\>`) },
    Token { Name: "<-",      Pattern: r(`\<\-`) },   // Pull from Iterator
    Token { Name: "<<",      Pattern: r(`\<\<`) },   // Bitwise SHL
    Token { Name: ">>",      Pattern: r(`\>\>`) },   // Bitwise SHR
    Token { Name: "<",       Pattern: r(`\<`) },
    Token { Name: ">",       Pattern: r(`\>`) },
    Token { Name: "!!",      Pattern: r(`\!\!`) },   // Bitwise NOT
    Token { Name: "!",       Pattern: r(`\!`) },     // Type Assertion
    Token { Name: "~",       Pattern: r(`~`) },      // Continuation
    Token { Name: "&&",      Pattern: r(`\&\&`) },   // Bitwise AND
    Token { Name: "&",       Pattern: r(`\&`) },     // (Reserved)
    Token { Name: "||",      Pattern: r(`\|\|`) },   // Bitwise OR
    Token { Name: "|",       Pattern: r(`\|`) },     // Pipeline
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
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+LF+`]+`) },
    //    { Name: "NoLF",    [ Inserted by Scanner ] },
    //    { Name: "Void",    [ Inserted by Scanner ] },
}

var ExtraTokens = [...] string { "NoLF", "Void" }

var ConditionalKeywords = [...] string {

    "@export", "@resolve", "@of", "@import", "@as",

    "@section",
    "@singleton",
    "@schema", "@static",
    "@class", "@is", "@init", "@private", "@native",
	"@trait", "@bound", "@interface", "@union",
    "@type",
    "@function",

    "@if", "@else",
    "@while", "@for", "@in", "@break", "@continue",
    "@yield",
    "@return",
    "@assert", "@panic", "@finally",
    "@let", "@initial", "@reset",
    "@do", "@nothing",

    "@not", "@and", "@or", "@super", "@try",
    "@callcc",
    "@new", "@struct", "@tuple",
    "@when", "@match", "@otherwise",
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
    "module = header export imports decls commands",
      "header = shebang resolve",
        "shebang? = Pragma",
        "resolve? = @resolve { resolve_item more_resolve_items }!",
		  "more_resolve_items? = resolve_item more_resolve_items",
          "resolve_item = name =! version string!",
            "version? = name @of!",
      "export? = @export { namelist! }! | @export namelist!",
        "namelist = name namelist_tail",
          "namelist_tail? = , name! namelist_tail",
          "name = Name",
      "imports? = import imports",
        "import = import_from alias | import_from import_names",
          "import_from = @import name ::!",
          "import_names = * | {! alias_list! }!",
            "alias_list = alias alias_list_tail",
              "alias_list_tail? = , alias! alias_list_tail",
              "alias = name @as alias_name | name",
                "alias_name = name!",
      // decls -> Group: Declaration
      "commands? = command commands",
        // command -> Group: Command
    /* Group: Type & Generics */
    "type = type_ordinary | type_trait | type_misc",
      "type_ordinary = module_prefix name type_args",
        "module_prefix? = name :: ",
        "type_args? = NoLF [ typelist! ]!",
          "typelist = type typelist_tail",
            "typelist_tail? = , type! typelist_tail",
      "type_trait = * trait!",
      "type_misc = tuple_t | function_t | iterator_t | continuation_t",
        "tuple_t = [ [ typelist! ]! ]!",
        "function_t = [ signature more_signature ]!",
          "more_signature? = _bar1 signature! more_signature",
          "signature = -> type! | typelist ->! type!",
        "iterator_t = $ [! type! ]!",
        "continuation_t = ~ [! type! ]!",
    "type_params? = [ type_param! more_type_param ]!",
      "more_type_param? = , type_param! more_type_param",
        "type_param = name : trait! | name",
          "trait = module_prefix name type_args | constraint",
    /* Group: Declaration */
    "decls? = decl decls",
      "decl = section | function | decl_type | decl_trait",
        "section = @section name { decls }!",
        "function = f_overload | f_single",
          "f_single = @function name type_params paralist! ret body!",
            "paralist = ( ) | ( typed_list! )!",
              "typed_list = typed_list_item typed_list_tail",
              "typed_list_tail? = , typed_list_item! typed_list_tail",
              "typed_list_item = name :! type!",
                // type -> Group: Type & Generics
            "ret = ->! type!",
            "body = { commands mock }!",
              "mock? = ... Pragma! {! commands }!",
          "f_overload = @function name type_params { f_item! more_f_items }",
              "more_f_items? = f_item more_f_items",
              "f_item = paralist ret body!",
        "decl_type = singleton | schema | class | type_alias",
          "singleton = @singleton namelist",
          "schema = attrs @struct name type_params is {! fields static }!",
            "attrs? = Pragma",
            "is? = @is base_list",
			  "base_list = ( typelist! )! | typelist!",
            "fields? = field fields",
              "field = name : type! field_default",
                "field_default? = = expr",
            "static? = @static {! class_body }!",
          "class = attrs @class name type_params is {! class_body static }!",
            "class_body = init methods",
              "init? = @init paralist! body!",
              "methods? = method methods",
                "method = opt_private name paralist! ret method_body!",
                  "opt_private? = @private",
                  "method_body = { @native } | body!",
          "type_alias = @type name type_params =! type",
        "decl_trait = @trait name type_params trait_arg trait_body",
		  "trait_arg? = ( name! )!",
          "trait_body = {! constraint! more_constraints }!",
            "more_constraints? = constraint more_constraints",
            "constraint = bound | interface | i_static | union | trait",
              "bound = @bound {! bound_op! type! }!",
                "bound_op = < | <= | > | >= ",
              "interface = @interface {! protos }!",
                "protos? = proto protos",
                "proto = name paralist! ret!",
              "i_static = @static {! protos }!",
              "union = @union {! typelist! }!",
    /* Group: Command */
    "command = LF cmd_group1 | LF cmd_group2 | LF cmd_group3",
      "cmd_group1 = cmd_cond | cmd_loop | cmd_loop_ctrl",
        "cmd_cond = cmd_if",
          "cmd_if = @if wrapped! block! elifs else",
            "block = { imports commands }!",
            "elifs? = elif elifs",
              "elif = @else @if wrapped! block!",
            "else? = @else block!",
        "cmd_loop = cmd_while | cmd_for",
          "cmd_while = @while wrapped! block!",
          "cmd_for = @for for_params! @in wrapped! block!",
            "for_params = { namelist! }! | name",
        "cmd_loop_ctrl = @break | @continue",
      "cmd_group2 = cmd_return | cmd_yield | cmd_panic | cmd_guard",
        "cmd_return = return_flags @return return_content",
          "return_flags? = Pragma",
          "return_content = Void | expr",
        "cmd_yield = @yield expr!",
        "cmd_panic = @panic expr!",
        "cmd_guard = cmd_assert | cmd_finally",
          "cmd_assert = @assert expr!",
          "cmd_finally = @finally { commands }!",  // block-level defer
      "cmd_group3 = cmd_scope | cmd_pass | cmd_side_effect",
        "cmd_scope = cmd_let | cmd_initial | cmd_reset",
          "cmd_let = opt_static @let let_pattern = expr!",
			"let_pattern = name | { namelist }!",
			"opt_static? = @static",
          "cmd_initial = opt_static @initial name = expr!",
          "cmd_reset = @reset name reset_operator = expr",
            "reset_operator? = op_arith",
          // operand_body -> Group: Operand
        "cmd_pass = @do @nothing",
        "cmd_side_effect = expr",
    /* Group: Expression */
    "expr = operand expr_tail",
      "expr_tail? = operator operand! expr_tail",
        // operand -> Group: Operand
        "operator = op_nil | op_compare | op_bitwise | op_logic | op_arith",
          "op_nil = ?? ",
          "op_compare = < | > | <= | >= | == | != ",
          "op_bitwise = << | >> | && | ^^ | _bar2 ",
          "op_logic = @and | @or ",
          "op_arith = + | - | * | / | % | ^ ",
    /* Group: Operand */
    "operand = unary operand_body accesses calls with pipelines",
      "unary? = @not | _exc2 | + | - | <- | @try | @super",
      "operand_body = lambda | wrapped | cast | callcc | misc | variable",
        "lambda = iterator | paralist_weak ret_weak body_flex",
          "iterator = $ yield_type => body!",
            "yield_type? = : type!",
          "paralist_weak? = name | ( ) | ( weak_param more_weak_params )",
            "more_weak_params? = , weak_param! more_weak_params",
            "weak_param = name : type! | name",
          "ret_weak? = : type",
          "body_flex = => body | => expr!",
        "wrapped = ( expr! )!",
        "cast = cast_flag [ cast_type ]! cast_expr",
          "cast_flag? = _exc1",
          "cast_type = narrow_type | type!",
            "narrow_type = - type",
          "cast_expr = { expr! }! | ( expr! )! | expr!",
        "callcc = @callcc NoLF [! type! ]! :! expr!",
        "misc = type_object | type_related | literal | seq | text | guard",
          "type_object = @type { type! }!",
          "type_related = new | attached",
		    "new = @new type!",
		    "attached = attached_name",
              "attached_name = _at type! :! name!",
		  "literal = struct | tuple | seq | text | const",
		    "struct = @struct type! {! struct_items }!",
		      "struct_items? = struct_item struct_items",
		        "struct_item = name =! expr!",
		    "tuple = @tuple type_args { exprlist! }!",
              "exprlist = expr exprlist_tail",
                "exprlist_tail? = , expr! exprlist_tail",
                // expr -> Group: Expression
            "const = string | int | float | bool",
              "string = String",
              "int = Dec | Hex | Oct | Bin",
              "float = Float | Exp",
              "bool = @Yes | @No",
          "seq = $ yield_type { seq_items }",
            "seq_items? = exprlist",
          "text = Text | TxBegin first_segment more_segments TxEnd!",
            "first_segment = segment_tag expr!",
              "segment_tag? = name : ",
            "more_segments? = next_segment more_segments",
              "next_segment = TxInner segment_tag expr!",
          "guard = when | match",
	        "when = @when {! case! more_cases }!",
	          "more_cases? = case more_cases",
              "case = when_key =>! block_or_expr",
                "when_key = @otherwise | wrapped!",
                "block_or_expr = block | expr!",
            "match = @match wrapped! {! type_case! more_type_cases }!",
              "more_type_cases? = type_case more_type_cases",
              "type_case = match_key match_pattern =>! block_or_expr",
                "match_key = @otherwise | type!",
                "match_pattern? = ( name )! | { namelist! }!",
        "variable = module_prefix name type_args",
      "accesses? = access accesses",
        "access = . name! method_call chain_opt",
          "method_call? = call | = expr!",
			"call = NoLF ( arglist )! | { arglist }!",
			  "arglist? = exprlist",
		  "chain_opt? = ?",
      "calls? = call calls",
      "with? = @with {! temps }!",
        "temps? = temp temps",
          "temp = name =! expr!",
      "pipelines? = pipeline pipelines",
        "pipeline = _bar1 operand_body pipeline_args",
        "pipeline_args? = NoLF ( arglist )! | { arglist }!",
}
