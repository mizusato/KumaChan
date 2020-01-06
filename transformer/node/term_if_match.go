package node


func (impl If) Term() {}
type If struct {
	Node                     `part:"if"`
	Argument   Tuple         `part:"tuple"`
	Branches   [] IfBranch   `list_more:"if_branch_list" item:"if_branch"`
	ElseValue  Expr          `part:"else.expr"`
}
type IfBranch struct {
	Node               `part:"if_branch"`
	Predicate  Tuple   `part:"tuple"`
	Value      Expr    `part:"expr"`
}

func (impl Match) Term() {}
type Match struct {
	Node                       `part:"match"`
	Argument  Tuple            `part:"tuple"`
	Branches  [] MatchBranch   `list_more:"match_branch_list" item:"match_branch"`
	Default   MaybeExpr        `part_opt:"default.expr"`
}
type MatchBranch struct {
	Node                      `part:"match_branch"`
	Type     ReprTuple        `part:"repr_tuple"`
	Pattern  VariousPattern   `part:"pattern"`
	Expr     Expr             `part:"expr"`
}
