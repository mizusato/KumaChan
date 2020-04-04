package ast


func (impl Switch) Term() {}
type Switch struct {
	Node                      `part:"switch"`
	Argument  VariousTerm     `part:"term"`
	Branches  [] Branch       `list_more:"branch_list" item:"branch"`
}
type Branch struct {
	Node                    `part:"branch"`
	Type     MaybeRef       `part_opt:"branch_key.type_ref.ref"`
	Pattern  MaybePattern   `part_opt:"branch_key.opt_pattern.pattern"`
	Expr     Expr           `part:"expr"`
}

func (impl MultiSwitch) Term() {}
type MultiSwitch struct {
	Node                        `part:"multi_switch"`
	Arguments  [] Expr          `list_more:"exprlist" item:"expr"`
	Branches   [] MultiBranch   `list_more:"multi_branch_list" item:"multi_branch"`
}
type MultiBranch struct {
	Node                    `part:"multi_branch"`
	Types    [] Ref         `list_more:"branch_key.multi_type_ref.ref_list" item:"ref"`
	Pattern  PatternTuple   `part:"branch_key.pattern_tuple"`
	Expr     Expr           `part:"expr"`
}


func (impl If) Term() {}
type If struct {
	Node                     `part:"if"`
	Condition  VariousTerm   `part:"term"`
	YesBranch  Expr          `part:"if_yes.expr"`
	NoBranch   Expr          `part:"if_no.expr"`
	ElIfs      [] ElIf       `list_rec:"elifs"`
}

type ElIf struct {
	Node                     `part:"elif"`
	Condition  VariousTerm   `part:"term"`
	YesBranch  Expr          `part:"expr"`
}
