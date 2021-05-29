package generator

import (
	"fmt"
	"reflect"
	"strings"
	"strconv"
	"kumachan/interpreter/base"
	ch "kumachan/interpreter/compiler/checker"
)


type DataInteger ch.IntegerLiteral
func (d DataInteger) ToValue() base.Value {
	return d.Value
}
func (d DataInteger) String() string {
	return fmt.Sprintf("BIG %s", d.Value.String())
}

type DataSmallInteger ch.SmallIntLiteral
func (d DataSmallInteger) ToValue() base.Value {
	return d.Value
}
func (d DataSmallInteger) String() string {
	return fmt.Sprintf("SMALL %s %v", reflect.TypeOf(d.Value).String(), d.Value)
}

type DataFloat ch.FloatLiteral
func (d DataFloat) ToValue() base.Value {
	return d.Value
}
func (d DataFloat) String() string {
	return fmt.Sprintf("FLOAT %f", d.Value)
}

type DataString struct { Value string }
func (d DataString) ToValue() base.Value {
	return d.Value
}
func (d DataString) String() string {
	return fmt.Sprintf("STRING %s", strconv.Quote(d.Value))
}

type DataStringFormatter ch.StringFormatter
func (d DataStringFormatter) ToValue() base.Value {
	var format_slice = func(args ([] base.Value)) string {
		var buf strings.Builder
		for i, seg := range d.Segments {
			buf.WriteString(seg)
			if uint(i) < d.Arity {
				var item = args[i].(string)
				buf.WriteString(item)
			}
		}
		return buf.String()
	}
	var f interface{}
	if d.Arity == 0 {
		f = func() string {
			return format_slice([] base.Value {})
		}
	} else if d.Arity == 1 {
		f = func(arg base.Value) string {
			return format_slice([] base.Value { arg })
		}
	} else {
		f = func(arg base.TupleValue) string {
			return format_slice(arg.Elements)
		}
	}
	return base.ValNativeFun(base.AdaptNativeFunction(f))
}
func (d DataStringFormatter) String() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "FORMAT %d ", d.Arity)
	for i, item := range d.Segments {
		buf.WriteString(strconv.Quote(item))
		if i != len(d.Segments)-1 || uint(len(d.Segments)) == d.Arity {
			buf.WriteString(fmt.Sprintf("$%d", i))
		}
	}
	return buf.String()
}

type DataArrayInfo base.ArrayInfo
func (d DataArrayInfo) ToValue() base.Value {
	return base.ArrayInfo(d)
}
func (d DataArrayInfo) String() string {
	return fmt.Sprintf("ARRAY %d %s", d.Length, d.ItemType.String())
}

