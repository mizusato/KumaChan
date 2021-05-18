package stdlib

import (
	"os"
	"fmt"
	"bytes"
	"image"
	"strings"
	"reflect"
	"image/png"
	"path/filepath"
	"math/big"
)


/* IMPORTANT: this go file should be consistent with corresponding km files */
var __ModuleDirectories = [] string {
	"core", "time", "l10n",
	"io", "os", "json", "net", "rpc", "image", "ui",
}
func GetModuleDirectoryNames() ([] string) { return __ModuleDirectories }
func GetDirectoryPath() string {
	var exe_path, err = os.Executable()
	if err != nil { panic(err) }
	var exe_dir = filepath.Dir(exe_path)
	var stdlib_dir = filepath.Join(exe_dir, "stdlib")
	return stdlib_dir
}

const Mod_core = "core"
const Mod_ui = "ui"
var core_types = []string {
	// types.km
	Bool, Yes, No,
	Maybe, Some, None,
	Result, Success, Failure,
	Ordering, Smaller, Equal, Bigger,
	Optional,
	// numeric.km
	Integer, Number,
	Float, NormalFloat,
	Complex, NormalComplex,
	// error.km
	Error,
	// binary.km
	Bit, Byte, Word, Dword, Qword, Bytes,
	// containers.km
	Seq, List, Heap, Set, Map, FlexList, FlexListKey,
	// rx.km
	Observable,
	Async, Sync,
	Source, Computed,
	Sink, Bus, Reactive, ReactiveEntity,
	ReactiveSnapshots, Mutex,
	Mutable, Buffer, HashMap,
	// string.km
	Char, String, HardCodedString,
}
func GetCoreScopedSymbols() []string {
	var list = make([]string, 0)
	list = append(list, core_types...)
	return list
}

// types.km
const Bool = "Bool"
const Yes = "Yes"
const No = "No"
const ( YesIndex = iota; NoIndex )
const Maybe = "Maybe"
const Some = "Some"
const None = "None"
const ( SomeIndex = iota; NoneIndex )
const Result = "Result"
const Success = "Success"
const Failure = "Failure"
const ( SuccessIndex = iota; FailureIndex )
const Ordering = "Ordering"
const Smaller = "<<"
const Equal = "=="
const Bigger = ">>"
const ( SmallerIndex = iota; EqualIndex; BiggerIndex )
const Optional = "Optional"
// error.km
const Error = "Error"
// binary.km
const Bit = "Bit"
const Byte = "Byte"
const Word = "Word"
const Dword = "Dword"
const Qword = "Qword"
const Bytes = "Bytes"
// numeric.km
const Number = "Number"
const Integer = "Integer"
const Float = "Float"
const NormalFloat = "NormalFloat"
const Complex = "Complex"
const NormalComplex = "NormalComplex"
// containers.km
const Seq = "Seq"
const List = "List"
const Heap = "Heap"
const Set = "Set"
const Map = "Map"
const FlexList = "FlexList"
const FlexListKey = "FlexListKey"
// rx.km
const Observable = "Observable"
const Async = "Async"
const Sync = "Sync"
const Source = "Source"
const Computed = "Computed"
const Sink = "Sink"
const Bus = "Bus"
const Reactive = "Reactive"
const ReactiveEntity = "ReactiveEntity"
const ReactiveSnapshots = "ReactiveSnapshots"
const Mutex = "Mutex"
const Mutable = "Mutable"
const Buffer = "Buffer"
const HashMap = "HashMap"
// string.km
const Char = "Char"
const String = "String"
const HardCodedString = "HardCodedString"

// ui
const AssetFile_T = "AssetFile"
type AssetFile struct {
	Path  string
}

func GetPrimitiveReflectType(name string) (reflect.Type, bool) {
	switch name {
	case Integer, Number:
		return reflect.TypeOf(big.NewInt(0)), true
	case Float, NormalFloat:
		return reflect.TypeOf(float64(0)), true
	case Complex, NormalComplex:
		return reflect.TypeOf(complex128(complex(0,1))), true
	case Bit:
		return reflect.TypeOf(true), true
	case Byte:
		return reflect.TypeOf(uint8(0)), true
	case Word:
		return reflect.TypeOf(uint16(0)), true
	case Dword:
		return reflect.TypeOf(uint32(0)), true
	case Qword:
		return reflect.TypeOf(uint64(0)), true
	case Char:
		return reflect.TypeOf(int32(0)), true
	default:
		return nil, false
	}
}

// rpc_service_template.km
const ServiceInstanceType = "Instance"
const ServiceArgumentType = "Argument"
const ServiceMethodsType = "Methods"
const ServiceIdentifierConst = "Identifier"
const ServiceCreateFunction = "create"

// path
type Path ([] string)
var __PathSep = string([] rune { os.PathSeparator })
func (path Path) String() string {
	return strings.Join(path, __PathSep)
}
func (path Path) Join(segments ([] string)) Path {
	for _, seg := range segments {
		if strings.Contains(seg, __PathSep) {
			panic(fmt.Sprintf("invalid path segment %s", seg))
		}
	}
	var new_path = make(Path, (len(path) + len(segments)))
	copy(new_path, path)
	copy(new_path[len(path):], segments)
	return new_path
}
func ParsePath(str string) Path {
	var raw = strings.Split(str, __PathSep)
	var path = make([] string, 0, len(raw))
	for i, segment := range raw {
		if i != 0 && segment == "" {
			continue
		}
		path = append(path, segment)
	}
	return path
}

// loader types
// image
const Image_M = "Image"
type Image interface { GetPixelData() image.Image }
// image:raw
const RawImage_T = "RawImage"
type RawImage struct {
	Data  image.Image
}
func (img *RawImage) GetPixelData() image.Image { return img.Data }
// image:png
const PNG_T = "PNG"
type PNG struct {
	Data  [] byte
}
func (img *PNG) GetPixelData() image.Image {
	var reader = bytes.NewReader(img.Data)
	var decoded, err = png.Decode(reader)
	if err != nil { panic(fmt.Errorf("failed to decode png data: %w", err)) }
	return decoded
}


// ui
var qtWidgetTypeNameMap = map[string] string {
	"QWidget":        "Widget",
	"QMainWindow":    "Window",
	"QWebView":       "PlainWebView",
	"WebView":        "WebView",
	"QLabel":         "NativeLabel",
	"QLineEdit":      "NativeInput",
	"QPlainTextEdit": "NativeInputMultiLine",
	"QPushButton":    "NativeButton",
	"QCheckBox":      "NativeCheckbox",
	"QComboBox":      "NativeSelect",
	"QListWidget":    "NativeList",
}
const QtActionType = "Action"
func GetQtWidgetTypeName(widget_name string) (string, bool) {
	var type_name, exists = qtWidgetTypeNameMap[widget_name]
	return type_name, exists
}

