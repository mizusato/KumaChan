package node


func (impl Match) Term() {}
type Match struct {
	Node                      `part:"match"`
	Argument  Expr            `part:"expr"`
	Branches  [] Branch       `list_more:"branch_list" item:"branch"`
}
type Branch struct {
	Node                    `part:"branch"`
	Type     MaybeRef       `part_opt:"branch_key.type_ref.ref"`
	Pattern  MaybePattern   `part_opt:"branch_key.opt_pattern.pattern"`
	Expr     Expr           `part:"expr"`
}


func (impl If) Term() {}
type If struct {
	Node              `part:"if"`
	Condition  Expr   `part:"if_cond.expr"`
	YesBranch  Expr   `part:"if_yes.expr"`
	NoBranch   Expr   `part:"if_no.expr"`
}
