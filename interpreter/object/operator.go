package object

type CustomOperator int
const (
    CO_Equal CustomOperator = iota
    CO_LessThan
    CO_Negate
    CO_Plus
    CO_Minus
    CO_Times
    CO_Divide
    CO_Power
)
