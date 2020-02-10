package node


func (impl Match) Term() {}
type Match struct {
	Node                      `part:"match"`
	Argument  VariousTerm     `part:"term"`
	Branches  [] Branch       `list_more:"branch_list" item:"branch"`
}
type Branch struct {
	Node                      `part:"branch"`
	Type     MaybeType        `part_opt:"branch_key.type"`
	Pattern  MaybePattern     `part_opt:"opt_pattern.pattern"`
	Expr     MaybeExpr        `part_opt:"branch_value.expr"`
}
type MaybePattern interface { MaybePattern() }


func (impl If) Term() {}
type If struct {
	Node              `part:"if"`
	Condition  Expr   `part:"if_cond.expr"`
	YesBranch  Expr   `part:"if_yes.expr"`
	NoBranch   Expr   `part:"if_no.expr"`
}
