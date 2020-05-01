package ast


func (impl Cast) Term() {}
type Cast struct {
	Node                  `part:"cast"`
	Target  VariousType   `part:"type"`
	Object  Expr          `part:"expr"`
}

func (impl Lambda) Body() {}
func (impl Lambda) Term() {}
type Lambda struct {
	Node                     `part:"lambda"`
	Input   VariousPattern   `part:"pattern"`
	Output  Expr             `part:"expr"`
}

func (impl Block) Term() {}
type Block struct {
	Node                  `part:"block"`
	Bindings [] Binding   `list_more:"" item:"binding"`
	Return   Expr         `part:"return.expr"`
}
type Binding struct {
	Node                     `part:"binding"`
	Recursive bool           `option:"binding_type.rec_opt.@rec"`
	Pattern   VariousPattern `part:"pattern"`
	Type      MaybeType      `part_opt:"binding_type.type"`
	Value     Expr           `part:"expr"`
}

func (impl Cps) Term() {}
type Cps struct {
	Node                       `part:"cps"`
	Callee   InlineRef         `part:"inline_ref"`
	Binding  MaybeCpsBinding   `part_opt:"cps_binding"`
	Input    Expr              `part:"cps_input.expr"`
	Output   Expr              `part:"cps_output.expr"`
}
type MaybeCpsBinding interface { MaybeCpsBinding() }
func (impl CpsBinding) MaybeCpsBinding() {}
type CpsBinding struct {
	Node                      `part:"cps_binding"`
	Pattern  VariousPattern   `part:"pattern"`
	Type     MaybeType        `part_opt:"binding_type.type"`
}

func (impl Array) Term() {}
type Array struct {
	Node            `part:"array"`
	Items  []Expr   `list_more:"exprlist" item:"expr"`
}

func (impl Infix) Term() {}
type Infix struct {
	Node                    `part:"infix"`
	Operand1  VariousTerm   `part:"operand1.term"`
	Operand2  VariousTerm   `part:"operand2.term"`
	Operator  VariousTerm   `part:"operator.term"`
}

func (impl Text) Term() {}
type Text struct {
	Node                `part:"text"`
	Template  [] rune   `content:"Text"`
}

func (impl InlineRef) Term() {}
type InlineRef struct {
	Node                       `part:"inline_ref"`
	Module    Identifier       `part_opt:"module_prefix.name"`
	Specific  bool             `option:"module_prefix.::"`
	Id        Identifier       `part:"name"`
	TypeArgs  [] VariousType   `list_more:"inline_type_args" item:"type"`
}

func (impl VariousLiteral) Term() {}
type VariousLiteral struct {
	Node               `part:"literal"`
	Literal  Literal   `use:"first"`
}
type Literal interface { Literal() }


func (impl IntegerLiteral) Literal() {}
type IntegerLiteral struct {
	Node             `part:"int"`
	Value  [] rune   `content:"Int"`
}

func (impl FloatLiteral) Literal() {}
type FloatLiteral struct {
	Node             `part:"float"`
	Value  [] rune   `content:"Float"`
}

func (impl StringLiteral) Literal() {}
type StringLiteral struct {
	Node             `part:"string"`
	Value  [] rune   `content:"String"`
}

func (impl CharLiteral) Literal() {}
type CharLiteral struct {
	Node             `part:"char"`
	Value  [] rune   `content:"Char"`
}
