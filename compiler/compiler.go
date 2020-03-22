package compiler

import (
	"kumachan/checker"
	"kumachan/loader"
	"kumachan/runtime/common"
	"kumachan/transformer/node"
)

type Code struct {
	InstSeq    [] common.Instruction
	SourceMap  [] *node.Node
}

func CodeSingleInst (inst common.Instruction, info checker.ExprInfo) Code {
	return Code {
		InstSeq:   [] common.Instruction { inst },
		SourceMap: [] *node.Node { &(info.ErrorPoint.Node) },
	}
}

type CodeBuffer struct {
	Code  *Code
}
func MakeCodeBuffer() CodeBuffer {
	var code = &Code {
		InstSeq:   make([] common.Instruction, 0),
		SourceMap: make([] *node.Node, 0),
	}
	return CodeBuffer { code }
}
func (buf CodeBuffer) Write(code Code) {
	var base = &(buf.Code.InstSeq)
	var base_size = uint(len(buf.Code.InstSeq))
	for _, inst := range code.InstSeq {
		if inst.OpCode == common.JIF || inst.OpCode == common.JMP {
			var dest_addr = (uint(inst.Arg1) + base_size)
			ValidateDestAddr(dest_addr)
			*base = append(*base, common.Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   common.Long(dest_addr),
			})
		} else {
			*base = append(*base, inst)
		}
	}
	buf.Code.SourceMap = append(buf.Code.SourceMap, code.SourceMap...)
}
func (buf CodeBuffer) Collect() Code {
	var code = buf.Code
	buf.Code = nil
	return *code
}

type Context struct {
	GlobalRefs  *([] GlobalRef)
	LocalScope  *Scope
}

func (ctx Context) MakeClosure() Context {
	var refs = make([] GlobalRef, 0)
	return Context {
		GlobalRefs: &refs,
		LocalScope: MakeClosureScope(ctx.LocalScope),
	}
}

func (ctx Context) AppendDataRef(v common.DataValue) uint {
	var refs = ctx.GlobalRefs
	var index = uint(len(*refs))
	*refs = append(*refs, RefData { v })
	return index
}

func (ctx Context) AppendFunRef(ref checker.RefFunction) uint {
	var refs = ctx.GlobalRefs
	var index = uint(len(*refs))
	*refs = append(*refs, RefFun(ref))
	return index
}

func (ctx Context) AppendConstRef(ref checker.RefConstant) uint {
	var refs = ctx.GlobalRefs
	var index = uint(len(*refs))
	*refs = append(*refs, RefConst(ref))
	return index
}

func (ctx Context) AppendClosureRef(f *common.Function) uint {
	var refs = ctx.GlobalRefs
	var index = uint(len(*refs))
	*refs = append(*refs, RefClosure { f })
	return index
}

type GlobalRef interface { GlobalRef() }
func (impl RefData) GlobalRef() {}
type RefData struct { common.DataValue }
func (impl RefFun) GlobalRef() {}
type RefFun  checker.RefFunction
func (impl RefConst) GlobalRef() {}
type RefConst  checker.RefConstant
func (impl RefClosure) GlobalRef() {}
type RefClosure struct { *common.Function }

type DataInteger  checker.IntLiteral
func (d DataInteger) ToValue() common.Value {
	return d.Value
}
type DataSmallInteger  checker.SmallIntLiteral
func (d DataSmallInteger) ToValue() common.Value {
	return d.Value
}
type DataFloat  checker.FloatLiteral
func (d DataFloat) ToValue() common.Value {
	return d.Value
}

type Scope struct {
	Bindings    [] Binding
	BindingMap  map[string] uint
	BindingPeek *uint
}
type Binding struct {
	Name  string
	Used  bool
}
func MakeClosureScope(outer *Scope) *Scope {
	var bindings = make([] Binding, 0)
	for i, e := range outer.Bindings {
		bindings[i] = Binding {
			Name: e.Name,
			Used: false,
		}
	}
	var binding_map = make(map[string] uint)
	for k, v := range outer.BindingMap {
		binding_map[k] = v
	}
	var peek = uint(0)
	return &Scope {
		Bindings:    bindings,
		BindingMap:  binding_map,
		BindingPeek: &peek,
	}
}
func MakeBranchScope(outer *Scope) *Scope {
	var bindings = make([] Binding, 0)
	copy(bindings, outer.Bindings)
	var binding_map = make(map[string] uint)
	for k, v := range outer.BindingMap {
		binding_map[k] = v
	}
	return &Scope {
		Bindings:    bindings,
		BindingMap:  binding_map,
		BindingPeek: outer.BindingPeek,
	}
}
func (scope *Scope) AddBinding(name string) uint {
	var _, exists = scope.BindingMap[name]
	if exists { panic("duplicate binding") }
	var list = &(scope.Bindings)
	var offset = uint(len(*list))
	*list = append(*list, Binding {
		Name: name,
		Used: false,
	})
	scope.BindingMap[name] = offset
	*(scope.BindingPeek) += 1
	return offset
}

func Compile(expr checker.Expr, ctx Context) Code {
	switch v := expr.Value.(type) {
	case checker.IntLiteral:
		var index = ctx.AppendDataRef(DataInteger(v))
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.SmallIntLiteral:
		var index = ctx.AppendDataRef(DataSmallInteger(v))
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.FloatLiteral:
		var index = ctx.AppendDataRef(DataFloat(v))
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.RefFunction:
		var index = ctx.AppendFunRef(v)
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.RefConstant:
		var index = ctx.AppendConstRef(v)
		return CodeSingleInst(InstGlobalRef(index), expr.Info)
	case checker.RefLocal:
		var offset, exists = ctx.LocalScope.BindingMap[v.Name]
		if !exists { panic("binding " + v.Name + " does not exist") }
		ctx.LocalScope.Bindings[offset].Used = true
		return CodeSingleInst(InstLocalRef(offset), expr.Info)
	case checker.Block:
		var buf = MakeCodeBuffer()
		for _, b := range v.Bindings {
			switch p := b.Pattern.Concrete.(type) {
			case checker.TrivialPattern:
				var offset uint
				var val_code Code
				if b.Recursive {
					offset = ctx.LocalScope.AddBinding(p.ValueName)
					var lambda, ok = b.Value.Value.(checker.Lambda)
					if !ok { panic("something went wrong") }
					var info = b.Value.Info
					var name = p.ValueName
					val_code = CompileClosure(lambda, info, true, name, ctx)
				} else {
					val_code = Compile(b.Value, ctx)
					offset = ctx.LocalScope.AddBinding(p.ValueName)
				}
				var bind_inst = InstAddBinding(offset)
				buf.Write(val_code)
				buf.Write(CodeSingleInst(bind_inst, b.Value.Info))
			case checker.TuplePattern:
				var val_code = Compile(b.Value, ctx)
				buf.Write(val_code)
				var info = b.Value.Info
				BindPatternItems(p.Items, ctx.LocalScope, buf, info)
			case checker.BundlePattern:
				var val_code = Compile(b.Value, ctx)
				buf.Write(val_code)
				var info = b.Value.Info
				BindPatternItems(p.Items, ctx.LocalScope, buf, info)
			default:
				panic("impossible branch")
			}
		}
		var ret_code = Compile(v.Returned, ctx)
		buf.Write(ret_code)
		return buf.Collect()
	default:
		panic("unknown expression kind")
	}
}

func CompileClosure (
	lambda      checker.Lambda,
	info        checker.ExprInfo,
	recursive   bool,
	rec_name    string,
	ctx         Context,
) Code {
	var inner_ctx = ctx.MakeClosure()
	var inner_scope = inner_ctx.LocalScope
	var inner_buf = MakeCodeBuffer()
	switch p := lambda.Input.Concrete.(type) {
	case checker.TrivialPattern:
		var offset = inner_scope.AddBinding(p.ValueName)
		var bind_inst = InstAddBinding(offset)
		inner_buf.Write(CodeSingleInst(bind_inst, info))
	case checker.TuplePattern:
		BindPatternItems(p.Items, inner_scope, inner_buf, info)
	case checker.BundlePattern:
		BindPatternItems(p.Items, inner_scope, inner_buf, info)
	default:
		panic("impossible branch")
	}
	var body_code = Compile(lambda.Output, inner_ctx)
	inner_buf.Write(body_code)
	var base_reserved_size = *inner_scope.BindingPeek
	if base_reserved_size >= common.LocalSlotMaxSize {
		panic("maximum quantity of local bindings exceeded")
	}
	var outer_bindings_size = uint(len(ctx.LocalScope.Bindings))
	var base_context_size = uint(0)
	var context_offset_map = make(map[uint] uint)
	var context_outer_offsets = make([] uint, 0)
	var rec_used bool
	var rec_outer_offset uint
	for i := uint(0); i < outer_bindings_size; i += 1 {
		var b = inner_scope.Bindings[i]
		if b.Used {
			if recursive && b.Name == rec_name {
				rec_used = true
				rec_outer_offset = i
				continue
			}
			var offset = base_context_size
			base_context_size += 1
			context_offset_map[i] = offset
			context_outer_offsets = append(context_outer_offsets, i)
		}
	}
	if recursive {
		base_context_size += 1
		if rec_used {
			context_offset_map[rec_outer_offset] = base_context_size
		}
	}
	if base_context_size >= common.ClosureMaxSize {
		panic("maximum closure size exceeded")
	}
	var raw_inner_code = inner_buf.Collect()
	var inst_seq_len = len(raw_inner_code.InstSeq)
	var final_inst_seq = make([] common.Instruction, inst_seq_len)
	for i, inst := range raw_inner_code.InstSeq {
		if inst.OpCode == common.LOAD {
			var offset = inst.GetOffset()
			var new_offset uint
			if offset < outer_bindings_size {
				new_offset = context_offset_map[offset]
			} else {
				new_offset = offset - outer_bindings_size + base_context_size
			}
			final_inst_seq[i] = common.Instruction {
				OpCode: common.LOAD,
				Arg0:   0,
				Arg1:   common.Long(new_offset),
			}
		} else {
			final_inst_seq[i] = inst
		}
	}
	var final_inner_code = Code {
		InstSeq:   final_inst_seq,
		SourceMap: raw_inner_code.SourceMap,
	}
	var f = &common.Function {
		IsNative: false,
		Code:     final_inner_code.InstSeq,
		BaseSize:    common.FrameBaseSize {
			Context:  common.Short(base_context_size),
			Reserved: common.Long(base_reserved_size),
		},
		Info:        common.FuncInfo {
			Name:      loader.NewSymbol("", "(closure)"),
			DeclPoint: info.ErrorPoint,
			SourceMap: final_inner_code.SourceMap,
		},
	}
	var index = ctx.AppendClosureRef(f)
	var outer_buf = MakeCodeBuffer()
	outer_buf.Write(CodeSingleInst(InstGlobalRef(index), info))
	for _, outer_offset := range context_outer_offsets {
		var capture_inst = InstLocalRef(outer_offset)
		outer_buf.Write(CodeSingleInst(capture_inst, info))
	}
	var prod_inst = InstProduct(uint(len(context_outer_offsets)))
	outer_buf.Write(CodeSingleInst(prod_inst, info))
	var rec_flag common.Short
	if recursive {
		rec_flag = 1
	} else {
		rec_flag = 0
	}
	var ctx_inst = common.Instruction {
		OpCode: common.CTX,
		Arg0:   rec_flag,
		Arg1:   0,
	}
	outer_buf.Write(CodeSingleInst(ctx_inst, info))
	return outer_buf.Collect()
}

// TODO: CompileFunction

func BindPatternItems (
	items  [] checker.PatternItem,
	scope  *Scope,
	buf    CodeBuffer,
	info   checker.ExprInfo,
) {
	for _, item := range items {
		var inst_get = InstGet(item.Index)
		var offset = scope.AddBinding(item.Name)
		var inst_bind = InstAddBinding(offset)
		buf.Write(CodeSingleInst(inst_get, info))
		buf.Write(CodeSingleInst(inst_bind, info))
	}
}

func InstGlobalRef(index uint) common.Instruction {
	ValidateGlobalIndex(index)
	var a0, a1 = common.RegIndex(index)
	return common.Instruction {
		OpCode: common.GLOBAL,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstLocalRef(offset uint) common.Instruction {
	ValidateLocalOffset(offset)
	return common.Instruction {
		OpCode: common.LOAD,
		Arg0:   0,
		Arg1:   common.Long(offset),
	}
}

func InstAddBinding(offset uint) common.Instruction {
	ValidateLocalOffset(offset)
	return common.Instruction {
		OpCode: common.STORE,
		Arg0:   0,
		Arg1:   common.Long(offset),
	}
}

func InstGet(index uint) common.Instruction {
	ValidateProductIndex(index)
	return common.Instruction {
		OpCode: common.GET,
		Arg0:   common.Short(index),
		Arg1:   0,
	}
}

func InstProduct(size uint) common.Instruction {
	ValidateProductSize(size)
	return common.Instruction{
		OpCode: common.PROD,
		Arg0:   common.Short(size),
		Arg1:   0,
	}
}

func ValidateGlobalIndex(index uint) {
	if index >= common.GlobalSlotMaxSize {
		panic("global value index exceeded maximum slot capacity")
	}
}

func ValidateLocalOffset(offset uint) {
	if offset >= common.LocalSlotMaxSize {
		panic("local binding offset exceeded maximum slot capacity")
	}
}

func ValidateDestAddr(addr uint) {
	if addr >= common.FunCodeMaxLength {
		panic("destination address exceeded limitation")
	}
}

func ValidateProductIndex(index uint) {
	if index >= common.ProductMaxSize {
		panic("value index exceeded maximum capacity of product type")
	}
}

func ValidateProductSize(size uint) {
	if size > common.ProductMaxSize {
		panic("given size exceeded maximum capacity of product type")
	}
}
