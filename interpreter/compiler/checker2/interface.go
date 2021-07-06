package checker2

import (
	"kumachan/interpreter/lang/common/source"
	"kumachan/interpreter/compiler/checker2/typsys"
	"kumachan/interpreter/lang/common/name"
)


// TODO: consider field getter methods { Index uint }
type DispatchMapping (map[ImplPair] ([] *Function))
type ImplPair struct {
	ConcreteType   *typsys.TypeDef
	InterfaceType  *typsys.TypeDef
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
				var methods = make([] *Function, len(required))
				for i, field := range required {
					var method_name = field.Name
					var method_t = methodConcreteType(con, impl, field.Type)
					var method_full_name = name.Name {
						ModuleName: con.Name.ModuleName,
						ItemName:   field.Name,
					}
					var info = ImplErrorInfo {
						Concrete:  con.Name.String(),
						Interface: impl.Name.String(),
						Method:    method_name,
					}
					var group, exists = functions[method_full_name]
					if !(exists) {
						return source.MakeError(con.Location,
							E_ImplMethodNoSuchFunction { ImplErrorInfo: info })
					}
					var method_f *Function = nil
					var found = false
					for _, f := range group {
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
									ImplErrorInfo: info,
								})
						}
						method_f = f
						found = true
					}
					if !(found) {
						return source.MakeError(con.Location,
							E_ImplMethodNoneCompatible {
								ImplErrorInfo: info,
							})
					}
					methods[i] = method_f
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

func methodConcreteType(
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


