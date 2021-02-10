package kmd

import (
	"io"
	"fmt"
	"strconv"
	"encoding/base64"
)


type serializeContext struct {
	*Serializer
	Key           string
	Depth         uint
	OmitType      bool
	OmitAllTypes  bool
}
const header = "KumaChan Data"
const omittedType = "-"

func Serialize(root Object, serializer *Serializer, output io.Writer) error {
	var ctx = serializeContext {
		Serializer:   serializer,
		Key:          "",
		Depth:        0,
		OmitType:     false,
		OmitAllTypes: false,
	}
	_, err := fmt.Fprintln(output, header)
	if err != nil { return err }
	return serialize(root, ctx, output)
}

func serialize(obj Object, ctx serializeContext, output io.Writer) error {
	var t = ctx.DetermineType(obj)
	err := writeIndent(output, ctx.Depth)
	if err != nil { return err }
	_, err = fmt.Fprintln(output, (func() string {
		if ctx.OmitType || ctx.OmitAllTypes {
			return omittedType
		} else {
			if ctx.Key != "" {
				return (ctx.Key + " " + t.String())
			} else {
				return t.String()
			}
		}
	})())
	if err != nil { return err }
	switch t.kind {
	case Bool:
		val := ctx.WriteBool(obj)
		if val {
			return writePrimitive(output, "true", ctx.Depth)
		} else {
			return writePrimitive(output, "false", ctx.Depth)
		}
	case Float:
		return writePrimitive(output, ctx.WriteFloat(obj), ctx.Depth)
	case Uint32:
		return writePrimitive(output, ctx.WriteUint32(obj), ctx.Depth)
	case Int32:
		return writePrimitive(output, ctx.WriteInt32(obj), ctx.Depth)
	case Uint64:
		return writePrimitive(output, ctx.WriteUint64(obj), ctx.Depth)
	case Int64:
		return writePrimitive(output, ctx.WriteInt64(obj), ctx.Depth)
	case Int:
		return writePrimitive(output, ctx.WriteInt(obj), ctx.Depth)
	case String:
		var str = ctx.WriteString(obj)
		return writePrimitive(output, strconv.Quote(str), ctx.Depth)
	case Binary:
		err := writeIndent(output, (ctx.Depth + 1))
		if err != nil { return err }
		var bin = ctx.WriteBinary(obj)
		var encoder = base64.NewEncoder(base64.StdEncoding, output)
		_, err = encoder.Write(bin)
		if err != nil { return err }
		err = encoder.Close()
		if err != nil { return err }
		return nil
	case Array:
		return ctx.IterateArray(obj, func(i uint, item Object) error {
			var item_ctx = serializeContext {
				Serializer:   ctx.Serializer,
				Key:          ctx.Key,
				Depth:        (ctx.Depth + 1),
				OmitType:     true,
				OmitAllTypes: (ctx.OmitAllTypes || (i > 0)),
			}
			return serialize(item, item_ctx, output)
		})
	case Optional:
		var inner, exists = ctx.UnwrapOptional(obj)
		if exists {
			var inner_ctx = serializeContext {
				Serializer:   ctx.Serializer,
				Key:          ctx.Key,
				Depth:        (ctx.Depth + 1),
				OmitType:     true,
				OmitAllTypes: ctx.OmitAllTypes,
			}
			return serialize(inner, inner_ctx, output)
		} else {
			return nil
		}
	case Record:
		return ctx.IterateRecord(obj, func(key string, value Object) error {
			var entry_ctx = serializeContext {
				Serializer:   ctx.Serializer,
				Key:          key,
				Depth:        (ctx.Depth + 1),
				OmitType:     false,
				OmitAllTypes: ctx.OmitAllTypes,
			}
			return serialize(value, entry_ctx, output)
		})
	case Tuple:
		return ctx.IterateTuple(obj, func(_ uint, element Object) error {
			var element_ctx = serializeContext {
				Serializer:   ctx.Serializer,
				Key:          "",
				Depth:        (ctx.Depth + 1),
				OmitType:     false,
				OmitAllTypes: ctx.OmitAllTypes,
			}
			return serialize(element, element_ctx, output)
		})
	case Enum:
		var case_ctx = serializeContext {
			Serializer:   ctx.Serializer,
			Key:          "",
			Depth:        (ctx.Depth + 1),
			OmitType:     false,
			OmitAllTypes: false,
		}
		return serialize(ctx.Enum2Case(obj), case_ctx, output)
	default:
		panic("impossible branch")
	}
}

var indent = [] byte { byte(rune(' ')) }
func writeIndent(output io.Writer, depth uint) error {
	for i := uint(0); i < depth; i += 1 {
		_, err := output.Write(indent)
		if err != nil { return err }
	}
	return nil
}
func writePrimitive(output io.Writer, content interface{}, depth uint) error {
	err := writeIndent(output, (depth + 1))
	if err != nil { return err }
	_, err = fmt.Fprintln(output, content)
	if err != nil { return err }
	return nil
}