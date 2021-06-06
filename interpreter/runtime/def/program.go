package def

import (
	"fmt"
	"strings"
)


type Program struct {
	MetaData   ProgramMetaData
	Functions  [] FunctionSeed
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
	for _, item := range p.Functions {
		fmt.Fprintf(&buf, "[%d] %s\n", i, item.String())
		i += 1
	}
	buf.WriteString(";; program end\n")
	return buf.String()
}
