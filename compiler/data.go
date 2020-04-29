package compiler

import (
	"fmt"
	"reflect"
	"strings"
	"encoding/base64"
	ch "kumachan/checker"
	c "kumachan/runtime/common"
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
	var format_slice = func(args []c.Value) []rune {
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
	var f interface{}
	if d.Arity == 0 {
		f = func() []rune {
			return format_slice([]c.Value {})
		}
	} else if d.Arity == 1 {
		f = func(arg c.Value) []rune {
			return format_slice([]c.Value { arg })
		}
	} else {
		f = func(arg c.ProductValue) []rune {
			return format_slice(arg.Elements)
		}
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

type DataArrayInfo c.ArrayInfo
func (d DataArrayInfo) ToValue() c.Value {
	return c.ArrayInfo(d)
}
func (d DataArrayInfo) String() string {
	return fmt.Sprintf("ARRAY %d %s", d.Length, d.ItemType.String())
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
