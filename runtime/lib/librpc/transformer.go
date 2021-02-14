package librpc

import (
	"fmt"
	"errors"
	"reflect"
	"math/big"
	"kumachan/rpc/kmd"
	"kumachan/runtime/lib/container"
	. "kumachan/lang"
)


type KmdTypedValue struct {
	Type   *kmd.Type
	Value  Value
}

type KmdFieldValue struct {
	Name   string
	Value  KmdTypedValue
}

func kmdCreateTransformer(ctx KmdTransformContext) kmd.Transformer {
	var conf = ctx.KmdGetInfo()
	var validate = func(obj kmd.Object, t kmd.TypeId) error {
		var v, exists = conf.KmdValidatorTable[kmd.ValidatorId(t)]
		if exists {
			var ok = ctx.KmdCallValidator(v, obj)
			if ok {
				return nil
			} else {
				return errors.New(fmt.Sprintf(
					"validation failed for type %s", t))
			}
		} else {
			return nil
		}
	}
	return kmd.Transformer {
		Serializer: &kmd.Serializer {
			DetermineType: func(obj kmd.Object) *kmd.Type {
				return obj.(KmdTypedValue).Type
			},
			PrimitiveSerializer: kmd.PrimitiveSerializer {
				WriteBool: func(obj kmd.Object) bool {
					var tv = obj.(KmdTypedValue)
					return FromBool(tv.Value.(SumValue))
				},
				WriteFloat: func(obj kmd.Object) float64 {
					return obj.(KmdTypedValue).Value.(float64)
				},
				WriteUint32: func(obj kmd.Object) uint32 {
					return obj.(KmdTypedValue).Value.(uint32)
				},
				WriteInt32: func(obj kmd.Object) int32 {
					return obj.(KmdTypedValue).Value.(int32)
				},
				WriteUint64: func(obj kmd.Object) uint64 {
					return obj.(KmdTypedValue).Value.(uint64)
				},
				WriteInt64:  func(obj kmd.Object) int64 {
					return obj.(KmdTypedValue).Value.(int64)
				},
				WriteInt:    func(obj kmd.Object) *big.Int {
					return obj.(KmdTypedValue).Value.(*big.Int)
				},
				WriteString: func(obj kmd.Object) string {
					var tv = obj.(KmdTypedValue)
					var str = tv.Value.(String)
					return GoStringFromString(str)
				},
				WriteBinary: func(obj kmd.Object) ([] byte) {
					return obj.(KmdTypedValue).Value.([] byte)
				},
			},
			ContainerSerializer: kmd.ContainerSerializer {
				IterateArray: func(obj kmd.Object, f func(uint, kmd.Object) error) error {
					var tv = obj.(KmdTypedValue)
					var arr = container.ArrayFrom(tv.Value)
					for i := uint(0); i < arr.Length; i += 1 {
						var item_v = arr.GetItem(i)
						var item_t = tv.Type.ElementType()
						var item_tv = KmdTypedValue {
							Type:  item_t,
							Value: item_v,
						}
						err := f(i, item_tv)
						if err != nil { return err }
					}
					return nil
				},
				UnwrapOptional: func(obj kmd.Object) (kmd.Object, bool) {
					var tv = obj.(KmdTypedValue)
					var sv = tv.Value.(SumValue)
					var inner, ok = Unwrap(sv)
					if ok {
						return KmdTypedValue {
							Type:  tv.Type.ElementType(),
							Value: inner,
						}, true
					} else {
						return nil, false
					}
				},
			},
			AlgebraicSerializer: kmd.AlgebraicSerializer{
				IterateRecord: func(obj kmd.Object, f func(string, kmd.Object) error) error {
					var tv = obj.(KmdTypedValue)
					var pv = tv.Value.(ProductValue)
					var tid = tv.Type.Identifier()
					var t, exists = conf.SchemaTable[tid]
					if !(exists) { panic("something went wrong") }
					var schema = t.(kmd.RecordSchema)
					var buffer = make([] KmdFieldValue, len(schema.Fields))
					for name, field := range schema.Fields {
						var field_t = field.Type
						var field_v = pv.Elements[field.Index]
						buffer[field.Index] = KmdFieldValue{
							Name:  name,
							Value: KmdTypedValue {
								Type:  field_t,
								Value: field_v,
							},
						}
					}
					for _, field := range buffer {
						var err = f(field.Name, field.Value)
						if err != nil { return err }
					}
					return nil
				},
				IterateTuple: func(obj kmd.Object, f func(uint,kmd.Object) error) error {
					var tv = obj.(KmdTypedValue)
					if tv.Value == nil {
						return nil
					}
					var tid = tv.Type.Identifier()
					var t, exists = conf.SchemaTable[tid]
					if !(exists) { panic("something went wrong") }
					var schema = t.(kmd.TupleSchema)
					var pv, is_pv = tv.Value.(ProductValue)
					if !(is_pv) {
						return f(0, KmdTypedValue {
							Type:  schema.Elements[0],
							Value: tv.Value,
						})
					}
					for i, elem_t := range schema.Elements {
						var elem_v = pv.Elements[i]
						err := f(uint(i), KmdTypedValue {
							Type:  elem_t,
							Value: elem_v,
						})
						if err != nil { return err }
					}
					return nil
				},
				Enum2Case: func(obj kmd.Object) kmd.Object {
					var tv = obj.(KmdTypedValue)
					var sv = tv.Value.(SumValue)
					var tid = tv.Type.Identifier()
					var t, exists = conf.SchemaTable[tid]
					if !(exists) { panic("something went wrong") }
					var schema = t.(kmd.EnumSchema)
					for case_tid, index := range schema.CaseIndexMap {
						if index == uint(sv.Index) {
							var case_t = conf.GetTypeFromId(case_tid)
							var case_v = sv.Value
							return KmdTypedValue {
								Type:  case_t,
								Value: case_v,
							}
						}
					}
					panic("something went wrong")
				},
			},
		},
		Deserializer: &kmd.Deserializer {
			PrimitiveDeserializer: kmd.PrimitiveDeserializer {
				ReadBool: func(v bool) kmd.Object {
					return ToBool(v)
				},
				ReadFloat:  func(v float64) kmd.Object { return v },
				ReadUint32: func(v uint32) kmd.Object { return v },
				ReadInt32:  func(v int32) kmd.Object { return v },
				ReadUint64: func(v uint64) kmd.Object { return v },
				ReadInt64:  func(v int64) kmd.Object { return v },
				ReadInt:    func(v *big.Int) kmd.Object { return v },
				ReadString: func(v string) kmd.Object {
					return StringFromGoString(v)
				},
				ReadBinary: func(v ([] byte)) kmd.Object { return v },
			},
			ContainerDeserializer: kmd.ContainerDeserializer{
				CreateArray: func(array_t *kmd.Type) kmd.Object {
					switch array_t.ElementType().Kind() {
					case kmd.Bool:   return make([] bool, 0)
					case kmd.Float:  return make([] float64, 0)
					case kmd.Uint32: return make([] uint32, 0)
					case kmd.Int32:  return make([] int32, 0)
					case kmd.Uint64: return make([] uint64, 0)
					case kmd.Int64:  return make([] int64, 0)
					default:         return make([] Value, 0)
					}
				},
				AppendItem: func(array_ptr *kmd.Object, item kmd.Object) {
					var array_v = reflect.ValueOf(*array_ptr)
					var item_v = reflect.ValueOf(item)
					var appended_v = reflect.Append(array_v, item_v)
					var appended = appended_v.Interface()
					*array_ptr = appended
				},
				Just: func(obj kmd.Object, _ *kmd.Type) kmd.Object {
					return Just(obj)
				},
				Nothing: func(_ *kmd.Type) kmd.Object {
					return Na()
				},
			},
			AlgebraicDeserializer: kmd.AlgebraicDeserializer {
				AssignObject: func(obj kmd.Object, from *kmd.Type, to *kmd.Type) (kmd.Object, error) {
					if kmd.TypeEqual(from, to) {
						return obj, nil
					} else if from.Identifier() != (kmd.TypeId {}) &&
						to.Identifier() != (kmd.TypeId {}) {
						var adapter_id = kmd.AdapterId {
							From: from.Identifier(),
							To:   to.Identifier(),
						}
						var info, exists = conf.KmdAdapterTable[adapter_id]
						if exists {
							var adapter = ctx.KmdGetAdapter(info.Index)
							var adapted = ctx.KmdCallAdapter(adapter, obj)
							return adapted, nil
						} else {
							return nil, errors.New(fmt.Sprintf(
								"the type %s cannot be adapted to the type %s",
								from, to))
						}
					} else {
						return nil, errors.New(fmt.Sprintf(
							"the type %s cannot be assigned to the type %s",
							from, to))
					}
				},
				CheckRecord: func(record_t kmd.TypeId, size uint) error {
					var t, exists = conf.SchemaTable[record_t]
					if !(exists) { return errors.New(fmt.Sprintf(
						"type %s does not exist", record_t)) }
					var schema, ok = t.(kmd.RecordSchema)
					if !(ok) { return errors.New(fmt.Sprintf(
						"type %s is not a record type", record_t)) }
					var schema_size = uint(len(schema.Fields))
					if schema_size != size { return errors.New(fmt.Sprintf(
						"record size not matching: given %d, require %d",
						size, schema_size)) }
					return nil
				},
				GetFieldInfo: func(record_t kmd.TypeId, name string) (*kmd.Type, uint, error) {
					var t, t_exists = conf.SchemaTable[record_t]
					if !(t_exists) { panic("something went wrong") }
					var schema = t.(kmd.RecordSchema)
					var schema_field, exists = schema.Fields[name]
					if !(exists) { return nil, ^uint(0), errors.New(fmt.Sprintf(
						"field %s does not exist on type %s", name, record_t)) }
					return schema_field.Type, schema_field.Index, nil
				},
				CreateRecord: func(record_t kmd.TypeId) kmd.Object {
					var t, exists = conf.SchemaTable[record_t]
					if !(exists) { panic("something went wrong") }
					var schema = t.(kmd.RecordSchema)
					var size = len(schema.Fields)
					return &ValProd {
						Elements: make([] Value, size),
					}
				},
				FillField: func(record kmd.Object, index uint, value kmd.Object) {
					var record_pv = record.(ProductValue)
					record_pv.Elements[index] = value
				},
				FinishRecord: func(record kmd.Object, t kmd.TypeId) (kmd.Object, error) {
					err := validate(record, t)
					if err != nil { return nil, err }
					return record, nil
				},
				CheckTuple: func(tuple_t kmd.TypeId, size uint) error {
					var t, exists = conf.SchemaTable[tuple_t]
					if !(exists) { return errors.New(fmt.Sprintf(
						"type %s does not exist", tuple_t)) }
					var schema, ok = t.(kmd.TupleSchema)
					if !(ok) { return errors.New(fmt.Sprintf(
						"type %s is not a tuple type", tuple_t)) }
					var schema_size = uint(len(schema.Elements))
					if schema_size != size { return errors.New(fmt.Sprintf(
						"tuple size not matching: given %d, require %d",
						size, schema_size)) }
					return nil
				},
				GetElementType: func(tuple_t kmd.TypeId, i uint) *kmd.Type {
					var t, t_exists = conf.SchemaTable[tuple_t]
					if !(t_exists) { panic("something went wrong") }
					var schema = t.(kmd.TupleSchema)
					return schema.Elements[i]
				},
				CreateTuple: func(tuple_t kmd.TypeId) kmd.Object {
					var t, exists = conf.SchemaTable[tuple_t]
					if !(exists) { panic("something went wrong") }
					var schema = t.(kmd.TupleSchema)
					var size = len(schema.Elements)
					return &ValProd {
						Elements: make([] Value, size),
					}
				},
				FillElement: func(tuple kmd.Object, i uint, value kmd.Object) {
					var tuple_pv = tuple.(ProductValue)
					tuple_pv.Elements[i] = value
				},
				FinishTuple: func(tuple kmd.Object, t kmd.TypeId) (kmd.Object, error) {
					var tuple_pv = tuple.(ProductValue)
					if len(tuple_pv.Elements) == 0 {
						tuple = nil
					} else if len(tuple_pv.Elements) == 1 {
						tuple = tuple_pv.Elements[0]
					}
					err := validate(tuple, t)
					if err != nil { return nil, err }
					return tuple, nil
				},
				Case2Enum: func(obj kmd.Object, enum_tid kmd.TypeId, case_tid kmd.TypeId) (kmd.Object, error) {
					var enum_t, exists = conf.SchemaTable[enum_tid]
					if !(exists) { return nil, errors.New(fmt.Sprintf(
						"type %s does not exist", enum_tid)) }
					var schema, ok = enum_t.(kmd.EnumSchema)
					if !(ok) { return nil, errors.New(fmt.Sprintf(
						"type %s is not a enum type", enum_tid)) }
					var index, is_case = schema.CaseIndexMap[case_tid]
					if !(is_case) { return nil, errors.New(fmt.Sprintf(
						"type %s is not a case type of the enum type %s",
						case_tid, enum_tid)) }
					if !(index < ProductMaxSize) {
						panic("something went wrong")
					}
					return &ValSum {
						Index: Short(index),
						Value: obj,
					}, nil
				},
			},
		},
	}
}


