package node


func (impl List) Term() {}
type List struct {
	Node             `part:"list"`
	Items  [] Expr   `list_more:"exprlist" item:"expr"`
}

func (impl Tuple) Term() {}
type Tuple struct {
	Node                `part:"tuple"`
	Elements  [] Expr   `list_more:"exprlist" item:"expr"`
}

func (impl Bundle) Term() {}
type Bundle struct {
	Node                      `part:"bundle"`
	Update    MaybeUpdate     `part_opt:"update.get"`
	Records   [] FieldValue   `list_more:"pairlist" item:"pair"`
}
type FieldValue struct {
	Node                `part:"pair"`
	Key    Identifier   `part:"name"`
	Value  MaybeExpr    `part_opt:"expr"`
}

type MaybeUpdate interface { MaybeUpdate() }
func (impl Get) MaybeUpdate() {}
func (impl Get) Term() {}
type Get struct {
	Node              `part:"get"`
	Base  Expr        `part:"expr"`
	Path  [] Member   `list_rec:"members"`
}
type Member struct {
	Node                   `part:"member"`
	Optional  bool         `option:"opt.?"`
	Name      Identifier   `part:"name"`
}
