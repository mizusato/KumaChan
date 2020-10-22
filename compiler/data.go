package compiler

import (
	"fmt"
	"reflect"
	"strings"
	"encoding/base64"
	ch "kumachan/checker"
	c "kumachan/runtime/common"
	"unicode/utf8"
)


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

type DataString struct { Value  [] uint32 }
func (d DataString) ToValue() c.Value {
	return d.Value
}
func (d DataString) String() string {
	var b64 = CharsToBase64String(d.Value)
	return fmt.Sprintf("STRING %s", b64)
}

type DataStringFormatter ch.StringFormatter
func (d DataStringFormatter) ToValue() c.Value {
	var format_slice = func(args []c.Value) ([] uint32) {
		var buf = make([] uint32, 0)
		for i, seg := range d.Segments {
			buf = append(buf, seg...)
			if uint(i) < d.Arity {
				var chars = args[i].([] uint32)
				buf = append(buf, chars...)
			}
		}
		return buf
	}
	var f interface{}
	if d.Arity == 0 {
		f = func() ([] uint32) {
			return format_slice([]c.Value {})
		}
	} else if d.Arity == 1 {
		f = func(arg c.Value) ([] uint32) {
			return format_slice([]c.Value { arg })
		}
	} else {
		f = func(arg c.ProductValue) ([] uint32) {
			return format_slice(arg.Elements)
		}
	}
	return c.NativeFunctionValue(c.AdaptNativeFunction(f))
}
func (d DataStringFormatter) String() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "FORMAT %d ", d.Arity)
	for i, item := range d.Segments {
		buf.WriteString(CharsToBase64String(item))
		if i != len(d.Segments)-1 {
			buf.WriteString(" ")
		}
	}
	return buf.String()
}

type DataArrayInfo c.ArrayInfo
func (d DataArrayInfo) ToValue() c.Value {
	return c.ArrayInfo(d)
}
func (d DataArrayInfo) String() string {
	return fmt.Sprintf("ARRAY %d %s", d.Length, d.ItemType.String())
}

func CharsToBase64String(chars ([] uint32)) string {
	var buf strings.Builder
	var encoder = base64.NewEncoder(base64.StdEncoding, &buf)
	var chunk ([4] byte)
	for _, char := range chars {
		var size = utf8.EncodeRune(chunk[:], rune(char))
		var _, err = encoder.Write(chunk[:size])
		if err != nil { panic(err) }
	}
	var err = encoder.Close()
	if err != nil { panic(err) }
	return buf.String()
}
