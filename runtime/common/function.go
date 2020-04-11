package common

import (
	"fmt"
	. "kumachan/error"
	"strings"
)


type Function struct {
	IsNative     bool
	NativeIndex  int
	Code         [] Instruction
	BaseSize     FrameBaseSize
	Info         FuncInfo
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

func (f *Function) ToValue(native_registry func(int)Value) Value {
	if f.IsNative {
		return native_registry(f.NativeIndex)
	} else {
		return FunctionValue {
			Underlying:    f,
			ContextValues: make([]Value, 0, 0),
		}
	}
}

func (f *Function) String() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "proc %d %d:", f.BaseSize.Context, f.BaseSize.Reserved)
	fmt.Fprintf(&buf, "   ; %s [%s]", f.Info.Name, f.Info.Module)
	var point = f.Info.DeclPoint.Node.Point
	var file = f.Info.DeclPoint.Node.CST.Name
	fmt.Fprintf(&buf, " at (%d, %d) in %s", point.Row, point.Col, file)
	buf.WriteRune('\n')
	if f.IsNative {
		fmt.Fprintf(&buf, "    NATIVE %d", f.NativeIndex)
		return buf.String()
	} else {
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
	}
}
