package ast


func (impl Tuple) Term() {}
type Tuple struct {
	Node                `part:"tuple"`
	Elements  [] Expr   `list_more:"exprlist" item:"expr"`
}

func (impl Bundle) Term() {}
type Bundle struct {
	Node                    `part:"bundle"`
	Update  MaybeUpdate     `part_opt:"update"`
	Values  [] FieldValue   `list_more:"pairlist" item:"pair"`
}
type FieldValue struct {
	Node                `part:"pair"`
	Key    Identifier   `part:"name"`
	Value  MaybeExpr    `part_opt:"expr"`
}
type MaybeUpdate interface { Maybe(Update,MaybeUpdate) }
func (impl Update) Maybe(Update,MaybeUpdate) {}
type Update struct {
	Node         `part:"update"`
	Base  Expr   `part:"expr"`
}

