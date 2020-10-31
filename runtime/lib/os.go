package lib

import (
	"os"
	"runtime"
	"strings"
	"kumachan/runtime/rx"
	. "kumachan/runtime/common"
	. "kumachan/runtime/lib/container"
)


type Path ([] string)
var __PathSep = string([]rune { os.PathSeparator })
func (path Path) String() string {
	return strings.Join(path, __PathSep)
}
func PathFrom(str string) Path {
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

var OS_Constants = map[string] NativeConstant {
	"OS::Kind":    func(h InteropContext) Value {
		return StringFromGoString(runtime.GOOS)
	},
	"OS::Arch":    func(h InteropContext) Value {
		return StringFromGoString(runtime.GOARCH)
	},
	"OS::Is64Bit": func(h InteropContext) Value {
		return ToBool(uint64(^uintptr(0)) == ^uint64(0))
	},
	"OS::Cwd":     func(h InteropContext) Value {
		var wd, err = os.Getwd()
		if err != nil { panic("unable to get current working directory") }
		return PathFrom(wd)
	},
	"OS::Env":     func(h InteropContext) Value {
		return GetEnv()
	},
	"OS::Args":    func(h InteropContext) Value {
		return GetArgs(h)
	},
	"OS::Stdin":   func(h InteropContext) Value {
		return rx.FileFrom(os.Stdin)
	},
	"OS::Stdout":  func(h InteropContext) Value {
		return rx.FileFrom(os.Stdout)
	},
	"OS::Stderr":  func(h InteropContext) Value {
		return rx.FileFrom(os.Stderr)
	},
}

func GetEnv() Map {
	var m = NewStrMap()
	for _, item := range os.Environ() {
		var str = StringFromGoString(item)
		var k = make(String, 0)
		var v = make(String, 0)
		var cut = false
		for _, r := range str {
			if !cut && r == '=' {
				cut = true
				continue
			}
			if cut {
				v = append(v, r)
			} else {
				k = append(k, r)
			}
		}
		m, _ = m.Inserted(k, v)
	}
	return m
}

func GetArgs(h InteropContext) ([] String) {
	var raw_args = h.GetArgs()
	var args = make([] String, len(raw_args))
	for i, raw := range raw_args {
		args[i] = StringFromGoString(raw)
	}
	return args
}

var OS_Functions = map[string] Value {
	"Path from String": func(str String) Path {
		return PathFrom(GoStringFromString(str))
	},
	"String from Path": func(path Path) String {
		return StringFromGoString(path.String())
	},
	"walk-dir": func(dir Path) rx.Effect {
		return rx.WalkDir(dir.String()).Map(func(val rx.Object) rx.Object {
			var item = val.(rx.FileItem)
			return ToTuple2(PathFrom(item.Path), Struct2Prod(item.State))
		})
	},
	"list-dir": func(dir Path) rx.Effect {
		return rx.ListDir(dir.String()).Map(func(val rx.Object) rx.Object {
			var item = val.(rx.FileItem)
			return ToTuple2(PathFrom(item.Path), Struct2Prod(item.State))
		})
	},
	"open-read-only": func(path Path) rx.Effect {
		return rx.OpenReadOnly(path.String())
	},
	"open-read-write": func(path Path) rx.Effect {
		return rx.OpenReadWrite(path.String())
	},
	"open-read-write-create": func(path Path) rx.Effect {
		return rx.OpenReadWriteCreate(path.String(), 0666)
	},
	"open-overwrite": func(path Path) rx.Effect {
		return rx.OpenOverwrite(path.String(), 0666)
	},
	"open-append": func(path Path) rx.Effect {
		return rx.OpenAppend(path.String(), 0666)
	},
	"file-close": func(f rx.File) rx.Effect {
		return f.Close()
	},
	"file-get-state": func(f rx.File) rx.Effect {
		return f.State().Map(func(state rx.Object) rx.Object {
			return Struct2Prod(state)
		})
	},
	"file-read": func(f rx.File, amount uint) rx.Effect {
		return f.Read(amount)
	},
	"file-write": func(f rx.File, data ([] byte)) rx.Effect {
		return f.Write(data)
	},
	"file-seek-start": func(f rx.File, offset uint64) rx.Effect {
		return f.SeekStart(offset)
	},
	"file-seek-forward": func(f rx.File, offset uint64) rx.Effect {
		return f.SeekForward(offset)
	},
	"file-seek-backward": func(f rx.File, offset uint64) rx.Effect {
		return f.SeekBackward(offset)
	},
	"file-seek-end": func(f rx.File, offset uint64) rx.Effect {
		return f.SeekEnd(offset)
	},
	"file-read-char": func(f rx.File) rx.Effect {
		return f.ReadChar()
	},
	"file-write-char": func(f rx.File, char Char) rx.Effect {
		return f.WriteChar(rune(char))
	},
	"file-read-string": func(f rx.File) rx.Effect {
		return f.ReadRunes().Map(func(line rx.Object) rx.Object {
			return StringFromRuneSlice(line.([] rune))
		})
	},
	"file-write-string": func(f rx.File, str String) rx.Effect {
		return f.WriteString(GoStringFromString(str))
	},
	"file-read-line": func(f rx.File) rx.Effect {
		return f.ReadLineRunes().Map(func(line rx.Object) rx.Object {
			return StringFromRuneSlice(line.([] rune))
		})
	},
	"file-write-line": func(f rx.File, line String) rx.Effect {
		return f.WriteLine(GoStringFromString(line))
	},
	"file-read-lines": func(f rx.File) rx.Effect {
		return f.ReadLinesRuneSlices().Map(func(runes rx.Object) rx.Object {
			return StringFromRuneSlice(runes.([] rune))
		})
	},
	"file-read-all": func(f rx.File) rx.Effect {
		return f.ReadAll()
	},
	"exit": func(code uint8) rx.Effect {
		return rx.NewSync(func() (rx.Object, bool) {
			os.Exit(int(code))
			panic("process should have exited")
		})
	},
}
