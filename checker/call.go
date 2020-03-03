package checker


func (impl Call) ExprVal() {}
type Call struct {
	Caller  Expr
	Callee  Expr
}
