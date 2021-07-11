package checker2

import (
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
)


type DispatchMapping (map[ImplPair] *DispatchTable)
type ImplPair struct {
	ConcreteType   *typsys.TypeDef
	InterfaceType  *typsys.TypeDef
}
type DispatchTable struct {
	Methods   [] Method
	Included  [] *DispatchTable
}
type Method interface { method() }
func (MethodFunction) method() {}
type MethodFunction struct {
	Function  *Function
}
func (MethodField) method() {}
type MethodField struct {
	Index  uint
}

func generateDispatchMapping (
	types      typeList,
	functions  FunctionRegistry,
) (DispatchMapping, source.Errors) {
	var mapping = make(DispatchMapping)
	var errs source.Errors
	for _, def := range types {
		var _, is_interface = def.Content.(*typsys.Interface)
		if is_interface {
			continue
		}
		var con = def.TypeDef
		var err = (func() *source.Error {
			for _, impl := range con.Implements {
				var _, err = makeDispatchTable(functions, con, impl, mapping)
				if err != nil { return err }
			}
			return nil
		})()
		source.ErrorsJoin(&errs, err)
	}
	if errs != nil { return nil, errs }
	return mapping, nil
}

func makeDispatchTable (
	f_reg    FunctionRegistry,
	con      *typsys.TypeDef,
	impl     *typsys.TypeDef,
	mapping  DispatchMapping,
) (*DispatchTable, *source.Error) {
	var _, con_invalid = con.Content.(*typsys.Interface)
	if con_invalid {
		panic("invalid argument")
	}
	var pair = ImplPair {
		ConcreteType:  con,
		InterfaceType: impl,
	}
	var existing, exists = mapping[pair]
	if exists {
		// diamond situation
		return existing, nil
	}
	var con_t = defType(con)
	var fields = impl.Content.(*typsys.Interface).Methods.Fields
	var methods = make([] Method, len(fields))
	for i, field := range fields {
		var method_name = field.Name
		var method_t = methodType(con, impl, field.Type)
		var method_full_name = name.Name {
			ModuleName: con.Name.ModuleName,
			ItemName:   field.Name,
		}
		var detail = func() ImplError { return ImplError {
			Concrete:  con.Name.String(),
			Interface: impl.Name.String(),
			Method:    method_name,
		} }
		var m_group, func_exists = f_reg[method_full_name]
		var m_index, field_exists = (func() (*uint, bool) {
			var record, is_record = getBoxedRecord(con)
			if is_record {
				var index, exists = record.FieldIndexMap[method_name]
				return &index, exists
			} else {
				return nil, false
			}
		})()
		if func_exists && field_exists {
			return nil, source.MakeError(con.Location,
				E_ImplMethodAmbiguous {
					ImplError: detail(),
				})
		} else if !(func_exists) && field_exists {
			methods[i] = MethodField { Index: *m_index }
		} else if func_exists && !(field_exists) {
			var method_f *Function = nil
			var found = false
			for _, f := range m_group {
				if len(f.Signature.ImplicitContext.Fields) > 0 {
					continue
				}
				var io = f.Signature.InputOutput
				var s0 = typsys.StartInferring(con.Parameters)
				var ctx = typsys.MakeAssignContextWithoutSubtyping(s0)
				var in_ok, s1 = typsys.Assign(io.Input, con_t, ctx)
				if !(in_ok) {
					continue
				}
				typsys.ApplyNewInferringState(&ctx, s1)
				var out_ok, s2 = typsys.Assign(method_t, io.Output, ctx)
				if !(out_ok) {
					continue
				}
				typsys.ApplyNewInferringState(&ctx, s2)
				if found {
					return nil, source.MakeError(con.Location,
						E_ImplMethodDuplicateCompatible {
							ImplError: detail(),
						})
				}
				method_f = f
				found = true
			}
			if !(found) {
				return nil, source.MakeError(con.Location,
					E_ImplMethodNoneCompatible {
						ImplError: detail(),
					})
			}
			methods[i] = MethodFunction { Function: method_f }
		} else {
			return nil, source.MakeError(con.Location,
				E_ImplMethodNoSuchFunctionOrField {
					ImplError: detail(),
				})
		}
	}
	var included = make([] *DispatchTable, len(impl.Implements))
	for i, impl_impl := range impl.Implements {
		var table, err = makeDispatchTable(f_reg, con, impl_impl, mapping)
		if err != nil { return nil, err }
		included[i] = table
	}
	var table = &DispatchTable {
		Methods:  methods,
		Included: included,
	}
	mapping[pair] = table
	return table, nil
}

func getBoxedRecord(def *typsys.TypeDef) (typsys.Record, bool) {
	var box, is_box = def.Content.(*typsys.Box)
	if is_box {
		var record, is_record = getRecord(box.InnerType)
		return record, is_record
	} else {
		return typsys.Record {}, false
	}
}
func defType(def *typsys.TypeDef) typsys.Type {
	return &typsys.NestedType {
		Content: typsys.Ref {
			Def:  def,
			Args: (func() ([] typsys.Type) {
				var args = make([] typsys.Type, len(def.Parameters))
				for i := range def.Parameters {
					var p = &(def.Parameters[i])
					args[i] = typsys.ParameterType { Parameter: p }
				}
				return args
			})(),
		},
	}
}
func methodType(con *typsys.TypeDef, impl *typsys.TypeDef, raw typsys.Type) typsys.Type {
	if len(con.Parameters) != len(impl.Parameters) {
		panic("something went wrong")
	}
	return typsys.TypeOpMap(raw, func(t typsys.Type) (typsys.Type, bool) {
		var p, is_parameter = t.(typsys.ParameterType)
		if is_parameter {
			for i := range impl.Parameters {
				var impl_p = &(impl.Parameters[i])
				if p.Parameter == impl_p {
					var con_p = &(con.Parameters[i])
					return typsys.ParameterType { Parameter: con_p }, true
				}
			}
			return nil, false
		} else {
			return nil, false
		}
	})
}


