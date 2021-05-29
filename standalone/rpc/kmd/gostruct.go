package kmd

import (
	"fmt"
	"errors"
	"reflect"
	"math/big"
)


const Tag = "kmd"
const TagIgnore = "kmd_ignore"
const MaybeMethod = "Maybe"

type IntegerKind uint
const (
	BigInt IntegerKind = iota
	Int64
)
type StringKind uint
const (
	GoString StringKind = iota
	RuneSlice
)
type GoStructOptions struct {
	Types  map[TypeId] reflect.Type
	IntegerKind
	StringKind
	GoStructSerializerOptions
	GoStructDeserializerOptions
}
type GoStructSerializerOptions struct {
}
type GoStructDeserializerOptions struct {
	Adapters  map[AdapterId] (func(Object) Object)
}
type GoInterfaceWorkaround struct {
	Type      reflect.Type
	Concrete  Object
}

func getInterfaceValueFromType(t reflect.Type) interface{} {
	if t.Kind() == reflect.Interface &&
		!(t.NumMethod() == 1 && t.Method(0).Name == MaybeMethod) {
		return GoInterfaceWorkaround { Type: t }
	} else {
		var ptr = reflect.New(t)
		var v = ptr.Elem()
		var i = v.Interface()
		return i
	}
}
func isEnumType(t reflect.Type) bool {
	var obj = getInterfaceValueFromType(t)
	var _, is = obj.(GoInterfaceWorkaround)
	return is
}

func CreateGoStructTransformer(opts GoStructOptions) Transformer {
	var types_rev = make(map[reflect.Type] TypeId)
	for id, rt := range opts.Types {
		var _, exists = types_rev[rt]
		if exists {
			panic(fmt.Sprintf("more than one id for the type %s", rt))
		}
		types_rev[rt] = id
	}
	var determine_type (func(Object) *Type)
	determine_type = func(obj Object) *Type {
		switch obj.(type) {
		case bool:
			return PrimitiveType(Bool)
		case float64:
			return PrimitiveType(Float)
		case complex128:
			return PrimitiveType(Complex)
		case *big.Int:
			if opts.IntegerKind != BigInt { panic("inconsistent integer kind") }
			return PrimitiveType(Integer)
		case int64:
			if opts.IntegerKind != Int64 { panic("inconsistent integer kind") }
			return PrimitiveType(Integer)
		case string:
			if opts.StringKind != GoString { panic("inconsistent string kind") }
			return PrimitiveType(String)
		case [] rune:
			if opts.StringKind != RuneSlice { panic("inconsistent string kind") }
			return PrimitiveType(String)
		case [] byte:
			return PrimitiveType(Binary)
		default:
			var t reflect.Type
			var workaround, is_workaround = obj.(GoInterfaceWorkaround)
			if is_workaround {
				t = workaround.Type
			} else {
				t = reflect.TypeOf(obj)
			}
			if t.Kind() == reflect.Slice {
				var elem = getInterfaceValueFromType(t.Elem())
				return ContainerType(Array, determine_type(elem))
			} else if t.Kind() == reflect.Interface {
				if t.NumMethod() == 1 && t.Method(0).Name == MaybeMethod {
					if t.Method(0).Type.NumIn() != 2 {
						panic(fmt.Sprintf("%s: Maybe() method should have signature (T,MaybeT)", t))
					}
					var elem_t = t.Method(0).Type.In(0)
					var elem = getInterfaceValueFromType(elem_t)
					return ContainerType(Optional, determine_type(elem))
				} else {
					var id, exists = types_rev[t]
					if !(exists) {
						panic(fmt.Sprintf("the type %s does not have an id", t))
					}
					return AlgebraicType(Enum, id)
				}
			} else if t.Kind() == reflect.Struct {
				var id, exists = types_rev[t]
				if !(exists) {
					panic(fmt.Sprintf("the type %s does not have an id", t))
				}
				return AlgebraicType(Record, id)
			} else {
				panic(fmt.Sprintf("unsupported type: %s", t))
			}
		}
	}
	var get_reflect_type func(*Type) reflect.Type
	get_reflect_type = func(t *Type) reflect.Type {
		switch t.kind {
		case Bool:    return reflect.TypeOf(true)
		case Float:   return reflect.TypeOf(float64(0.0))
		case Complex: return reflect.TypeOf(complex128(complex(0.0, 0.0)))
		case Integer:
			switch opts.IntegerKind {
			case BigInt:
				return reflect.TypeOf((*big.Int)(nil))
			case Int64:
				return reflect.TypeOf(int64(0))
			default:
				panic("impossible branch")
			}
		case String:
			switch opts.StringKind {
			case GoString:
				return reflect.TypeOf("")
			case RuneSlice:
				return reflect.TypeOf(([] rune)(""))
			default:
				panic("impossible branch")
			}
		case Binary:
			return reflect.TypeOf([] byte {})
		case Array:
			return reflect.SliceOf(get_reflect_type(t.elementType))
		case Optional:
			var elem_t = get_reflect_type(t.elementType)
			var method, ok = elem_t.MethodByName(MaybeMethod)
			if !ok {
				panic(fmt.Sprintf("%s: Maybe() method not found", elem_t))
			}
			if method.Type.NumIn() != 2 {
				panic(fmt.Sprintf("%s: Maybe() method should have signature (T,MaybeT)", elem_t))
			}
			return method.Type.In(1)
		case Record:
			var rt, ok = opts.Types[t.identifier]
			if !ok { panic(fmt.Sprintf("unknown type %s", t.identifier)) }
			return rt
		case Tuple:
			panic("tuple is not supported in Go")
		case Enum:
			var rt, ok = opts.Types[t.identifier]
			if !ok { panic(fmt.Sprintf("unknown type %s", t.identifier)) }
			return rt
		default:
			panic("impossible branch")
		}
	}
	var serializer = Serializer {
		DetermineType: determine_type,
		PrimitiveSerializer: PrimitiveSerializer {
			WriteBool:    func(obj Object) bool { return obj.(bool) },
			WriteFloat:   func(obj Object) float64 { return obj.(float64) },
			WriteComplex: func(obj Object) complex128 { return obj.(complex128) },
			WriteInteger: func(obj Object) *big.Int {
				switch opts.IntegerKind {
				case BigInt:
					return obj.(*big.Int)
				case Int64:
					return big.NewInt(obj.(int64))
				default:
					panic("impossible branch")
				}
			},
			WriteString: func(obj Object) string {
				switch opts.StringKind {
				case GoString:
					return obj.(string)
				case RuneSlice:
					return string(obj.([] rune))
				default:
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
				var elem_t = v.Type().Elem()
				var is_elem_enum = isEnumType(elem_t)
				for i := 0; i < v.Len(); i += 1 {
					var elem = v.Index(i).Interface()
					if is_elem_enum {
						elem = GoInterfaceWorkaround {
							Type:     elem_t,
							Concrete: elem,
						}
					}
					err := f(uint(i), elem)
					if err != nil { return err }
				}
				return nil
			},
			UnwrapOptional: func(obj Object) (Object, bool) {
				if obj != nil {
					return obj, true
				} else {
					return nil, false
				}
			},
		},
		AlgebraicSerializer: AlgebraicSerializer {
			IterateRecord: func(obj Object, f func(string, Object) error) error {
				var v = reflect.ValueOf(obj)
				for i := 0; i < v.NumField(); i += 1 {
					var field_info = v.Type().Field(i)
					var field_t = field_info.Type
					var field_v = v.Field(i)
					var field_obj = field_v.Interface()
					if isEnumType(field_t) {
						field_obj = GoInterfaceWorkaround {
							Type:     field_t,
							Concrete: field_obj,
						}
					}
					var _, ignore = field_info.Tag.Lookup(TagIgnore)
					if !(ignore) {
						var tagged_name = field_info.Tag.Get(Tag)
						var name string
						if tagged_name != "" {
							name = tagged_name
						} else {
							name = field_info.Name
						}
						err := f(name, field_obj)
						if err != nil { return err }
					}
				}
				return nil
			},
			IterateTuple:  func(Object, func(uint,Object) error) error {
				panic("tuple is not supported in Go")
			},
			Enum2Case: func(obj Object) Object {
				var workaround, is_workaround = obj.(GoInterfaceWorkaround)
				if is_workaround {
					return workaround.Concrete
				} else {
					panic("something went wrong")
				}
			},
		},
	}
	var deserializer = Deserializer {
		PrimitiveDeserializer: PrimitiveDeserializer {
			ReadBool:    func(obj bool) Object { return obj },
			ReadFloat:   func(obj float64) Object { return obj },
			ReadComplex: func(obj complex128) Object { return obj },
			ReadInteger: func(obj *big.Int) (Object, bool) {
				switch opts.IntegerKind {
				case BigInt:
					return obj, true
				case Int64:
					if obj.IsInt64() {
						return obj.Int64(), true
					} else {
						return nil, false
					}
				default:
					panic("impossible branch")
				}
			},
			ReadString: func(str string) Object {
				switch opts.StringKind {
				case GoString:
					return str
				case RuneSlice:
					return ([] rune)(str)
				default:
					panic("impossible branch")
				}
			},
			ReadBinary: func(bytes ([] byte)) Object {
				return bytes
			},
		},
		ContainerDeserializer: ContainerDeserializer {
			CreateArray: func(array_t *Type) Object {
				var elem_t = array_t.elementType
				var slice_t = reflect.SliceOf(get_reflect_type(elem_t))
				var slice_v = reflect.MakeSlice(slice_t, 0, 0)
				return slice_v.Interface()
			},
			AppendItem: func(array *Object, item Object) {
				var array_v = reflect.ValueOf(*array)
				var item_v = reflect.ValueOf(item)
				var appended_v = reflect.Append(array_v, item_v)
				var appended = appended_v.Interface()
				*array = appended
			},
			Some: func(obj Object, opt_t *Type) Object {
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
				if from.kind == Record && to.kind == Record &&
					from.identifier.TypeIdFuzzy == to.identifier.TypeIdFuzzy &&
					from.identifier.Version != to.identifier.Version {
					var adapter_id = AdapterId {
						From: from.identifier,
						To:   to.identifier,
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
					var _, ignore = rt.Field(i).Tag.Lookup(TagIgnore)
					if !(ignore) {
						valid_field_count += 1
					}
				}
				if valid_field_count != size { return errors.New(fmt.Sprintf(
					"record size not matching: given %d, require %d",
					size, valid_field_count))}
				return nil
			},
			GetFieldInfo: func(record_t TypeId, field string) (*Type, uint, error) {
				var rt, ok = opts.Types[record_t]
				if !(ok) { panic("record type existence should be checked" +
					" before trying to get a field type") }
				var field_info reflect.StructField
				var exists = false
				for i := 0; i < rt.NumField(); i += 1 {
					var this = rt.Field(i)
					var _, ignore = this.Tag.Lookup(TagIgnore)
					if !(ignore) {
						if this.Tag.Get(Tag) == field || this.Name == field {
							field_info = this
							exists = true
							break
						}
					}
				}
				if !(exists) { return nil, ^uint(0), errors.New(fmt.Sprintf(
					"field %s does not exist on type %s", field, record_t))}
				var obj = getInterfaceValueFromType(field_info.Type)
				var t = determine_type(obj)
				var index = uint(field_info.Index[0])
				return t, index, nil
			},
			CreateRecord: func(record_t TypeId) Object {
				var rt, ok = opts.Types[record_t]
				if !(ok) { panic("record type existence should be checked" +
					" before trying to create a record") }
				var struct_ptr = reflect.New(rt)
				return struct_ptr.Interface()
			},
			FillField: func(record Object, index uint, value Object) {
				var record_v = reflect.ValueOf(record).Elem()
				var field_v = record_v.Field(int(index))
				field_v.Set(reflect.ValueOf(value))
			},
			FinishRecord: func(record Object, _ TypeId) (Object, error) {
				return reflect.ValueOf(record).Elem().Interface(), nil
			},
			CheckTuple: func(TypeId, uint) error {
				panic("tuple is not supported in Go")
			},
			GetElementType: func(TypeId, uint) *Type {
				panic("tuple is not supported in Go")
			},
			CreateTuple: func(TypeId) Object {
				panic("tuple is not supported in Go")
			},
			FillElement: func(Object, uint, Object) {
				panic("tuple is not supported in Go")
			},
			FinishTuple: func(Object, TypeId) (Object, error) {
				panic("tuple is not supported in Go")
			},
			Case2Enum: func(obj Object, enum_t TypeId, case_t TypeId) (Object, error) {
				var enum_rt, exists = opts.Types[enum_t]
				if !(exists) { return nil, errors.New(fmt.Sprintf(
					"type %s dose not exist", enum_t)) }
				var obj_v = reflect.ValueOf(obj)
				if obj_v.Type().ConvertibleTo(enum_rt) {
					var obj_enum_v = obj_v.Convert(enum_rt)
					return obj_enum_v.Interface(), nil
				} else {
					return nil, errors.New(fmt.Sprintf(
						"%s is not a case type of the enum type %s",
						case_t, enum_t))
				}
			},
		},
	}
	return Transformer {
		Serializer:   &serializer,
		Deserializer: &deserializer,
	}
}
