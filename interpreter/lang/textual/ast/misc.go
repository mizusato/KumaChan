package ast


func (impl Lambda) Body() {}
func (impl Lambda) Term() {}
type Lambda struct {
	Node                     `part:"lambda"`
	Input   VariousPattern   `part:"pattern"`
	Output  Expr             `part:"expr"`
}

func (impl ConstructorLambda) Term() {}
type ConstructorLambda struct {
	Node             `part:"ctor_lambda"`
	Type   TypeRef   `part:"type_ref"`
	Exact  bool      `option:"ctor_modifier.@exact"`
}

func (impl PipelineLambda) Term() {}
type PipelineLambda struct {
	Node                       `part:"pipeline_lambda"`
	Pipeline  [] VariousPipe   `list_rec:"pipes"`
}

func (impl Block) Term() {}
type Block struct {
	Node                   `part:"block"`
	Bindings  [] Binding   `list_more:"" item:"binding"`
	Return    Expr         `part:"block_value.expr"`
}
type Binding struct {
	Node                        `part:"binding"`
	Recursive  bool             `option:"binding_type.rec_opt.@rec"`
	Pattern    VariousPattern   `part:"pattern"`
	Type       MaybeType        `part_opt:"binding_type.type"`
	Value      Expr             `part:"expr"`
}

func (impl Cps) Term() {}
type Cps struct {
	Node                       `part:"cps"`
	Callee   InlineRef         `part:"inline_ref"`
	Binding  MaybeCpsBinding   `part_opt:"cps_binding"`
	Input    Expr              `part:"cps_input.expr"`
	Output   Expr              `part:"cps_output.expr"`
}
type MaybeCpsBinding interface { Maybe(CpsBinding,MaybeCpsBinding) }
func (impl CpsBinding) Maybe(CpsBinding,MaybeCpsBinding) {}
type CpsBinding struct {
	Node                      `part:"cps_binding"`
	Pattern  VariousPattern   `part:"pattern"`
	Type     MaybeType        `part_opt:"binding_type.type"`
}


func (impl InlineRef) Term() {}
type InlineRef struct {
	Node                       `part:"inline_ref"`
	Module    Identifier       `part_opt:"module_prefix.name"`
	Id        Identifier       `part:"name"`
	TypeArgs  [] VariousType   `list_more:"inline_type_args" item:"type"`
}

func (impl Array) Term() {}
type Array struct {
	Node             `part:"array"`
	Items  [] Expr   `list_more:"exprlist" item:"expr"`
}

func (impl IntegerLiteral) Term() {}
type IntegerLiteral struct {
	Node             `part:"int"`
	Value  [] rune   `content:"Int"`
}

func (impl FloatLiteral) Term() {}
type FloatLiteral struct {
	Node             `part:"float"`
	Value  [] rune   `content:"Float"`
}

