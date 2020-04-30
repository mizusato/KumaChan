package lib

import (
	"os"
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
	var m = NewMap(func(v1 Value, v2 Value) Ordering {
		var s1 = v1.(String)
		var s2 = v2.(String)
		return StringCompare(s1, s2)
	})
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
		return strings.Split(string(str), __PathSep)
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
