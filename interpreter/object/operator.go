package object

type CustomOperator int
const (
    CO_Plus CustomOperator = iota
    CO_Minus
    CO_Times
    CO_Divide
    CO_Power
)
