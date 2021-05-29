package kmd

import (
	"io"
	"fmt"
	"bufio"
	"errors"
	"strings"
	"strconv"
	"math/big"
	"io/ioutil"
	"kumachan/standalone/util"
	"encoding/base64"
)


type deserializeContext struct {
	*Deserializer
	Depth        uint
	RequireKey   bool
	ReturnKey    *string
	ReturnType   **Type
	TypesInfo    *([] omittedTypeInfo)
	TypesCursor  *uint
}
type omittedTypeInfo struct {
	Key   string
	Type  *Type
}
type deserializeReader struct {
	*bufio.Reader
	unreadIndention  uint
	linesRead        uint
}

func Deserialize(input io.Reader, deserializer *Deserializer) (Object, *Type, error) {
	var t *Type
	var ctx = deserializeContext {
		Deserializer: deserializer,
		Depth:        0,
		RequireKey:   false,
		ReturnKey:    nil,
		ReturnType:   &t,
		TypesInfo:    nil,
		TypesCursor:  nil,
	}
	var line string
	_, err := util.WellBehavedFscanln(input, &line)
	if err != nil { return nil, nil, err }
	line = strings.TrimRight(line, "\r")
	if line != header { return nil, nil, errors.New("invalid header") }
	var reader = &deserializeReader {
		Reader: bufio.NewReader(input),
		unreadIndention: 0,
		linesRead: 1,
	}
	obj, err := deserialize(reader, ctx)
	var n = (reader.linesRead + 1)
	if err != nil { return nil, nil,
		fmt.Errorf("error near line %d: %w", n, err) }
	if t == nil { panic("something went wrong") }
	return obj, t, nil
}

func deserialize(input *deserializeReader, ctx deserializeContext) (Object, error) {
	var line string
	_, err := readLine(input, &line)
	if err != nil { return nil, err }
	var t *Type
	var key string
	if ctx.TypesInfo != nil && ctx.TypesCursor != nil &&
		*ctx.TypesCursor < uint(len(*ctx.TypesInfo)) {
		if line != omittedType { return nil, errors.New("type should be omitted") }
		var info = (*ctx.TypesInfo)[*ctx.TypesCursor]
		*ctx.TypesCursor += 1
		t = info.Type
		key = info.Key
		if (ctx.RequireKey && key == "") ||
			(!(ctx.RequireKey) && key != "") {
			panic("something went wrong")
		}
	} else {
		var type_str string
		if ctx.RequireKey {
			var key_part, type_part = stringSplitFirstSegment(line)
			if key_part == "" { return nil, errors.New("missing field name") }
			key = key_part
			type_str = type_part
		} else {
			type_str = line
		}
		var parsed_type, ok = TypeParse(type_str)
		if !(ok) { return nil, errors.New(fmt.Sprintf("invalid type: %s", line)) }
		t = parsed_type
		if ctx.TypesInfo != nil &&
			(ctx.TypesCursor == nil ||
				*ctx.TypesCursor == uint(len(*ctx.TypesInfo))) {
			*ctx.TypesInfo = append(*ctx.TypesInfo, omittedTypeInfo{
				Key:  key,
				Type: t,
			})
			*ctx.TypesCursor += 1
		}
	}
	if ctx.ReturnKey != nil {
		if key == "" { panic("something went wrong") }
		*ctx.ReturnKey = key
	}
	if ctx.ReturnType != nil {
		*ctx.ReturnType = t
	}
	switch t.kind {
	case Bool:
		return readPrimitive(input, ctx.Depth, func(str string) (Object, error) {
			switch str {
			case "true":  return ctx.ReadBool(true), nil
			case "false": return ctx.ReadBool(false), nil
			default:      return nil, errors.New("invalid bool")
			}
		})
	case Float:
		return readPrimitive(input, ctx.Depth, func(str string) (Object, error) {
			value, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid float: %w", err)
			}
			if !(util.IsNormalFloat(value)) {
				return nil, errors.New("invalid float")
			}
			return ctx.ReadFloat(value), nil
		})
	case Complex:
		return readPrimitive(input, ctx.Depth, func(str string) (Object, error) {
			value, err := strconv.ParseComplex(str, 128)
			if err != nil {
				return nil, fmt.Errorf("invalid complex: %w", err)
			}
			if !(util.IsNormalComplex(value)) {
				return nil, errors.New("invalid complex")
			}
			return ctx.ReadComplex(value), nil
		})
	case Integer:
		return readPrimitive(input, ctx.Depth, func(str string) (Object, error) {
			var value big.Int
			var _, ok = value.SetString(str, 10)
			if !(ok) { return nil, errors.New("invalid int") }
			var obj, fit = ctx.ReadInteger(&value)
			if fit {
				return obj, nil
			} else {
				return nil, errors.New("integer too big")
			}
		})
	case String:
		return readPrimitive(input, ctx.Depth, func(str string) (Object, error) {
			value, err := strconv.Unquote(str)
			if err != nil { return nil, fmt.Errorf("invalid string: %w", err) }
			return ctx.ReadString(value), nil
		})
	case Binary:
		return readPrimitive(input, ctx.Depth, func(str string) (Object, error) {
			var str_reader = strings.NewReader(str)
			var decoder = base64.NewDecoder(base64.StdEncoding, str_reader)
			value, err := ioutil.ReadAll(decoder)
			if err != nil { return nil, fmt.Errorf("invalid binary: %w", err) }
			return ctx.ReadBinary(value), nil
		})
	case Array:
		var array = ctx.CreateArray(t)
		var types_info = [] omittedTypeInfo {
			{ Key: "", Type: t.elementType},
		}
		var types_cursor = uint(0)
		var i = 0
		for {
			n, err := readIndent(input)
			if err != nil { return nil, err }
			if n == (ctx.Depth + 1) {
				if i > 0 {
					types_cursor = 0
				}
				var item_ctx = deserializeContext {
					Deserializer: ctx.Deserializer,
					Depth:        (ctx.Depth + 1),
					TypesInfo:    &types_info,
					TypesCursor:  &types_cursor,
				}
				item, err := deserialize(input, item_ctx)
				if err != nil { return nil, err }
				ctx.AppendItem(&array, item)
				i += 1
			} else if n <= ctx.Depth {
				unreadIndent(input, n)
				return array, nil
			} else {
				return nil, errors.New("wrong indention")
			}
		}
	case Optional:
		var types_info = [] omittedTypeInfo{
			{ Key: "", Type: t.elementType},
		}
		var types_cursor = uint(0)
		n, err := readIndent(input)
		if err != nil { return nil, err }
		if n == (ctx.Depth + 1) {
			var inner_ctx = deserializeContext {
				Deserializer: ctx.Deserializer,
				Depth:        (ctx.Depth + 1),
				TypesInfo:    &types_info,
				TypesCursor:  &types_cursor,
			}
			inner, err := deserialize(input, inner_ctx)
			if err != nil { return nil, err }
			return ctx.Some(inner, t), nil
		} else if n <= ctx.Depth {
			unreadIndent(input, n)
			return ctx.Nothing(t), nil
		} else {
			return nil, errors.New("wrong indention")
		}
	case Record:
		var entries = make(map[string] Object)
		var types = make(map[string] *Type)
		for {
			n, err := readIndent(input)
			if err != nil { return nil, err }
			if n == (ctx.Depth + 1) {
				var key string
				var value_t *Type
				var entry_ctx = deserializeContext {
					Deserializer: ctx.Deserializer,
					Depth:        (ctx.Depth + 1),
					RequireKey:   true,
					ReturnKey:    &key,
					ReturnType:   &value_t,
					TypesInfo:    ctx.TypesInfo,
					TypesCursor:  ctx.TypesCursor,
				}
				value, err := deserialize(input, entry_ctx)
				if err != nil { return nil, err }
				var _, exists = entries[key]
				if exists { return nil, errors.New(fmt.Sprintf(
					"duplicate field %s", key))}
				entries[key] = value
				types[key] = value_t
			} else if n <= ctx.Depth {
				unreadIndent(input, n)
				var tid = t.identifier
				err := ctx.CheckRecord(tid, uint(len(entries)))
				if err != nil { return nil, err }
				var draft = ctx.CreateRecord(tid)
				for key, value := range entries {
					var value_t = types[key]
					field_t, field_index, err := ctx.GetFieldInfo(tid, key)
					if err != nil { return nil, err }
					adapted, err := ctx.AssignObject(value, value_t, field_t)
					if err != nil { return nil, err }
					ctx.FillField(draft, field_index, adapted)
				}
				record, err := ctx.FinishRecord(draft, tid)
				if err != nil { return nil, err }
				return record, nil
			} else {
				return nil, errors.New("wrong indention")
			}
		}
	case Tuple:
		var elements = make([] Object, 0)
		var types = make([] *Type, 0)
		for {
			n, err := readIndent(input)
			if err != nil { return nil, err }
			if n == (ctx.Depth + 1) {
				var value_t *Type
				var entry_ctx = deserializeContext {
					Deserializer: ctx.Deserializer,
					Depth:        (ctx.Depth + 1),
					ReturnType:   &value_t,
					TypesInfo:    ctx.TypesInfo,
					TypesCursor:  ctx.TypesCursor,
				}
				value, err := deserialize(input, entry_ctx)
				if err != nil { return nil, err }
				elements = append(elements, value)
				types = append(types, value_t)
			} else if n <= ctx.Depth {
				unreadIndent(input, n)
				var tid = t.identifier
				err := ctx.CheckTuple(tid, uint(len(elements)))
				if err != nil { return nil, err }
				var draft = ctx.CreateTuple(tid)
				for i, value := range elements {
					var value_t = types[i]
					el_t := ctx.GetElementType(tid, uint(i))
					adapted, err := ctx.AssignObject(value, value_t, el_t)
					if err != nil { return nil, err }
					ctx.FillElement(draft, uint(i), adapted)
				}
				tuple, err := ctx.FinishTuple(draft, tid)
				if err != nil { return nil, err }
				return tuple, nil
			} else {
				return nil, errors.New("wrong indention")
			}
		}
	case Enum:
		n, err := readIndent(input)
		if err != nil { return nil, err }
		if n == (ctx.Depth + 1) {
			var case_t *Type
			var case_ctx = deserializeContext {
				Deserializer: ctx.Deserializer,
				Depth:        (ctx.Depth + 1),
				ReturnType:   &case_t,
				TypesInfo:    nil,
				TypesCursor:  nil,
			}
			case_value, err := deserialize(input, case_ctx)
			if err != nil { return nil, err }
			var enum_tid = t.identifier
			var case_tid = case_t.identifier
			enum_value, err := ctx.Case2Enum(case_value, enum_tid, case_tid)
			if err != nil { return nil, err }
			return enum_value, nil
		} else {
			return nil, errors.New("wrong indention")
		}
	default:
		panic("impossible branch")
	}
}

func readIndent(input *deserializeReader) (uint, error) {
	var n = uint(0) + input.unreadIndention
	input.unreadIndention = 0
	for {
		char, err := input.ReadByte()
		if err == io.EOF { return 0, nil }
		if err != nil { return 0, err }
		if rune(char) == ' ' {
			n += 1
		} else {
			err := input.UnreadByte()
			if err != nil { panic("something went wrong") }
			return n, nil
		}
	}
}
func unreadIndent(input *deserializeReader, n uint) {
	input.unreadIndention += n
}
func readPrimitive(input *deserializeReader, depth uint, f func(string)(Object,error)) (Object,error) {
	n, err := readIndent(input)
	if err != nil { return nil, err }
	if n != (depth + 1) { return nil, errors.New("wrong indention") }
	var line string
	_, err = readLine(input, &line)
	if err != nil { return nil, err }
	return f(line)
}
func readLine(input *deserializeReader, to *string) (int, error) {
	n, err := util.WellBehavedFscanln(input, to)
	if err == nil {
		*to = strings.TrimRight(*to, "\r")
		input.linesRead += 1
	}
	return n, err
}

