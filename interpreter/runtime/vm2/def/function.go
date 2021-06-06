package def

import (
	"fmt"
	"strings"
	"strconv"
	. "kumachan/standalone/util/error"
)


type FunctionEntity struct {
	Code Code
	FunctionEntityInfo
}
type FunctionInfo struct {
	Symbol  Symbol
	Name    string
	Decl    ErrorPoint
	SrcMap  [] ErrorPoint  // TODO: reduce memory usage
}
type FunctionEntityInfo struct {
	FunctionInfo
	IsEffect       bool
	ContextLength  LocalSize
}

type FunctionSeed interface {
	FunctionSeed()
	fmt.Stringer
}

func (_ *FunctionSeedLibraryNative) FunctionSeed() {}
func (s *FunctionSeedLibraryNative) String() string { return strconv.Quote(s.Id) }
type FunctionSeedLibraryNative struct {
	Id    string
	Info  FunctionInfo
}

func (_ *FunctionSeedGeneratedNative) FunctionSeed() {}
func (s *FunctionSeedGeneratedNative) String() string { return s.Data.String() }
type FunctionSeedGeneratedNative struct {
	Data  GeneratedNativeFunctionSeed
	Info  FunctionInfo
}
type GeneratedNativeFunctionSeed interface {
	GeneratedNativeFunctionSeed()
	fmt.Stringer
}

func (*FunctionSeedUsual) FunctionSeed() {}
type FunctionSeedUsual struct {
	Trunk   *BranchData
	Static  [] StaticValueSeed
	CtxLen  LocalSize
}
type BranchData struct {
	InstList   [] Instruction
	ExtIdxMap  ExternalIndexMapping
	Stages     [] Stage
	Branches   [] *BranchData
	Closures   [] *FunctionSeedUsual
	Info       FunctionInfo
}
type StaticValueSeed interface {
	Evaluate() Value
	fmt.Stringer
}

func GetTrunkSymbol(seed *FunctionSeedUsual) Symbol {
	return seed.Trunk.Info.Symbol
}
func CreateFunctionEntity(seed *FunctionSeedUsual) *FunctionEntity {
	var static = make(AddrSpace, len(seed.Static))
	for i, s := range seed.Static {
		static[i] = s.Evaluate()
	}
	var frame_size = uint(0)
	var f = createFunctionEntity(0, &frame_size, seed.Trunk, static)
	f.Code.frameSize = LocalSize(frame_size)
	f.ContextLength = seed.CtxLen
	return f
}
func createFunctionEntity(offset uint, fs *uint, this *BranchData, static AddrSpace) *FunctionEntity {
	var required_fs = offset + uint(len(this.InstList))
	if required_fs >= MaxFrameValues { panic("frame too big") }
	if required_fs > *fs {
		*fs = required_fs
	}
	var branches = make([] *FunctionEntity, len(this.Branches))
	for i, b := range this.Branches {
		branches[i] = createFunctionEntity(required_fs, fs, b, static)
	}
	var closures = make([] *FunctionEntity, len(this.Closures))
	for i, cl := range this.Closures {
		closures[i] = CreateFunctionEntity(cl)
	}
	return &FunctionEntity {
		Code: Code {
			dstOffset: LocalSize(offset),
			instList:  this.InstList,
			extIdxMap: this.ExtIdxMap,
			stages:    this.Stages,
			branches:  branches,
			closures:  closures,
			frameSize: 0,
			static:    static,
		},
		FunctionEntityInfo: FunctionEntityInfo {
			FunctionInfo:  this.Info,
			ContextLength: 0,
		},
	}
}

func (*UiObjectSeed) GeneratedNativeFunctionSeed() {}
type UiObjectSeed struct {
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
func (seed *UiObjectSeed) String() string {
	return fmt.Sprintf("%s %s %s",
		strconv.Quote(seed.Object),
		strconv.Quote(seed.Group.GroupName),
		strconv.Quote(seed.Group.BaseDir))
}

func (f *FunctionSeedUsual) String() string {
	var buf strings.Builder
	buf.WriteString(".function")
	buf.WriteRune('\n')
	f.writeContent(&buf)
	return buf.String()
}

func (f *FunctionSeedUsual) writeContent(buf *strings.Builder) string {
	fmt.Fprintf(buf, "   .FUNC %d   ; %s", f.CtxLen, f.Trunk.Info.Name)
	var point = f.Trunk.Info.Decl.Node.Point
	var file = f.Trunk.Info.Decl.Node.CST.Name
	fmt.Fprintf(buf, " at (%d, %d) in %s", point.Row, point.Col, file)
	buf.WriteRune('\n')
	buf.WriteString(".static")
	buf.WriteRune('\n')
	for i, s := range f.Static {
		fmt.Fprintf(buf, "    [%d] %s", i, s)
	}
	buf.WriteString(".code")
	buf.WriteRune('\n')
	writeBranchData(buf, 0, [] uint {}, f.Trunk)
	return buf.String()
}

func writeBranchData (
	buf     *strings.Builder,
	offset  uint,
	path    [] uint,
	branch  *BranchData,
) {
	if len(path) == 0 {
		buf.WriteString(".branch-trunk")
	} else {
		var t = make([] string, len(path))
		for i, index := range path {
			t[i] = fmt.Sprint(index)
		}
		fmt.Fprintf(buf, ".branch-%s", strings.Join(t, "-"))
	}
	buf.WriteRune('\n')
	if len(branch.ExtIdxMap) != 0 {
		buf.WriteString(".ext")
		buf.WriteRune('\n')
		for i, m := range branch.ExtIdxMap {
			var default_target string
			if m.HasDefault {
				default_target = "()"
			} else {
				default_target = fmt.Sprintf("()[%d]", m.Default)
			}
			var targets = make([] string, 0)
			for vec, target := range m.VectorMap {
				var t = make([] string, 0)
				for j, idx := range vec.Decode() {
					t[j] = fmt.Sprint(idx)
				}
				var vec_str = strings.Join(t, "-")
				var target_ = fmt.Sprintf("(%s)[%d]", vec_str, target)
				targets = append(targets, target_)
			}
			var targets_ = strings.Join(targets, " ")
			fmt.Fprintf(buf, "    [%d] %s %s", i, default_target, targets_)
			buf.WriteRune('\n')
		}
	}
	var flow_map = make([] string, len(branch.InstList))
	var gen_flow_map func([] uint, [] Stage)
	gen_flow_map = func(path ([] uint), stages ([] Stage)) {
		for i, stage := range branch.Stages {
			for j, flow := range stage {
				var i_j_path = make([] uint, len(path), len(path) + 2)
				copy(i_j_path, path)
				i_j_path = append(i_j_path, uint(i), uint(j))
				if flow.Simple() {
					var t = make([] string, len(i_j_path))
					for k, index := range i_j_path {
						t[k] = fmt.Sprint(index)
					}
					var i_j_path_ = strings.Join(t, "-")
					for k := flow.Start; k <= flow.End; k += 1 {
						if flow_map[k] == "" {
							flow_map[k] = i_j_path_
						} else {
							flow_map[k] = "(ERROR)"
						}
					}
				} else {
					gen_flow_map(i_j_path, flow.Stages)
				}
			}
		}
	}
	gen_flow_map([] uint {}, branch.Stages)
	var prev_flow = ""
	for i, inst := range branch.InstList {
		var this_flow = flow_map[i]
		if this_flow != prev_flow {
			prev_flow = this_flow
			fmt.Fprintf(buf, "    .FLOW %s", this_flow)
			buf.WriteRune('\n')
		}
		fmt.Fprintf(buf, "    [%d] %s", (offset + uint(i)), inst.String())
		if i < len(branch.Info.SrcMap) {
			var point = branch.Info.SrcMap[i]
			var n = point.Node
			fmt.Fprintf(buf, "   ; (%d, %d)", n.Point.Row, n.Point.Col)
			if point.Node.CST != branch.Info.Decl.Node.CST {
				fmt.Fprintf(buf, " in %s", point.Node.CST.Name)
			}
		}
		buf.WriteRune('\n')
	}
	for i, b := range branch.Branches {
		var b_path = make([] uint, len(path), (len(path) + 1))
		copy(b_path, path)
		b_path = append(b_path, uint(i))
		var b_offset = (offset + uint(len(branch.InstList)))
		writeBranchData(buf, b_offset, b_path, b)
	}
	for i, cl := range branch.Closures {
		fmt.Fprintf(buf, ".closure-%d", i)
		cl.writeContent(buf)
	}
}


