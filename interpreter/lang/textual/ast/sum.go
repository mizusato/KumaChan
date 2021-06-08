package ast


func (impl Switch) Term() {}
type Switch struct {
	Node                        `part:"switch"`
	Argument  Expr              `part:"expr"`
	Branches  [] SwitchBranch   `list_more:"sw_branch_list" item:"sw_branch"`
}
type SwitchBranch struct {
	Node                   `part:"sw_branch"`
	Cases    [] TypeRef    `list_more:"sw_key.namelist" item:"name"`
	Pattern  MaybePattern  `part_opt:"sw_key.opt_pattern.pattern"`
	Expr     Expr          `part:"expr"`
}

func (impl Select) Term() {}
type Select struct {
	Node                         `part:"select"`
	Arguments  [] Expr           `list_more:"exprlist" item:"expr"`
	Branches   [] SelectBranch   `list_more:"sl_branch_list" item:"sl_branch"`
}
type SelectBranch struct {
	Node                         `part:"sl_branch"`
	Cases    [] SelectCase       `list_more:"sl_key.sl_case_list" item:"sl_case"`
	Pattern  MaybePatternTuple   `part_opt:"sl_key.sl_pattern.pattern_tuple"`
	Expr     Expr                `part:"expr"`
}
type SelectCase struct {
	Node                        `part:"sl_case"`
	Components  [] Identifier   `list_more:"namelist" item:"name"`
}

func (impl If) Term() {}
type If struct {
	Node                 `part:"if"`
	Condition  Expr      `part:"cond.expr"`
	YesBranch  Expr      `part:"if_yes.expr"`
	NoBranch   Expr      `part:"if_no.expr"`
	ElIfs      [] ElIf   `list_rec:"elifs"`
}
type ElIf struct {
	Node              `part:"elif"`
	Condition  Expr   `part:"cond.expr"`
	YesBranch  Expr   `part:"expr"`
}


