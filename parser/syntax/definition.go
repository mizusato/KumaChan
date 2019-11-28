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
    Token { Name: "<-",      Pattern: r(`\<\-`) },
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
    Token { Name: "++",      Pattern: r(`\+\+`) },   // Pull from Iterator
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
        "tuple_t = ^ [! typelist! ]!",
        "function_t = [ signature ]!",
          "signature = -> type! | typelist! ->! type!",
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
      "decl_type = singleton | struct | type_alias",
        "singleton = attrs @singleton namelist",
          "namelist = name namelist_tail",
            "namelist_tail? = , name! namelist_tail",
            "name = Name",
        "struct = attrs @struct name type_params is {! members }!",
          "is? = @is traitlist",
            "traitlist = trait traitlist_tail",
              "traitlist_tail? = , trait traitlist_tail",
          "members? = member members",
            "member = field | method",
            "field = LF name :! type! field_default",
              "field_default? = ( const_expr! )!",
                // const_expr -> Group: Expression
              "methods? = method methods",
                "method = prereq opt_static name paralist! ret body!",
                  "prereq = type_params",
                  "opt_static? = @static",
      "decl_trait = interface | union",
        "interface = @struct is {! protos }!",
          "proto? = proto protos",
            "proto = field_proto | method_proto",
              "field_proto = LF name :! type!",
              "method_proto = LF opt_static name paralist! ret!",
        "union = @union {! typelist! }!",
      "decl_const = @const name =! const_expr!",
    /* Group: Command */
    "command = LF cmd_group1 | LF cmd_group2 | LF cmd_group3",
      "cmd_group1 = cmd_cond | cmd_loop | cmd_loop_ctrl",
        "cmd_cond = cmd_if | cmd_if_else",
          "cmd_if = if { if_branch! more_if_branches }!",
            "more_if_braches? = if_branch more_if_branches",
            "if_branch = cond block!",
              "cond = cast_cond | expr",
                "cast_cond = ? [! cast_type! ]! name! =! expr!",
          "cmd_if_else = @if cond! block! @else! block!",
        "cmd_loop = cmd_while | cmd_for",
          "cmd_while = @while cond! block!",
          "cmd_for = @for pattern! @in! expr! block!",
            "pattern = name | tuple_pattern | bundle_pattern",
              "tuple_pattern = ( namelist! )!",
              "bundle_pattern = { namelist! }!",
        "cmd_loop_ctrl = @break | @continue",
      "cmd_group2 = cmd_return | cmd_yield | cmd_abort | cmd_assert",
        "cmd_return = @return return_content",
          "return_content = Void | expr!",
        "cmd_yield = @yield expr!",
        "cmd_abort = @abort expr!",
        "cmd_assert = @assert expr!",
      "cmd_group3 = cmd_scope | cmd_pass | cmd_side_effect",
        "cmd_scope = cmd_let | cmd_initial | cmd_reset",
          "cmd_let = @let pattern! =! expr!",
          "cmd_initial = binding_flag @initial pattern! =! expr!",
          "cmd_reset = @reset pattern! reset_operator =! expr!",
            "reset_operator? = op_arith",
          // operand_body -> Group: Operand
        "cmd_pass = @do @nothing",
        "cmd_side_effect = expr",
    /* Group: Expression */
    "const_expr = const | const_cast | const_bundle | const_tuple | const_name",
      "const_cast = [ type! ]! const_expr!",
      "const_bundle = { } | { const_pairlist! }!",
        "const_pairlist = const_pair const_pairlist_tail",
          "const_pairlist_tail? = , const_pair! const_pairlist_tail",
          "const_pair = name :! const_expr!",
      "const_tuple = ^ ( ) | ^ (! const_exprlist! )!",
        "const_exprlist = const_expr! const_exprlist_tail",
          "const_exprlist_tail? = , const_expr const_exprlist_tail",
      "const_name = module_prefix name",
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
      "unary? = @not | _exc2 | + | - | ++",
      "operand_body = lambda | wrapped | misc | variable",
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
        "misc = literal | text | guard | cast | callcc",
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
          "guard = when | type_switch",
	        "when = @when {! case! more_cases }!",
	          "more_cases? = case more_cases",
              "case = when_key :! expr!",
                "when_key = @otherwise | ( expr! )!",
            "type_switch = [ ? ]! expr! {! type_case! more_type_cases }!",
              "more_type_cases? = type_case more_type_cases",
              "type_case = switch_key switch_rename :! expr!",
                "switch_key = @otherwise | [ type! ]!",
                "switch_rename? = name | ( name )!",
          "cast = force_flag [ cast_type ]! expr!",
            "force_flag? = _exc1",
            "cast_type = narrow_type | type!",
              "narrow_type = - type!",
          "callcc = @callcc NoLF [! type! ]! expr!",
        "variable = module_prefix name type_args",
      "accesses? = access accesses",
        "access = . name! access_action nullable_opt",
          "access_action? = method_call | assignment",
            "method_call = call",
			  "call = NoLF ( arglist )! | { arglist }!",
			    "arglist? = exprlist",
			"assignment = = expr!",
		  "nullable_opt? = ?",
      "calls? = call calls",
      "with? = @with {! temps }!",
        "temps? = temp temps",
          "temp = name =! expr!",
      "pipelines? = pipeline pipelines",
        "pipeline = _bar1 operand_body pipeline_args",
        "pipeline_args? = NoLF ( arglist )! | { arglist }!",
}
