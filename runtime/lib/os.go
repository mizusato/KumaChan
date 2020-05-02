package lib

import (
	"os"
	"time"
	"runtime"
	"strings"
	"kumachan/runtime/common/rx"
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

var OS_Constants = map[string] Value {
	"OS::Kind":    String(runtime.GOOS),
	"OS::Arch":    String(runtime.GOARCH),
	"OS::Is64Bit": ToBool(uint64(^uintptr(0)) == ^uint64(0)),
	"OS::Env":     GetEnv(),
	"OS::Args":    GetArgs(),
	"OS::Stdin":   rx.FileFrom(os.Stdin),
	"OS::Stdout":  rx.FileFrom(os.Stdout),
	"OS::Stderr":  rx.FileFrom(os.Stderr),
}

func GetEnv() Map {
	var m = NewStrMap()
	for _, item := range os.Environ() {
		var str = String(item)
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
		m = m.Insert(k, v)
	}
	return m
}

func GetArgs() ([] String) {
	// TODO: further process may be useful to extinguish interpreter arguments
	var args = make([] String, len(os.Args))
	for i, raw := range os.Args {
		args[i] = ([] rune)(raw)
	}
	return args
}

var OS_Functions = map[string] Value {
	"Path from String": func(str String) Path {
		return PathFrom(string(str))
	},
	"String from Path": func(path Path) String {
		return String(path.String())
	},
	"walk-dir": func(dir Path) rx.Effect {
		return rx.WalkDir(dir.String()).Map(func(val rx.Object) rx.Object {
			var item = val.(rx.FileItem)
			return ToTuple2(PathFrom(item.Path), item.Info)
		})
	},
	"list-dir": func(dir Path) rx.Effect {
		return rx.ListDir(dir.String()).Map(func(val rx.Object) rx.Object {
			var item = val.(rx.FileItem)
			return ToTuple2(PathFrom(item.Path), item.Info)
		})
	},
	"file-state-get-name": func(state os.FileInfo) String {
		return String(state.Name())
	},
	"file-state-get-size": func(state os.FileInfo) uint64 {
		return uint64(state.Size())
	},
	"file-state-get-mode": func(state os.FileInfo) os.FileMode {
		return state.Mode()
	},
	"file-state-get-is-dir": func(state os.FileInfo) SumValue {
		return ToBool(state.IsDir())
	},
	"file-state-get-last-modified": func(state os.FileInfo) time.Time {
		return state.ModTime()
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
		return f.State()
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
	"file-write-char": func(f rx.File, char rune) rx.Effect {
		return f.WriteChar(char)
	},
	"file-read-line": func(f rx.File) rx.Effect {
		return f.ReadLine().Map(func(line rx.Object) rx.Object {
			return ([] rune)(line.(string))
		})
	},
	"file-write-line": func(f rx.File, line ([] rune)) rx.Effect {
		return f.WriteLine(string(line))
	},
	"exit": func(code uint8) rx.Effect {
		return rx.CreateBlockingEffect(func() (rx.Object, bool) {
			os.Exit(int(code))
			panic("process should have exited")
		})
	},
}
