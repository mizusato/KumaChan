package common


type Program struct {
	DataValues  [] DataValue
	Functions   [] *Function
	Closures    [] *Function
	Constants   [] *Function
	Effects     [] *Function
}

type DataValue interface {
	ToValue()  Value
	// Marshal()  interface{}  // TODO (+Unmarshal)
}

func (p Program) String() string {
	return ""  // TODO
}
