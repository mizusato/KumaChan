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
    Token { Name: "...",     Pattern: r(`\.\.\.`) }, // (Reserved)
    Token { Name: "..",      Pattern: r(`\.\.`) },   // (Reserved)
    Token { Name: ".",       Pattern: r(`\.`) },
    Token { Name: ",",       Pattern: r(`\,`) },
    Token { Name: "::",      Pattern: r(`\:\:`) },   // Module Namespace
    Token { Name: ":",       Pattern: r(`\:`) },
    Token { Name: "$",       Pattern: r(`\$`) },     // Iterator Literal
    Token { Name: "$$",      Pattern: r(`\$$`) },    // (Reserved)
    Token { Name: "@",       Pattern: r(`@`) },
    Token { Name: "@@",      Pattern: r(`@@`) },     // (Reserved)
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
    Token { Name: "&",       Pattern: r(`\&`) },     // Procedure Flag
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
    Token { Name: "^",       Pattern: r(`\^`) },     // Tuple
    Token { Name: "Name",    Pattern: r(`[^`+Symbols+Blanks+LF+`]+`) },
    //    { Name: "NoLF",    [ Inserted by Scanner ] },
    //    { Name: "Void",    [ Inserted by Scanner ] },
}

var ExtraTokens = [...] string { "NoLF", "Void" }

var ConditionalKeywords = [...] string {

    "@external", "@import", "@as",

    "@section",
    "@procedure", "@function",
    "@singleton",
    "@struct", "@class", "@mutable",
    "@extends", "@static", "@init", "@data", "@private", "@native",
    "@type",
    "@trait", "@is", "@union",
    "@const",

    "@if", "@else",
    "@while", "@for", "@in", "@break", "@continue",
    "@yield",
    "@return",
    "@assert", "@panic", "@finally",
    "@let", "@initial", "@reset",
    "@do", "@nothing",

    "@not", "@and", "@or", "@try",
    "@callcc",
    "@new",
    "@when", "@switch", "@otherwise",
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
    "module = header imports decls commands",
      "header = shebang externals",
        "shebang? = Pragma",
        "externals? = external externals",
		  "external = @external name version =! string!",
            "version? = major .! minor!",
              "major = Dec",
              "minor = Dec",
      "imports? = import imports",
        "import = import_from alias | import_from {! alias_list! }!",
          "import_from = @import name ::!",
          "alias_list = alias alias_list_tail",
            "alias_list_tail? = , alias! alias_list_tail",
            "alias = name @as alias_name | name",
              "alias_name = name!",
      "decls? = decl decls",
        // decl -> Group: Declaration
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
    "decl = section | function | decl_type | decl_trait | decl_const",
      "section = @section name { decls }!",
      "function = attrs func_kind name type_params paralist! ret body!",
        "func_kind = @function | @procedure",
        "attrs? = Pragma",
        "paralist = ( ) | ( typed_list! )!",
          "typed_list = typed_list_item typed_list_tail",
          "typed_list_tail? = , typed_list_item! typed_list_tail",
          "typed_list_item = name :! type!",
            // type -> Group: Type & Generics
        "ret = ->! type!",
        "body = { @native } | block!",
          "block = { imports commands }!",
      "decl_type = singleton | struct | class | type_alias",
        "singleton = attrs @singleton namelist",
          "namelist = name namelist_tail",
            "namelist_tail? = , name! namelist_tail",
            "name = Name",
        "struct = attrs @struct name type_params {! struct_body }!",
          "struct_body = extends fields static",
            "extends? = _at @extends! trait!",
            "fields? = field fields",
              "field = LF name :! type! field_default",
                "field_default? = ( const_expr! )!",
                  // const_expr -> Group: Expression
            "static? = _at @static {! methods }!",
              "methods? = method methods",
                "method = prereq opt_private name paralist! ret body!",
                  "prereq = type_params",
                  "opt_private? = @private",
        "class = attrs class_kind name type_params {! class_body }!",
          "class_kind = @class | @mutable",
          "class_body = init methods static",
            "init? = _at @init paralist! init_data body!",
              "init_data? = : _at! @data! (! namelist! )!",
        "type_alias = @type name type_params =! type",
      "decl_trait = @trait name type_params trait_arg trait_body",
        "trait_arg? = ( name! )!",
        "trait_body = {! constraint! more_constraints }!",
          "more_constraints? = constraint more_constraints",
          "constraint = ab_struct | ab_class | union",
            "ab_struct = @struct is {! field_protos static_interface }!",
              "is? = @is trait",
              "field_protos = field_proto field_protos",
                "field_proto = LF name :! type!",
              "static_interface? = _at @static! {! protos }!",
                "protos? = proto protos",
                  "proto = name paralist! ret!",
            "ab_class = class_kind is {! protos static_interface }!",
            "union = @union {! typelist! }!",
      "decl_const = @const name =! const_expr!",
    /* Group: Command */
    "command = LF cmd_group1 | LF cmd_group2 | LF cmd_group3",
      "cmd_group1 = cmd_cond | cmd_loop | cmd_loop_ctrl",
        "cmd_cond = cmd_if",
          "cmd_if = @if expr! block! elifs else",
            "elifs? = elif elifs",
              "elif = @else @if expr! block!",
            "else? = @else block!",
        "cmd_loop = cmd_while | cmd_for",
          "cmd_while = while_typical | while_cast",
            "while_typical = @while expr! block!",
            "while_cast = @while name! =! try_cast block!",
              "try_cast = ?! [! cast_type! ]! expr!",
          "cmd_for = @for for_params! @in expr! block!",
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
          "cmd_let = binding_flag @let pattern =! expr!",
            "binding_flag? = _at @static | _at @data",
			"pattern = name | { namelist }!",
          "cmd_initial = binding_flag @initial pattern =! expr!",
          "cmd_reset = @reset pattern reset_operator =! expr!",
            "reset_operator? = op_arith",
          // operand_body -> Group: Operand
        "cmd_pass = @do @nothing",
        "cmd_side_effect = expr",
    /* Group: Expression */
    "const_expr = const | const_cast | const_bundle | const_tuple",
      "const_cast = [ type! ]! const_expr!",
      "const_bundle = { } | { const_pairlist! }!",
        "const_pairlist = const_pair const_pairlist_tail",
          "const_pairlist_tail? = , const_pair! const_pairlist_tail",
          "const_pair = name :! const_expr!",
      "const_tuple = ^ ( ) | ^ (! const_exprlist! )!",
        "const_exprlist = const_expr! const_exprlist_tail",
          "const_exprlist_tail? = , const_expr const_exprlist_tail",
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
      "unary? = @not | _exc2 | + | - | <- | @try",
      "operand_body = lambda | wrapped | cast | callcc | misc | variable",
        "lambda = lambda_iterator | lambda_typical",
          "lambda_iterator = $ -> opt_type body!",
            "opt_type? = type",
          "lambda_typical = proc_flag paralist_weak ret_weak block_or_expr",
            "proc_flag? = &",
            "paralist_weak? = name | ( ) | ( weak_param more_weak_params )",
              "more_weak_params? = , weak_param! more_weak_params",
              "weak_param = name : type! | name",
            "ret_weak? = : type",
            "block_or_expr = -> block! | => expr!",
        "wrapped = ( expr! )!",
        "cast = cast_flag [ cast_type ]! expr!",
          "cast_flag? = _exc1",
          "cast_type = narrow_type | type!",
            "narrow_type = - type",
        "callcc = @callcc NoLF [! type! ]! expr!",
        "misc = new | attached | literal | text | guard",
		  "new = @new type!",
		  "attached = _at type! :! name!",
		  "literal = bundle | tuple | iterator | const",
		    "bundle = { } | { pairlist! }!",
		      "pairlist? = pair pairlist_tail",
		        "pairlist_tail? = , pair! pairlist_tail",
		        "pair = pair_typical | pair_brief",
		          "pair_typical = name : expr!",
		          "pair_brief = name",
		    "tuple = ^ ( ) | ^ (! exprlist! )!",
              "exprlist = expr exprlist_tail",
                "exprlist_tail? = , expr! exprlist_tail",
            "iterator = $ ( ) | $ ( exprlist! )!",
            "const = string | int | float | bool",
              "string = String",
              "int = Dec | Hex | Oct | Bin",
              "float = Float | Exp",
              "bool = @Yes | @No",
          "text = Text | TxBegin first_segment more_segments TxEnd!",
            "first_segment = segment",
              "segment = segment_tag expr!",
                "segment_tag? = name : ",
            "more_segments? = next_segment more_segments",
              "next_segment = TxInner segment",
          "guard = when | switch",
	        "when = @when {! case! more_cases }!",
	          "more_cases? = case more_cases",
              "case = when_key block_or_expr!",
                "when_key = @otherwise | (! expr! )!",
            "switch = @switch expr! {! type_case! more_type_cases }!",
              "more_type_cases? = type_case more_type_cases",
              "type_case = switch_key switch_rename block_or_expr!",
                "switch_key = @otherwise | [! type! ]!",
                "switch_rename? = ( name )!",
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
