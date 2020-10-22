package kmd

import (
	"fmt"
	"errors"
	"reflect"
	"math/big"
)


const TagKmd = "kmd"
const KmdIgnore = "ignore"

type StringKind uint
const (
	GoString StringKind = iota
	RuneSlice
)
type AdapterId struct {
	From  TypeId
	To    TypeId
}
type GoStructOptions struct {
	StringKind
	Types               map[TypeId] reflect.Type
	GetAlgebraicTypeId  (func(reflect.Type) TypeId)
	GoStructSerializerOptions
	GoStructDeserializerOptions
}
type GoStructSerializerOptions struct {
}
type GoStructDeserializerOptions struct {
	Adapters  map[AdapterId] (func(Object) Object)
}

func getInterfaceValueFromType(t reflect.Type) interface{} {
	var ptr = reflect.New(t)
	var v = ptr.Elem()
	var i = v.Interface()
	return i
}

func CreateGoStructTransformer(opts GoStructOptions) Transformer {
	var determine_type (func(Object) *Type)
	determine_type = func(obj Object) *Type {
		switch obj.(type) {
		case bool:
			return PrimitiveType(Bool)
		case float64:
			return PrimitiveType(Float)
		case uint32:
			return PrimitiveType(Uint32)
		case int32:
			return PrimitiveType(Int32)
		case uint64:
			return PrimitiveType(Uint64)
		case int64:
			return PrimitiveType(Int64)
		case *big.Int:
			return PrimitiveType(Int)
		case string:
			if opts.StringKind == GoString {
				return PrimitiveType(String)
			} else {
				panic("inconsistent string kind")
			}
		case [] byte:
			return PrimitiveType(Binary)
		case [] rune:
			if opts.StringKind == RuneSlice {
				return PrimitiveType(String)
			} else {
				panic("inconsistent string kind")
			}
		default:
			var v = reflect.ValueOf(obj)
			var t = v.Type()
			if t.Kind() == reflect.Slice {
				var elem = getInterfaceValueFromType(t.Elem())
				return ContainerType(Array, determine_type(elem))
			} else if t.Kind() == reflect.Interface {
				if t.NumMethod() == 1 && t.Method(0).Name == "Maybe" {
					if t.Method(0).Type.NumIn() != 2 {
						panic(fmt.Sprintf("%s: Maybe() method should have signature (T,MaybeT)", t))
					}
					var elem_t = t.Method(0).Type.In(0)
					var elem = getInterfaceValueFromType(elem_t)
					return ContainerType(Optional, determine_type(elem))
				} else {
					var id = opts.GetAlgebraicTypeId(t)
					return AlgebraicType(Union, id)
				}
			} else if t.Kind() == reflect.Struct {
				var id = opts.GetAlgebraicTypeId(t)
				return AlgebraicType(Record, id)
			} else {
				panic(fmt.Sprintf("unsupported type: %s", t))
			}
		}
	}
	var get_reflect_type func(*Type) reflect.Type
	get_reflect_type = func(t *Type) reflect.Type {
		switch t.Kind {
		case Bool:   return reflect.TypeOf(true)
		case Float:  return reflect.TypeOf(float64(0.0))
		case Uint32: return reflect.TypeOf(uint32(0))
		case Int32:  return reflect.TypeOf(int32(0))
		case Uint64: return reflect.TypeOf(uint64(0))
		case Int64:  return reflect.TypeOf(int64(0))
		case Int:    return reflect.TypeOf((*big.Int)(nil))
		case String:
			if opts.StringKind == GoString {
				return reflect.TypeOf("")
			} else if opts.StringKind == RuneSlice {
				return reflect.TypeOf(([] rune)(""))
			} else {
				panic("impossible branch")
			}
		case Binary:
			return reflect.TypeOf([] byte {})
		case Array:
			return reflect.SliceOf(get_reflect_type(t.ElementType))
		case Optional:
			var elem_t = get_reflect_type(t.ElementType)
			var method, ok = elem_t.MethodByName("Maybe")
			if !ok {
				panic(fmt.Sprintf("%s: Maybe() method not found", elem_t))
			}
			if method.Type.NumIn() != 2 {
				panic(fmt.Sprintf("%s: Maybe() method should have signature (T,MaybeT)", elem_t))
			}
			return method.Type.In(1)
		case Record:
			var rt, ok = opts.Types[t.Identifier]
			if !ok { panic(fmt.Sprintf("unknown type %s", t.Identifier)) }
			return rt
		case Tuple:
			panic("tuple is not supported in Go")
		case Union:
			var rt, ok = opts.Types[t.Identifier]
			if !ok { panic(fmt.Sprintf("unknown type %s", t.Identifier)) }
			return rt
		default:
			panic("impossible branch")
		}
	}
	var serializer = Serializer {
		DetermineType: determine_type,
		PrimitiveSerializer: PrimitiveSerializer {
			WriteBool:   func(obj Object) bool { return obj.(bool) },
			WriteFloat:  func(obj Object) float64 { return obj.(float64) },
			WriteUint32: func(obj Object) uint32 { return obj.(uint32) },
			WriteInt32:  func(obj Object) int32 { return obj.(int32) },
			WriteUint64: func(obj Object) uint64 { return obj.(uint64) },
			WriteInt64:  func(obj Object) int64 { return obj.(int64) },
			WriteInt:    func(obj Object) *big.Int { return obj.(*big.Int) },
			WriteString: func(obj Object) string {
				if opts.StringKind == GoString {
					return obj.(string)
				} else if opts.StringKind == RuneSlice {
					return string(obj.([] rune))
				} else {
					panic("impossible branch")
				}
			},
			WriteBinary: func(obj Object) ([] byte) {
				return obj.([] byte)
			},
		},
		ContainerSerializer: ContainerSerializer {
			IterateArray: func(obj Object, f func(uint, Object) error) error {
				var v = reflect.ValueOf(obj)
				for i := 0; i < v.Len(); i += 1 {
					err := f(uint(i), v.Index(i).Interface())
					if err != nil { return err }
				}
				return nil
			},
			UnwrapOptional: func(obj Object) (Object, bool) {
				if obj != nil {
					var v = reflect.ValueOf(obj)
					return v.Elem().Interface(), true
				} else {
					return nil, false
				}
			},
		},
		AlgebraicSerializer: AlgebraicSerializer {
			IterateRecord: func(obj Object, f func(string, Object) error) error {
				var v = reflect.ValueOf(obj)
				for i := 0; i < v.NumField(); i += 1 {
					var field_t = v.Type().Field(i)
					var field_v = v.Field(i)
					if field_t.Tag.Get(TagKmd) != KmdIgnore {
						err := f(field_t.Name, field_v)
						if err != nil { return err }
					}
				}
				return nil
			},
			IterateTuple:  func(Object, func(Object) error) error {
				panic("tuple is not supported in Go")
			},
			Union2Case: func(obj Object) Object {
				var v = reflect.ValueOf(obj)
				return v.Elem().Interface()
			},
		},
	}
	var deserializer = Deserializer {
		PrimitiveDeserializer: PrimitiveDeserializer {
			ReadBool:   func(obj bool) Object { return obj },
			ReadFloat:  func(obj float64) Object { return obj },
			ReadUint32: func(obj uint32) Object { return obj },
			ReadInt32:  func(obj int32) Object { return obj },
			ReadUint64: func(obj uint64) Object { return obj },
			ReadInt64:  func(obj int64) Object { return obj },
			ReadInt:    func(obj *big.Int) Object { return obj },
			ReadString: func(str string) Object {
				if opts.StringKind == GoString {
					return str
				} else if opts.StringKind == RuneSlice {
					return ([] rune)(str)
				} else {
					panic("impossible branch")
				}
			},
			ReadBinary: func(bytes ([] byte)) Object {
				return bytes
			},
		},
		ContainerDeserializer: ContainerDeserializer {
			CreateArray: func(array_t *Type, cap uint) Object {
				var elem_t = array_t.ElementType
				var slice_t = reflect.SliceOf(get_reflect_type(elem_t))
				var slice_v = reflect.MakeSlice(slice_t, 0, int(cap))
				return slice_v.Interface()
			},
			AppendItem: func(array Object, item Object) Object {
				var array_v = reflect.ValueOf(array)
				var item_v = reflect.ValueOf(item)
				var appended_v = reflect.Append(array_v, item_v)
				return appended_v.Interface()
			},
			Just: func(obj Object, opt_t *Type) Object {
				var opt_rt = get_reflect_type(opt_t)
				var v = reflect.ValueOf(obj)
				var just_v = v.Convert(opt_rt)
				return just_v.Interface()
			},
			Nothing: func(opt_t *Type) Object {
				var opt_rt = get_reflect_type(opt_t)
				var elem_v = reflect.New(opt_rt).Elem()
				return elem_v.Interface()
			},
		},
		AlgebraicDeserializer: AlgebraicDeserializer {
			AssignObject: func(obj Object, from *Type, to *Type) (Object, error) {
				if from.Kind == Record && to.Kind == Record &&
					from.Identifier.TypeIdBase == to.Identifier.TypeIdBase &&
					from.Identifier.Version != to.Identifier.Version {
					var adapter_id = AdapterId {
						From: from.Identifier,
						To:   to.Identifier,
					}
					var adapter, exists = opts.Adapters[adapter_id]
					if exists {
						return adapter(obj), nil
					} else {
						return nil, errors.New("types are not compatible: " +
							fmt.Sprintf("\n\t%s\nis not adaptable to\n\t%s\n", from, to))
					}
				} else if TypeEqual(from, to) {
					return obj, nil
				} else {
					return nil, errors.New("types are not compatible: " +
						fmt.Sprintf("\n\t%s\ndoes not equal to\n\t%s\n", from, to))
				}
			},
			CheckRecord: func(record_t TypeId, size uint) error {
				var rt, exists = opts.Types[record_t]
				if !(exists) { return errors.New(fmt.Sprintf(
					"type %s does not exist", record_t)) }
				if rt.Kind() != reflect.Struct { return errors.New(fmt.Sprintf(
					"type %s is not a record type", record_t))}
				var valid_field_count = uint(0)
				for i := 0; i < rt.NumField(); i += 1 {
					if rt.Field(i).Tag.Get(TagKmd) != KmdIgnore {
						valid_field_count += 1
					}
				}
				if valid_field_count != size { return errors.New(fmt.Sprintf(
					"record size not matching: given %d, require %d",
					size, valid_field_count))}
				return nil
			},
			GetFieldType: func(record_t TypeId, field string) (*Type, error) {
				var rt, ok = opts.Types[record_t]
				if !(ok) { panic("record type existence should be checked" +
					" before trying to get a field type") }
				var field_info, exists = rt.FieldByName(field)
				if !(exists) { return nil, errors.New(fmt.Sprintf(
					"field %s does not exist on type %s", field, record_t))}
				var obj = getInterfaceValueFromType(field_info.Type)
				var t = determine_type(obj)
				return t, nil
			},
			CreateRecord: func(record_t TypeId) Object {
				var rt, ok = opts.Types[record_t]
				if !(ok) { panic("record type existence should be checked" +
					" before trying to create a record") }
				var struct_ptr = reflect.New(rt)
				var struct_v = struct_ptr.Elem()
				return struct_v.Interface()
			},
			FillField: func(record Object, field string, value Object) {
				var record_v = reflect.ValueOf(record)
				record_v.FieldByName(field).Set(reflect.ValueOf(value))
			},
			CheckTuple: func(TypeId, uint) error {
				panic("tuple is not supported in Go")
			},
			GetElementType: func(TypeId, uint) (*Type, error) {
				panic("tuple is not supported in Go")
			},
			CreateTuple: func(TypeId) Object {
				panic("tuple is not supported in Go")
			},
			FillElement: func(Object, uint, Object) {
				panic("tuple is not supported in Go")
			},
			Case2Union: func(obj Object, union_t TypeId, case_t TypeId) (Object, error) {
				var union_rt, exists = opts.Types[union_t]
				if !(exists) { return nil, errors.New(fmt.Sprintf(
					"type %s dose not exist", union_t)) }
				var obj_v = reflect.ValueOf(obj)
				if obj_v.Type().ConvertibleTo(union_rt) {
					var obj_union_v = obj_v.Convert(union_rt)
					return obj_union_v.Interface(), nil
				} else {
					return nil, errors.New(fmt.Sprintf(
						"%s is not a case type of the union type %s",
						case_t, union_t))
				}
			},
		},
	}
	return Transformer {
		Serializer:   serializer,
		Deserializer: deserializer,
	}
}
