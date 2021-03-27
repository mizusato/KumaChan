package lang

import (
	"fmt"
	"strconv"
	"strings"
	. "kumachan/misc/util/error"
)


type FunctionKind int
const (
	F_USER FunctionKind = iota
	F_NATIVE
	F_PREDEFINED
)
type Function struct {
	Kind        FunctionKind
	NativeId    string
	Predefined  interface{}
	Code        [] Instruction
	BaseSize    FrameBaseSize
	Info        FuncInfo
}

type FrameBaseSize struct {
	Context   Short
	Reserved  Long
}

type FuncInfo struct {
	Module     string
	Name       string
	DeclPoint  ErrorPoint
	SourceMap  [] ErrorPoint
}

type UiObjectThunk struct {
	Object  string
	Group   *UiObjectGroup
}

type UiObjectGroup struct {
	GroupName  string
	BaseDir    string
	XmlDef     string
	RootName   string
	Widgets    [] string
	Actions    [] string
}

func (f *Function) String() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "proc %d %d:", f.BaseSize.Context, f.BaseSize.Reserved)
	fmt.Fprintf(&buf, "   ; %s [%s]", f.Info.Name, f.Info.Module)
	var point = f.Info.DeclPoint.Node.Point
	var file = f.Info.DeclPoint.Node.CST.Name
	fmt.Fprintf(&buf, " at (%d, %d) in %s", point.Row, point.Col, file)
	buf.WriteRune('\n')
	switch f.Kind {
	case F_USER:
		for i, inst := range f.Code {
			fmt.Fprintf(&buf, "    [%d] %s", i, inst.String())
			if i < len(f.Info.SourceMap) {
				var point = f.Info.SourceMap[i]
				var n = point.Node
				fmt.Fprintf(&buf, "   ; (%d, %d)", n.Point.Row, n.Point.Col)
				if point.Node.CST != f.Info.DeclPoint.Node.CST {
					fmt.Fprintf(&buf, " in %s", point.Node.CST.Name)
				}
			}
			if i != len(f.Code)-1 {
				buf.WriteRune('\n')
			}
		}
		return buf.String()
	case F_NATIVE:
		fmt.Fprintf(&buf, "    NATIVE %s", strconv.Quote(f.NativeId))
		return buf.String()
	case F_PREDEFINED:
		fmt.Fprintf(&buf, "    PREDEFINED")
		return buf.String()
	default:
		panic("impossible branch")
	}
}
