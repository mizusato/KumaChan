package def

import (
	"fmt"
	"strings"
)


type Program struct {
	MetaData    ProgramMetaData
	DataValues  [] DataValue
	Functions   [] *Function
	Closures    [] *Function
	Effects     [] *Function
	KmdInfo
	RpcInfo
}
type ProgramMetaData struct {
	EntryModulePath  string
}

type DataValue interface {
	fmt.Stringer
	ToValue()  Value
}

func (p Program) String() string {
	var i = 0
	var buf strings.Builder
	buf.WriteString(";; program begin\n")
	buf.WriteString("data:\n")
	for _, item := range p.DataValues {
		fmt.Fprintf(&buf, "    [%d] %s\n", i, item.String())
		i += 1
	}
	buf.WriteString(";;\n")
	buf.WriteString("code functions:\n")
	for _, item := range p.Functions {
		fmt.Fprintf(&buf, "[%d] %s\n", i, item.String())
		i += 1
	}
	buf.WriteString(";;\n")
	buf.WriteString("code closures:\n")
	for _, item := range p.Closures {
		fmt.Fprintf(&buf, "[%d] %s\n", i, item.String())
		i += 1
	}
	buf.WriteString(";;\n")
	buf.WriteString(";;\n")
	buf.WriteString("code effects:\n")
	for _, item := range p.Effects {
		fmt.Fprintf(&buf, "%s\n", item.String())
	}
	buf.WriteString(";; program end\n")
	return buf.String()
}
