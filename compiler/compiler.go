package compiler

import (
	"fmt"
	"reflect"
	"strings"
	"encoding/base64"
	"kumachan/runtime/lib"
	"kumachan/transformer/ast"
	. "kumachan/error"
	c "kumachan/runtime/common"
	ch "kumachan/checker"
)

type Code struct {
	InstSeq    [] c.Instruction
	SourceMap  [] *ast.Node
}
func CodeFrom(inst c.Instruction, info ch.ExprInfo) Code {
	return Code {
		InstSeq:   [] c.Instruction { inst },
		SourceMap: [] *ast.Node { &(info.ErrorPoint.Node) },
	}
}
func (code Code) Length() uint {
	return uint(len(code.InstSeq))
}

type CodeBuffer struct {
	Code  *Code
}
func MakeCodeBuffer() CodeBuffer {
	var code = &Code {
		InstSeq:   make([] c.Instruction, 0),
		SourceMap: make([] *ast.Node, 0),
	}
	return CodeBuffer { code }
}
func (buf CodeBuffer) Write(code Code) {
	var base = &(buf.Code.InstSeq)
	var base_size = uint(len(buf.Code.InstSeq))
	for _, inst := range code.InstSeq {
		if inst.OpCode == c.JIF || inst.OpCode == c.JMP {
			var dest_addr = (uint(inst.Arg1) + base_size)
			ValidateDestAddr(dest_addr)
			*base = append(*base, c.Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   c.Long(dest_addr),
			})
		} else {
			*base = append(*base, inst)
		}
	}
	buf.Code.SourceMap = append(buf.Code.SourceMap, code.SourceMap...)
}
func (buf CodeBuffer) WriteAbsolute(code Code) {
	buf.Code.InstSeq = append(buf.Code.InstSeq, code.InstSeq...)
	buf.Code.SourceMap = append(buf.Code.SourceMap, code.SourceMap...)
}
func (buf CodeBuffer) WriteBranch(code Code, tail_addr uint) {
	var base = &(buf.Code.InstSeq)
	var base_size = uint(len(buf.Code.InstSeq))
	var last_addr = (code.Length() - 1)
	for _, inst := range code.InstSeq {
		if inst.OpCode == c.JIF || inst.OpCode == c.JMP {
			var dest_addr = (uint(inst.Arg1) + base_size)
			ValidateDestAddr(dest_addr)
			if dest_addr == last_addr {
				dest_addr = tail_addr
			}
			*base = append(*base, c.Instruction {
				OpCode: inst.OpCode,
				Arg0:   inst.Arg0,
				Arg1:   c.Long(dest_addr),
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
func MakeContext() Context {
	var refs = make([] GlobalRef, 0)
	return Context {
		GlobalRefs: &refs,
		LocalScope: MakeScope(),
	}
}
func (ctx Context) MakeClosure() Context {
	var refs = make([] GlobalRef, 0)
	return Context {
		GlobalRefs: &refs,
		LocalScope: MakeClosureScope(ctx.LocalScope),
	}
}
func (ctx Context) MakeBranch() Context {
	return Context {
		GlobalRefs: ctx.GlobalRefs,
		LocalScope: MakeBranchScope(ctx.LocalScope),
	}
}
func (ctx Context) AppendDataRef(v c.DataValue) uint {
	var refs = ctx.GlobalRefs
	var index = uint(len(*refs))
	*refs = append(*refs, RefData { v })
	return index
}
func (ctx Context) AppendFunRef(ref ch.RefFunction) uint {
	var refs = ctx.GlobalRefs
	var index = uint(len(*refs))
	*refs = append(*refs, RefFun(ref.AbsRef))
	return index
}
func (ctx Context) AppendConstRef(ref ch.RefConstant) uint {
	var refs = ctx.GlobalRefs
	var index = uint(len(*refs))
	*refs = append(*refs, RefConst(ref))
	return index
}
func (ctx Context) AppendClosureRef(f *c.Function, refs []GlobalRef) uint {
	var outer_refs = ctx.GlobalRefs
	var index = uint(len(*outer_refs))
	*outer_refs = append(*outer_refs, RefClosure {
		Function:   f,
		GlobalRefs: refs,
	})
	return index
}

type GlobalRef interface { GlobalRef() }
func (impl RefData) GlobalRef() {}
type RefData struct { c.DataValue }
func (impl RefFun) GlobalRef() {}
type RefFun ch.AbsRefFunction
func (impl RefConst) GlobalRef() {}
type RefConst ch.RefConstant
func (impl RefClosure) GlobalRef() {}
type RefClosure struct {
	Function    *c.Function
	GlobalRefs  [] GlobalRef
}

type DataInteger ch.IntLiteral
func (d DataInteger) ToValue() c.Value {
	return d.Value
}
func (d DataInteger) String() string {
	return fmt.Sprintf("BIG %s", d.Value.String())
}
type DataSmallInteger ch.SmallIntLiteral
func (d DataSmallInteger) ToValue() c.Value {
	return d.Value
}
func (d DataSmallInteger) String() string {
	return fmt.Sprintf("SMALL %s %v", reflect.TypeOf(d.Value).String(), d.Value)
}
type DataFloat ch.FloatLiteral
func (d DataFloat) ToValue() c.Value {
	return d.Value
}
func (d DataFloat) String() string {
	return fmt.Sprintf("FLOAT %f", d.Value)
}
type DataString struct { Value  [] rune }
func (d DataString) ToValue() c.Value {
	return d.Value
}
func (d DataString) String() string {
	var b64 = RunesToBase64String(d.Value)
	return fmt.Sprintf("STRING %s", b64)
}
type DataStringFormatter ch.StringFormatter
func (d DataStringFormatter) ToValue() c.Value {
	var f = func(args []c.Value) []rune {
		var buf = make([]rune, 0)
		for i, seg := range d.Segments {
			buf = append(buf, seg...)
			if uint(i) < d.Arity {
				var runes = args[i].([]rune)
				buf = append(buf, runes...)
			}
		}
		return buf
	}
	return c.NativeFunctionValue(c.AdaptNativeFunction(f))
}
func (d DataStringFormatter) String() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "FORMAT %d ", d.Arity)
	for i, item := range d.Segments {
		buf.WriteString(RunesToBase64String(item))
		if i != len(d.Segments)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

func RunesToBase64String(runes []rune) string {
	var buf strings.Builder
	var encoder = base64.NewEncoder(base64.StdEncoding, &buf)
	var data = []byte(string(runes))
	var n, err = encoder.Write(data)
	if n != len(data) { panic("something went wrong") }
	if err != nil { panic(err) }
	_ = encoder.Close()
	return buf.String()
}

type Scope struct {
	Bindings     [] Binding
	BindingMap   map[string] uint
	BindingPeek  *uint
	Children     [] *Scope
	NextId       *uint
}
type Binding struct {
	Name   string
	Used   bool
	Point  ErrorPoint
	Id     uint
}
func MakeScope() *Scope {
	return &Scope {
		Bindings:    make([] Binding, 0),
		BindingMap:  make(map[string] uint),
		BindingPeek: new(uint),
		Children:    make([] *Scope, 0),
		NextId:      new(uint),
	}
}
func MakeClosureScope(outer *Scope) *Scope {
	var bindings = make([] Binding, 0)
	for i, e := range outer.Bindings {
		bindings[i] = Binding {
			Name:  e.Name,
			Used:  false,
			Point: e.Point,
		}
	}
	var binding_map = make(map[string] uint)
	for k, v := range outer.BindingMap {
		binding_map[k] = v
	}
	var child = &Scope {
		Bindings:    bindings,
		BindingMap:  binding_map,
		BindingPeek: new(uint),
		Children:    make([] *Scope, 0),
		NextId:      outer.NextId,
	}
	outer.Children = append(outer.Children, child)
	return child
}
func MakeBranchScope(outer *Scope) *Scope {
	var bindings = make([] Binding, 0)
	copy(bindings, outer.Bindings)
	var binding_map = make(map[string] uint)
	for k, v := range outer.BindingMap {
		binding_map[k] = v
	}
	var child = &Scope {
		Bindings:    bindings,
		BindingMap:  binding_map,
		BindingPeek: outer.BindingPeek,
		Children:    make([] *Scope, 0),
		NextId:      outer.NextId,
	}
	outer.Children = append(outer.Children, child)
	return child
}
func (scope *Scope) AddBinding(name string, point ErrorPoint) uint {
	var _, exists = scope.BindingMap[name]
	if exists {
		// shadowing: do nothing
	}
	var list = &(scope.Bindings)
	var offset = uint(len(*list))
	*list = append(*list, Binding {
		Name:  name,
		Used:  false,
		Point: point,
		Id:    *(scope.NextId),
	})
	*(scope.NextId) += 1
	scope.BindingMap[name] = offset
	*(scope.BindingPeek) += 1
	return offset
}
func (scope *Scope) CollectUnused() ([] Binding) {
	var all = make(map[uint] *Binding)
	var collect func(*Scope)
	collect = func(scope *Scope) {
		for _, b := range scope.Bindings {
			var existing, exists = all[b.Id]
			if exists {
				existing.Used = (existing.Used || b.Used)
			} else {
				var b_copied = b  // make copy more clearly
				all[b.Id] = &b_copied
			}
		}
		for _, child := range scope.Children {
			collect(child)
		}
	}
	collect(scope)
	var unused = make([] Binding, 0)
	for _, b := range all {
		if !(b.Used) {
			unused = append(unused, *b)
		}
	}
	return unused
}
func (scope *Scope) CollectUnusedAsErrors() ([] *Error) {
	var unused = scope.CollectUnused()
	if len(unused) == 0 {
		return nil
	} else {
		var errs = make([] *Error, 0)
		for _, b := range unused {
			if b.Name == ch.IgnoreMark {
				continue
			}
			errs = append(errs, &Error {
				Point:    b.Point,
				Concrete: E_UnusedBinding { b.Name },
			})
		}
		return errs
	}
}

func Compile(expr ch.Expr, ctx Context) Code {
	switch v := expr.Value.(type) {
	case ch.UnitValue:
		var inst_nil = c.Instruction { OpCode: c.NIL }
		return CodeFrom(inst_nil, expr.Info)
	case ch.IntLiteral:
		var index = ctx.AppendDataRef(DataInteger(v))
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.SmallIntLiteral:
		var index = ctx.AppendDataRef(DataSmallInteger(v))
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.FloatLiteral:
		var index = ctx.AppendDataRef(DataFloat(v))
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.StringLiteral:
		var index = ctx.AppendDataRef(DataString { v.Value })
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.StringFormatter:
		var index = ctx.AppendDataRef(DataStringFormatter(v))
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.RefFunction:
		var index = ctx.AppendFunRef(v)
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.RefConstant:
		var index = ctx.AppendConstRef(v)
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.RefLocal:
		var offset, exists = ctx.LocalScope.BindingMap[v.Name]
		if !exists { panic("binding " + v.Name + " does not exist") }
		ctx.LocalScope.Bindings[offset].Used = true
		return CodeFrom(InstLocalRef(offset), expr.Info)
	case ch.Array:
		var buf = MakeCodeBuffer()
		var inst_array = InstArray(uint(len(v.Items)))
		buf.Write(CodeFrom(inst_array, expr.Info))
		for _, item := range v.Items {
			var item_code = Compile(item, ctx)
			buf.Write(item_code)
			var inst_append = c.Instruction {
				OpCode: c.APPEND,
			}
			buf.Write(CodeFrom(inst_append, item.Info))
		}
		return buf.Collect()
	case ch.Product:
		var buf = MakeCodeBuffer()
		for _, element := range v.Values {
			var element_code = Compile(element, ctx)
			buf.Write(element_code)
		}
		var inst_prod = InstProduct(uint(len(v.Values)))
		buf.Write(CodeFrom(inst_prod, expr.Info))
		return buf.Collect()
	case ch.Get:
		var buf = MakeCodeBuffer()
		var base_code = Compile(v.Product, ctx)
		buf.Write(base_code)
		var inst_get = InstGet(v.Index)
		buf.Write(CodeFrom(inst_get, expr.Info))
		return buf.Collect()
	case ch.Set:
		var buf = MakeCodeBuffer()
		var base_code = Compile(v.Product, ctx)
		buf.Write(base_code)
		var new_value_code = Compile(v.NewValue, ctx)
		buf.Write(new_value_code)
		var inst_set = InstSet(v.Index)
		buf.Write(CodeFrom(inst_set, expr.Info))
		return buf.Collect()
	case ch.Sum:
		var buf = MakeCodeBuffer()
		var concrete = Compile(v.Value, ctx)
		buf.Write(concrete)
		var inst_sum = InstSum(v.Index)
		buf.Write(CodeFrom(inst_sum, expr.Info))
		return buf.Collect()
	case ch.Match:
		var raw_branches = make([]ch.Branch, len(v.Branches))
		var i = 0
		var default_occurred = false
		for _, b := range v.Branches {
			if b.IsDefault {
				if default_occurred { panic("something went wrong") }
				raw_branches[len(raw_branches)-1] = b
				default_occurred = true
			} else {
				raw_branches[i] = b
				i += 1
			}
		}
		var branches = make([]Code, len(raw_branches))
		for i, b := range raw_branches {
			var branch_buf = MakeCodeBuffer()
			var branch_ctx = ctx.MakeBranch()
			var branch_scope = branch_ctx.LocalScope
			var pattern, ok = b.Pattern.(ch.Pattern)
			if ok {
				switch p := pattern.Concrete.(type) {
				case ch.TrivialPattern:
					var offset = branch_scope.AddBinding (
						p.ValueName, pattern.Point,
					)
					var bind_inst = InstAddBinding(offset)
					branch_buf.Write(CodeFrom(bind_inst, b.Value.Info))
				case ch.TuplePattern:
					BindPatternItems (
						pattern,       p.Items,
						branch_scope,  branch_buf,
					)
				case ch.BundlePattern:
					BindPatternItems (
						pattern,       p.Items,
						branch_scope,  branch_buf,
					)
				}
			}
			var expr_code = Compile(b.Value, branch_ctx)
			branch_buf.Write(expr_code)
			branches[i] = branch_buf.Collect()
		}
		var arg_code = Compile(v.Argument, ctx)
		var branch_count = uint(len(branches))
		var addrs = make([]uint, branch_count)
		var addr = arg_code.Length() + uint(branch_count)
		for i := uint(0); i < branch_count; i += 1 {
			addrs[i] = addr
			addr += (branches[i].Length() + 1)
		}
		var tail_addr = addr
		var buf = MakeCodeBuffer()
		buf.Write(arg_code)
		for i := uint(0); i < branch_count; i += 1 {
			var index = raw_branches[i].Index
			var jump c.Instruction
			if raw_branches[i].IsDefault {
				jump = InstJump(addrs[i], true)
			} else {
				jump = InstJumpIf(index, addrs[i])
			}
			buf.WriteAbsolute(CodeFrom(jump, v.Argument.Info))
		}
		for i, branch_code := range branches {
			buf.WriteBranch(branch_code, tail_addr)
			var goto_tail = InstJump(tail_addr, false)
			var info = raw_branches[i].Value.Info
			buf.WriteAbsolute(CodeFrom(goto_tail, info))
		}
		var nop = c.Instruction { OpCode: c.NOP }
		buf.Write(CodeFrom(nop, v.Argument.Info))
		return buf.Collect()
	case ch.Lambda:
		return CompileClosure(v, expr.Info, false, "", ctx)
	case ch.Block:
		var buf = MakeCodeBuffer()
		for _, b := range v.Bindings {
			switch p := b.Pattern.Concrete.(type) {
			case ch.TrivialPattern:
				var offset uint
				var val_code Code
				if b.Recursive {
					offset = ctx.LocalScope.AddBinding(p.ValueName, p.Point)
					var lambda, ok = b.Value.Value.(ch.Lambda)
					if !ok { panic("something went wrong") }
					var info = b.Value.Info
					var name = p.ValueName
					val_code = CompileClosure(lambda, info, true, name, ctx)
				} else {
					val_code = Compile(b.Value, ctx)
					offset = ctx.LocalScope.AddBinding(p.ValueName, p.Point)
				}
				var bind_inst = InstAddBinding(offset)
				buf.Write(val_code)
				buf.Write(CodeFrom(bind_inst, b.Value.Info))
			case ch.TuplePattern:
				var val_code = Compile(b.Value, ctx)
				buf.Write(val_code)
				BindPatternItems (
					b.Pattern,       p.Items,
					ctx.LocalScope,  buf,
				)
			case ch.BundlePattern:
				var val_code = Compile(b.Value, ctx)
				buf.Write(val_code)
				BindPatternItems (
					b.Pattern,       p.Items,
					ctx.LocalScope,  buf,
				)
			default:
				panic("impossible branch")
			}
		}
		var ret_code = Compile(v.Returned, ctx)
		buf.Write(ret_code)
		return buf.Collect()
	case ch.Call:
		var buf = MakeCodeBuffer()
		var arg_code = Compile(v.Argument, ctx)
		var f_code = Compile(v.Function, ctx)
		buf.Write(arg_code)
		buf.Write(f_code)
		var inst_call = c.Instruction {
			OpCode: c.CALL,
		}
		buf.Write(CodeFrom(inst_call, expr.Info))
		return buf.Collect()
	default:
		panic("unknown expression kind")
	}
}

func CompileClosure (
	lambda      ch.Lambda,
	info        ch.ExprInfo,
	recursive   bool,
	rec_name    string,
	ctx         Context,
) Code {
	var inner_ctx = ctx.MakeClosure()
	var inner_scope = inner_ctx.LocalScope
	var inner_buf = MakeCodeBuffer()
	var pattern = lambda.Input
	switch p := pattern.Concrete.(type) {
	case ch.TrivialPattern:
		var offset = inner_scope.AddBinding(p.ValueName, p.Point)
		var bind_inst = InstAddBinding(offset)
		inner_buf.Write(CodeFrom(bind_inst, info))
	case ch.TuplePattern:
		BindPatternItems(pattern, p.Items, inner_scope, inner_buf)
	case ch.BundlePattern:
		BindPatternItems(pattern, p.Items, inner_scope, inner_buf)
	default:
		panic("impossible branch")
	}
	var body_code = Compile(lambda.Output, inner_ctx)
	inner_buf.Write(body_code)
	var base_reserved_size = *(inner_scope.BindingPeek)
	if base_reserved_size >= c.LocalSlotMaxSize {
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
	if base_context_size >= c.ClosureMaxSize {
		panic("maximum closure size exceeded")
	}
	var raw_inner_code = inner_buf.Collect()
	var inst_seq_len = len(raw_inner_code.InstSeq)
	var final_inst_seq = make([] c.Instruction, inst_seq_len)
	for i, inst := range raw_inner_code.InstSeq {
		if inst.OpCode == c.LOAD {
			var offset = inst.GetOffset()
			var new_offset uint
			if offset < outer_bindings_size {
				new_offset = context_offset_map[offset]
			} else {
				new_offset = offset - outer_bindings_size + base_context_size
			}
			final_inst_seq[i] = c.Instruction {
				OpCode: c.LOAD,
				Arg0:   0,
				Arg1:   c.Long(new_offset),
			}
		} else {
			final_inst_seq[i] = inst
		}
	}
	var final_inner_code = Code {
		InstSeq:   final_inst_seq,
		SourceMap: raw_inner_code.SourceMap,
	}
	var f = &c.Function {
		IsNative: false,
		Code:     final_inner_code.InstSeq,
		BaseSize:    c.FrameBaseSize {
			Context:  c.Short(base_context_size),
			Reserved: c.Long(base_reserved_size),
		},
		Info:        c.FuncInfo {
			Name:      "(closure)",
			DeclPoint: info.ErrorPoint,
			SourceMap: final_inner_code.SourceMap,
		},
	}
	var index = ctx.AppendClosureRef(f, *(inner_ctx.GlobalRefs))
	var outer_buf = MakeCodeBuffer()
	outer_buf.Write(CodeFrom(InstGlobalRef(index), info))
	for _, outer_offset := range context_outer_offsets {
		var capture_inst = InstLocalRef(outer_offset)
		outer_buf.Write(CodeFrom(capture_inst, info))
	}
	var prod_inst = InstProduct(uint(len(context_outer_offsets)))
	outer_buf.Write(CodeFrom(prod_inst, info))
	var rec_flag c.Short
	if recursive {
		rec_flag = 1
	} else {
		rec_flag = 0
	}
	var ctx_inst = c.Instruction {
		OpCode: c.CTX,
		Arg0:   rec_flag,
		Arg1:   0,
	}
	outer_buf.Write(CodeFrom(ctx_inst, info))
	return outer_buf.Collect()
}

func CompileFunction (
	body   ch.ExprLike,
	mod    string,
	name   string,
	point  ErrorPoint,
) (*c.Function, []GlobalRef, []*Error) {
	switch b := body.(type) {
	case ch.ExprNative:
		var native_name = b.Name
		var index, exists = lib.NativeFunctionIndex[native_name]
		var errs []*Error = nil
		if !exists {
			errs = []*Error { &Error {
				Point:    b.Point,
				Concrete: E_NativeFunctionNotFound { native_name },
			} }
		}
		return &c.Function {
			IsNative:    true,
			NativeIndex: index,
			Code:        nil,
			BaseSize:    c.FrameBaseSize {},
			Info:        c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([]GlobalRef, 0), errs
	case ch.ExprExpr:
		var body_expr = ch.Expr(b)
		var lambda = body_expr.Value.(ch.Lambda)
		var pattern = lambda.Input
		var ctx = MakeContext()
		var scope = ctx.LocalScope
		var buf = MakeCodeBuffer()
		var info = body_expr.Info
		switch p := pattern.Concrete.(type) {
		case ch.TrivialPattern:
			var offset = scope.AddBinding(p.ValueName, p.Point)
			var bind_inst = InstAddBinding(offset)
			buf.Write(CodeFrom(bind_inst, info))
		case ch.TuplePattern:
			BindPatternItems(pattern, p.Items, scope, buf)
		case ch.BundlePattern:
			BindPatternItems(pattern, p.Items, scope, buf)
		default:
			panic("impossible branch")
		}
		var out_code = Compile(lambda.Output, ctx)
		var errs = ctx.LocalScope.CollectUnusedAsErrors()
		buf.Write(out_code)
		var code = buf.Collect()
		var binding_peek = *(scope.BindingPeek)
		if binding_peek >= c.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &c.Function {
			IsNative:    false,
			NativeIndex: -1,
			Code:        code.InstSeq,
			BaseSize:    c.FrameBaseSize {
				Context:  0,
				Reserved: c.Long(binding_peek),
			},
			Info:        c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: code.SourceMap,
			},
		}, *(ctx.GlobalRefs), errs
	default:
		panic("impossible branch")
	}
}

func CompileConstant (
	body   ch.ExprLike,
	mod    string,
	name   string,
	point  ErrorPoint,
) (*c.Function, []GlobalRef, []*Error) {
	switch b := body.(type) {
	case ch.ExprNative:
		var native_name = b.Name
		var index, exists = lib.NativeConstantIndex[native_name]
		var errs []*Error = nil
		if !exists {
			errs = []*Error { &Error {
				Point:    b.Point,
				Concrete: E_NativeConstantNotFound { native_name },
			} }
		}
		return &c.Function {
			IsNative:    true,
			NativeIndex: index,
			Code:        nil,
			BaseSize:    c.FrameBaseSize {},
			Info: c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: nil,
			},
		}, make([] GlobalRef, 0), errs
	case ch.ExprExpr:
		var body_expr = ch.Expr(b)
		var ctx = MakeContext()
		var code = Compile(body_expr, ctx)
		var errs = ctx.LocalScope.CollectUnusedAsErrors()
		var binding_peek = *(ctx.LocalScope.BindingPeek)
		if binding_peek >= c.LocalSlotMaxSize {
			panic("maximum quantity of local bindings exceeded")
		}
		return &c.Function {
			IsNative:    false,
			NativeIndex: -1,
			Code:        code.InstSeq,
			BaseSize:    c.FrameBaseSize {
				Context:  0,
				Reserved: c.Long(binding_peek),
			},
			Info:        c.FuncInfo {
				Module:    mod,
				Name:      name,
				DeclPoint: point,
				SourceMap: code.SourceMap,
			},
		}, *(ctx.GlobalRefs), errs
	default:
		panic("impossible branch")
	}
}

func BindPatternItems (
	pattern  ch.Pattern,
	items    [] ch.PatternItem,
	scope    *Scope,
	buf      CodeBuffer,
) {
	var info = ch.ExprInfo {
		ErrorPoint: pattern.Point,
	}
	for _, item := range items {
		var inst_get = InstGet(item.Index)
		var offset = scope.AddBinding(item.Name, item.Point)
		var inst_bind = InstAddBinding(offset)
		buf.Write(CodeFrom(inst_get, info))
		buf.Write(CodeFrom(inst_bind, info))
	}
}

func InstGlobalRef(index uint) c.Instruction {
	ValidateGlobalIndex(index)
	var a0, a1 = c.GlobalIndex(index)
	return c.Instruction {
		OpCode: c.GLOBAL,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstLocalRef(offset uint) c.Instruction {
	ValidateLocalOffset(offset)
	return c.Instruction {
		OpCode: c.LOAD,
		Arg0:   0,
		Arg1:   c.Long(offset),
	}
}

func InstAddBinding(offset uint) c.Instruction {
	ValidateLocalOffset(offset)
	return c.Instruction {
		OpCode: c.STORE,
		Arg0:   0,
		Arg1:   c.Long(offset),
	}
}

func InstGet(index uint) c.Instruction {
	ValidateProductIndex(index)
	return c.Instruction {
		OpCode: c.GET,
		Arg0:   c.Short(index),
		Arg1:   0,
	}
}

func InstSet(index uint) c.Instruction {
	ValidateProductIndex(index)
	return c.Instruction {
		OpCode: c.SET,
		Arg0:   c.Short(index),
		Arg1:   0,
	}
}

func InstProduct(size uint) c.Instruction {
	ValidateProductSize(size)
	return c.Instruction {
		OpCode: c.PROD,
		Arg0:   c.Short(size),
		Arg1:   0,
	}
}

func InstArray(size uint) c.Instruction {
	ValidateArraySize(size)
	var a0, a1 = c.ArraySize(size)
	return c.Instruction {
		OpCode: c.ARRAY,
		Arg0:   a0,
		Arg1:   a1,
	}
}

func InstSum(index uint) c.Instruction {
	ValidateSumIndex(index)
	return c.Instruction {
		OpCode: c.SUM,
		Arg0:   c.Short(index),
		Arg1:   0,
	}
}

func InstJumpIf(index uint, dest uint) c.Instruction {
	ValidateSumIndex(index)
	ValidateDestAddr(dest)
	return c.Instruction {
		OpCode: c.JIF,
		Arg0:   c.Short(index),
		Arg1:   c.Long(dest),
	}
}

func InstJump(dest uint, narrow bool) c.Instruction {
	ValidateDestAddr(dest)
	var narrow_flag uint
	if narrow {
		narrow_flag = 1
	} else {
		narrow_flag = 0
	}
	return c.Instruction {
		OpCode: c.JMP,
		Arg0:   c.Short(narrow_flag),
		Arg1:   c.Long(dest),
	}
}

func ValidateGlobalIndex(index uint) {
	if index >= c.GlobalSlotMaxSize {
		panic("global value index exceeded maximum slot capacity")
	}
}

func ValidateLocalOffset(offset uint) {
	if offset >= c.LocalSlotMaxSize {
		panic("local binding offset exceeded maximum slot capacity")
	}
}

func ValidateDestAddr(addr uint) {
	if addr >= c.FunCodeMaxLength {
		panic("destination address exceeded limitation")
	}
}

func ValidateProductIndex(index uint) {
	if index >= c.ProductMaxSize {
		panic("value index exceeded maximum capacity of product type")
	}
}

func ValidateProductSize(size uint) {
	if size > c.ProductMaxSize {
		panic("given size exceeded maximum capacity of product type")
	}
}

func ValidateSumIndex(index uint) {
	if index >= c.SumMaxBranches {
		panic("given index exceeded maximum branch limit of sum type")
	}
}

func ValidateArraySize(size uint) {
	if size > c.ArrayMaxSize {
		panic("given size exceeded maximum capacity of array literal")
	}
}
