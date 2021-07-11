package checker2

import (
	"kumachan/interpreter/lang/common/name"
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
)


type DispatchMapping (map[ImplPair] ([] Method))
type ImplPair struct {
	ConcreteType   *typsys.TypeDef
	InterfaceType  *typsys.TypeDef
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
	var impls_cache = make(map[*typsys.TypeDef] ([] *typsys.TypeDef))
	var errs source.Errors
	for _, def := range types {
		var _, is_interface = def.Content.(*typsys.Interface)
		if is_interface {
			continue
		}
		var con = def.TypeDef
		var con_t typsys.Type = &typsys.NestedType {
			Content: typsys.Ref {
				Def:  con,
				Args: (func() ([] typsys.Type) {
					var args = make([] typsys.Type, len(con.Parameters))
					for i := range con.Parameters {
						var p = &(con.Parameters[i])
						args[i] = typsys.ParameterType { Parameter: p }
					}
					return args
				})(),
			},
		}
		var impls = getAllImpls(con, impls_cache)
		var err = (func() *source.Error {
			for _, impl := range impls {
				var interface_ = impl.Content.(*typsys.Interface)
				var required = interface_.Methods.Fields
				var methods = make([] Method, len(required))
				for i, field := range required {
					var method_name = field.Name
					var method_t = methodConcreteType(con, impl, field.Type)
					var method_full_name = name.Name {
						ModuleName: con.Name.ModuleName,
						ItemName:   field.Name,
					}
					var detail = func() ImplError { return ImplError {
						Concrete:  con.Name.String(),
						Interface: impl.Name.String(),
						Method:    method_name,
					} }
					var m_group, func_exists = functions[method_full_name]
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
						return source.MakeError(con.Location,
							E_ImplMethodAmbiguous {
								ImplError: detail(),
							})
					}
					if field_exists {
						methods[i] = MethodField { Index: *m_index }
					}
					if !(func_exists) {
						return source.MakeError(con.Location,
							E_ImplMethodNoSuchFunction {
								ImplError: detail(),
							})
					}
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
							return source.MakeError(con.Location,
								E_ImplMethodDuplicateCompatible {
									ImplError: detail(),
								})
						}
						method_f = f
						found = true
					}
					if !(found) {
						return source.MakeError(con.Location,
							E_ImplMethodNoneCompatible {
								ImplError: detail(),
							})
					}
					methods[i] = MethodFunction { Function: method_f }
				}
				var pair = ImplPair {
					ConcreteType:  con,
					InterfaceType: impl,
				}
				mapping[pair] = methods
			}
			return nil
		})()
		source.ErrorsJoin(&errs, err)
	}
	if errs != nil { return nil, errs }
	return mapping, nil
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

func getAllImpls (
	def    *typsys.TypeDef,
	cache  (map[*typsys.TypeDef] ([] *typsys.TypeDef)),
) ([] *typsys.TypeDef) {
	var existing, exists = cache[def]
	if exists {
		return existing
	}
	var impls = make([] *typsys.TypeDef, 0)
	for _, impl := range def.Implements {
		var _, is_interface = impl.Content.(*typsys.Interface)
		if !(is_interface) { panic("something went wrong") }
		impls = append(impls, getAllImpls(impl, cache)...)
	}
	impls = append(impls, def.Implements...)
	cache[def] = impls
	return impls
}

func methodConcreteType (
	con       *typsys.TypeDef,
	impl      *typsys.TypeDef,
	raw_type  typsys.Type,
) typsys.Type {
	if len(con.Parameters) != len(impl.Parameters) {
		panic("something went wrong")
	}
	return typsys.TypeOpMap(raw_type, func(t typsys.Type) (typsys.Type, bool) {
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


