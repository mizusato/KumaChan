package node


func (impl Lambda) Body() {}
func (impl Lambda) Term() {}
type Lambda struct {
	Node                    `part:"lambda"`
	Input   VariousPattern  `part:"pattern"`
	Output  Pipe            `part:"pipe"`
}

func (impl Block) Term() {}
type Block struct {
	Node                   `part:"block"`
	Bindings  [] Binding   `list_more:"" item:"binding"`
	Return    Expr         `part:"return.expr"`
}
type Binding struct {
	Node                      `part:"binding"`
	Pattern  VariousPattern   `part:"pattern"`
	Type     MaybeType        `part_opt:"binding_type.type"`
	Value    Expr             `part:"expr"`
}

func (impl With) Term() {}
type With struct {
	Node                `part:"with"`
	Unboxes  [] Unbox   `list_more:"" item:"unbox"`
	Return   Expr       `part:"return.expr"`
}
type Unbox struct {
	Node                    `part:"unbox"`
	Pattern  MaybePattern   `part_opt:"unbox_assign.pattern"`
	Type     MaybeType      `part_opt:"unbox_assign.binding_type.type"`
	Value    Expr           `part:"expr"`
}

func (impl Text) Term() {}
type Text struct {
	Node                `part:"text"`
	Template  [] rune   `content:"Text"`
}

func (impl Ref) Term() {}
type Ref struct {
	Node                       `part:"ref"`
	Module    Identifier       `part_opt:"module_prefix.name"`
	Specific  bool             `option:"module_prefix.::"`
	Id        Identifier       `part:"name"`
	TypeArgs  [] VariousType   `list_more:"type_args" item:"type"`
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