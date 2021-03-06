package generator

import (
	ch "kumachan/interpreter/compiler/checker"
	"kumachan/interpreter/def"
)


type Context struct {
	GlobalRefs  *([] GlobalRef)
	LocalScope  *Scope
}

func MakeContextWithImplicit(fields ([] string)) Context {
	var refs = make([] GlobalRef, 0)
	return Context {
		GlobalRefs: &refs,
		LocalScope: MakeScopeWithImplicit(fields),
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

func (ctx Context) AppendDataRef(v def.DataValue) uint {
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

func (ctx Context) AppendClosureRef(f *def.Function, refs []GlobalRef) uint {
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
type RefData struct { def.DataValue }

func (impl RefFun) GlobalRef() {}
type RefFun ch.AbsRefFunction

func (impl RefConst) GlobalRef() {}
type RefConst ch.RefConstant

func (impl RefClosure) GlobalRef() {}
type RefClosure struct {
	Function    *def.Function
	GlobalRefs  [] GlobalRef
}


func CompileExpr(expr ch.Expr, ctx Context) Code {
	switch v := expr.Value.(type) {
	case ch.UnitValue:
		var inst_nil = def.Instruction { OpCode: def.NIL }
		return CodeFrom(inst_nil, expr.Info)
	case ch.IntegerLiteral:
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
		var buf = MakeCodeBuffer()
		buf.Write(CodeFrom(InstGlobalRef(index), expr.Info))
		if len(v.Implicit) > 0 {
			var n = uint(len(v.Implicit))
			for _, ref := range v.Implicit {
				var ref_expr = ch.Expr {
					Type:  nil,
					Value: ref,
					Info:  expr.Info,
				}
				var ref_code = CompileExpr(ref_expr, ctx)
				buf.Write(ref_code)
			}
			buf.Write(CodeFrom(InstProduct(n), expr.Info))
			buf.Write(CodeFrom(def.Instruction { OpCode: def.CTX }, expr.Info))
		}
		return buf.Collect()
	case ch.RefConstant:
		var index = ctx.AppendConstRef(v)
		return CodeFrom(InstGlobalRef(index), expr.Info)
	case ch.RefLocal:
		var offset, exists = ctx.LocalScope.BindingMap[v.Name]
		if !exists { panic("binding " + v.Name + " does not exist") }
		ctx.LocalScope.Bindings[offset].Used = true
		return CodeFrom(InstLocalRef(offset), expr.Info)
	case ch.Array:
		var length = uint(len(v.Items))
		if length > def.ArrayMaxSize {
			panic("array literal length exceeded limit")
		}
		var info = ch.GetArrayInfo(length, v.ItemType)
		var info_index = ctx.AppendDataRef(DataArrayInfo(info))
		var buf = MakeCodeBuffer()
		var inst_array = InstArray(info_index)
		buf.Write(CodeFrom(inst_array, expr.Info))
		for _, item := range v.Items {
			var item_code = CompileExpr(item, ctx)
			buf.Write(item_code)
			var inst_append = def.Instruction {
				OpCode: def.APPEND,
			}
			buf.Write(CodeFrom(inst_append, item.Info))
		}
		return buf.Collect()
	case ch.Product:
		var buf = MakeCodeBuffer()
		for _, element := range v.Values {
			var element_code = CompileExpr(element, ctx)
			buf.Write(element_code)
		}
		var inst_prod = InstProduct(uint(len(v.Values)))
		buf.Write(CodeFrom(inst_prod, expr.Info))
		return buf.Collect()
	case ch.Get:
		var buf = MakeCodeBuffer()
		var base_code = CompileExpr(v.Product, ctx)
		buf.Write(base_code)
		var inst_get = InstPopGet(v.Index)
		buf.Write(CodeFrom(inst_get, expr.Info))
		return buf.Collect()
	case ch.Set:
		var buf = MakeCodeBuffer()
		var base_code = CompileExpr(v.Product, ctx)
		buf.Write(base_code)
		var new_value_code = CompileExpr(v.NewValue, ctx)
		buf.Write(new_value_code)
		var inst_set = InstSet(v.Index)
		buf.Write(CodeFrom(inst_set, expr.Info))
		return buf.Collect()
	case ch.Reference:
		var buf = MakeCodeBuffer()
		var base_code = CompileExpr(v.Base, ctx)
		buf.Write(base_code)
		var inst_proj = InstRef(v.Index, v.Kind, v.Operand)
		buf.Write(CodeFrom(inst_proj, expr.Info))
		return buf.Collect()
	case ch.Sum:
		var buf = MakeCodeBuffer()
		var concrete = CompileExpr(v.Value, ctx)
		buf.Write(concrete)
		var inst_sum = InstSum(v.Index)
		buf.Write(CodeFrom(inst_sum, expr.Info))
		return buf.Collect()
	case ch.Switch:
		var raw_branches = make([] ch.Branch, len(v.Branches))
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
		var branches = make([] Code, len(raw_branches))
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
					var inst_store = InstStore(offset)
					branch_buf.Write(CodeFrom(inst_store, b.Value.Info))
				case ch.TuplePattern:
					BindPatternItems (
						pattern,       p.Items,
						branch_scope,  branch_buf,
					)
				case ch.RecordPattern:
					BindPatternItems (
						pattern,       p.Items,
						branch_scope,  branch_buf,
					)
				default:
					panic("impossible branch")
				}
			} else {
				var pop_inst = def.Instruction { OpCode: def.POP }
				branch_buf.Write(CodeFrom(pop_inst, v.Argument.Info))
			}
			var expr_code = CompileExpr(b.Value, branch_ctx)
			branch_buf.Write(expr_code)
			branches[i] = branch_buf.Collect()
		}
		var arg_code = CompileExpr(v.Argument, ctx)
		var branch_count = uint(len(branches))
		var addrs = make([]uint, branch_count)
		var addr = arg_code.Length() + branch_count
		for i := uint(0); i < branch_count; i += 1 {
			addrs[i] = addr
			addr += (branches[i].Length() + 1)
		}
		var tail_addr = addr
		var buf = MakeCodeBuffer()
		buf.Write(arg_code)
		for i := uint(0); i < branch_count; i += 1 {
			var index = raw_branches[i].Index
			var jump def.Instruction
			if raw_branches[i].IsDefault {
				jump = InstJump(addrs[i])
			} else {
				jump = InstJumpIf(index, addrs[i])
			}
			buf.WriteAbsolute(CodeFrom(jump, v.Argument.Info))
		}
		for i, branch_code := range branches {
			buf.WriteBranch(branch_code, tail_addr)
			var goto_tail = InstJump(tail_addr)
			var info = raw_branches[i].Value.Info
			buf.WriteAbsolute(CodeFrom(goto_tail, info))
		}
		var nop = def.Instruction { OpCode: def.NOP }
		buf.Write(CodeFrom(nop, v.Argument.Info))
		return buf.Collect()
	case ch.MultiSwitch:
		var arg = ch.GetMultiSwitchArgumentTuple(v, expr.Info)
		var A = uint(len(v.Arguments))
		var raw_branches = make([]ch.MultiBranch, len(v.Branches))
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
				case ch.TuplePattern:
					BindPatternItems (
						pattern,       p.Items,
						branch_scope,  branch_buf,
					)
				default:
					panic("something went wrong")
				}
			} else {
				var pop_inst = def.Instruction { OpCode: def.POP }
				branch_buf.Write(CodeFrom(pop_inst, arg.Info))
			}
			var expr_code = CompileExpr(b.Value, branch_ctx)
			branch_buf.Write(expr_code)
			branches[i] = branch_buf.Collect()
		}
		var arg_code = CompileExpr(arg, ctx)
		var branch_count = uint(len(branches))
		var cond_code_length uint
		if default_occurred {
			cond_code_length = (((branch_count - 1) * (A + 2)) + 1)
		} else {
			cond_code_length = (branch_count * (A + 2))
		}
		var addrs = make([]uint, branch_count)
		var addr = arg_code.Length() + cond_code_length
		for i := uint(0); i < branch_count; i += 1 {
			addrs[i] = addr
			addr += (branches[i].Length() + 1)
		}
		var tail_addr = addr
		var buf = MakeCodeBuffer()
		buf.Write(arg_code)
		for i := uint(0); i < branch_count; i += 1 {
			if raw_branches[i].IsDefault {
				var jump = InstJump(addrs[i])
				buf.WriteAbsolute(CodeFrom(jump, arg.Info))
			} else {
				var element_indexes = raw_branches[i].Indexes
				var ms = def.Instruction { OpCode: def.MS }
				buf.WriteAbsolute(CodeFrom(ms, arg.Info))
				if uint(len(element_indexes)) != A {
					panic("something went wrong")
				}
				for j, el := range element_indexes {
					var el_inst def.Instruction
					if el.IsDefault {
						el_inst = def.Instruction { OpCode: def.MSD }
					} else {
						el_inst = InstMultiSwitchIndex(el.Index)
					}
					buf.WriteAbsolute(CodeFrom(el_inst, v.Arguments[j].Info))
				}
				var jump = InstMultiSwitchJump(addrs[i])
				buf.WriteAbsolute(CodeFrom(jump, arg.Info))
			}
		}
		for i, branch_code := range branches {
			buf.WriteBranch(branch_code, tail_addr)
			var goto_tail = InstJump(tail_addr)
			var info = raw_branches[i].Value.Info
			buf.WriteAbsolute(CodeFrom(goto_tail, info))
		}
		var nop = def.Instruction { OpCode: def.NOP }
		buf.Write(CodeFrom(nop, arg.Info))
		return buf.Collect()
	case ch.Lambda:
		return CompileClosure(v, expr.Info, false, "", ctx)
	case ch.PipelineLambdaArgument:
		return CodeFrom(def.Instruction { OpCode: def.NOP }, expr.Info)
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
					val_code = CompileExpr(b.Value, ctx)
					offset = ctx.LocalScope.AddBinding(p.ValueName, p.Point)
				}
				var inst_store = InstStore(offset)
				buf.Write(val_code)
				buf.Write(CodeFrom(inst_store, b.Value.Info))
			case ch.TuplePattern:
				var val_code = CompileExpr(b.Value, ctx)
				buf.Write(val_code)
				BindPatternItems (
					b.Pattern,       p.Items,
					ctx.LocalScope,  buf,
				)
			case ch.RecordPattern:
				var val_code = CompileExpr(b.Value, ctx)
				buf.Write(val_code)
				BindPatternItems (
					b.Pattern,       p.Items,
					ctx.LocalScope,  buf,
				)
			default:
				panic("impossible branch")
			}
		}
		var ret_code = CompileExpr(v.Returned, ctx)
		buf.Write(ret_code)
		return buf.Collect()
	case ch.Call:
		var buf = MakeCodeBuffer()
		var arg_code = CompileExpr(v.Argument, ctx)
		var f_code = CompileExpr(v.Function, ctx)
		buf.Write(arg_code)
		buf.Write(f_code)
		var inst_call = def.Instruction {
			OpCode: def.CALL,
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
	case ch.NullPattern:
		// do nothing
	case ch.TrivialPattern:
		var offset = inner_scope.AddBinding(p.ValueName, p.Point)
		var inst_store = InstStore(offset)
		inner_buf.Write(CodeFrom(inst_store, info))
	case ch.TuplePattern:
		BindPatternItems(pattern, p.Items, inner_scope, inner_buf)
	case ch.RecordPattern:
		BindPatternItems(pattern, p.Items, inner_scope, inner_buf)
	default:
		panic("impossible branch")
	}
	var body_code = CompileExpr(lambda.Output, inner_ctx)
	inner_buf.Write(body_code)
	var base_reserved_size = *(inner_scope.BindingPeek)
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
				ctx.LocalScope.Bindings[i].Used = true
				continue
			}
			var offset = base_context_size
			base_context_size += 1
			context_offset_map[i] = offset
			context_outer_offsets = append(context_outer_offsets, i)
			ctx.LocalScope.Bindings[i].Used = true
		}
	}
	if recursive {
		if rec_used {
			context_offset_map[rec_outer_offset] = base_context_size
		}
		base_context_size += 1
	}
	if base_context_size > def.ClosureMaxSize {
		panic("maximum closure size exceeded")
	}
	if (base_context_size + base_reserved_size) > def.LocalSlotMaxSize {
		panic("maximum quantity of local bindings exceeded")
	}
	var raw_inner_code = inner_buf.Collect()
	var inst_seq_len = len(raw_inner_code.InstSeq)
	var final_inst_seq = make([] def.Instruction, inst_seq_len)
	for i, inst := range raw_inner_code.InstSeq {
		if inst.OpCode == def.LOAD || inst.OpCode == def.STORE {
			var offset = inst.GetOffset()
			var new_offset uint
			if offset < outer_bindings_size {
				new_offset = context_offset_map[offset]
			} else {
				new_offset = offset - outer_bindings_size + base_context_size
			}
			final_inst_seq[i] = def.Instruction {
				OpCode: inst.OpCode,
				Arg0:   0,
				Arg1:   def.Long(new_offset),
			}
		} else {
			final_inst_seq[i] = inst
		}
	}
	var final_inner_code = Code {
		InstSeq:   final_inst_seq,
		SourceMap: raw_inner_code.SourceMap,
	}
	var f = &def.Function {
		Kind: def.F_USER,
		Code: final_inner_code.InstSeq,
		BaseSize: def.FrameBaseSize {
			Context:  def.Short(base_context_size),
			Reserved: def.Long(base_reserved_size),
		},
		Info:     def.FuncInfo {
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
	var rec_flag def.Short
	if recursive {
		rec_flag = 1
	} else {
		rec_flag = 0
	}
	var ctx_inst = def.Instruction {
		OpCode: def.CTX,
		Arg0:   rec_flag,
		Arg1:   0,
	}
	outer_buf.Write(CodeFrom(ctx_inst, info))
	return outer_buf.Collect()
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
		var get = InstGet(item.Index)
		var offset = scope.AddBinding(item.Name, item.Point)
		var store = InstStore(offset)
		buf.Write(CodeFrom(get, info))
		buf.Write(CodeFrom(store, info))
	}
	var pop = def.Instruction { OpCode: def.POP }
	buf.Write(CodeFrom(pop, info))
}
