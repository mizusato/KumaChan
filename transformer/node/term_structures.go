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
	Node                    `part:"bundle"`
	Update    MaybeUpdate   `part_opt:"update"`
	Records   [] Record     `list_more:"pairlist" item:"pair"`
}
type MaybeUpdate interface { MaybeUpdate() }
func (impl Update) MaybeUpdate() {}
type Update struct {
	Node              `part:"update"`
	Term  Term        `part:"term"`
	Base  [] Member   `list_rec:"base_members.members"`
	Path  [] Member   `list_rec:"members"`
}
type Member struct {
	Node                   `part:"member"`
	Optional  bool         `option:"opt.?"`
	Name      Identifier   `part:"name"`
}
type Record struct {
	Node                `part:"pair"`
	Key    Identifier   `part:"name"`
	Value  MaybeExpr    `part_opt:"expr"`
}

func (impl Get) Term() {}
type Get struct {
	Node              `part:"get"`
	Base  Term        `part:"term"`
	Path  [] Member   `list_rec:"members"`
}
