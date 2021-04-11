package api

import (
	"os"
	"time"
	"runtime"
	"strings"
	"path/filepath"
	"kumachan/misc/rx"
	"kumachan/stdlib"
	"kumachan/runtime/lib/ui/qt"
	. "kumachan/lang"
	. "kumachan/runtime/lib/container"
)


type Locale struct {
	Language     String
	TimeZone     *time.Location
	Alternative  SumValue
}
func GetSystemLanguage() string {
	if runtime.GOOS == "windows" {
		// TODO: get system language from Windows API
		return "C"
	} else {
		var strip_encoding = func(value string) string {
			var t = strings.Split(value, ".")
			if len(t) > 0 {
				return t[0]
			} else {
				return "C"
			}
		}
		lc_all, exists := os.LookupEnv("LC_ALL")
		if exists { return strip_encoding(lc_all) }
		lang, exists := os.LookupEnv("LANG")
		if exists { return strip_encoding(lang) }
		return "C"
	}
}
var cwd = (func() stdlib.Path {
	var wd, err = os.Getwd()
	if err != nil { panic("unable to get current working directory") }
	return stdlib.ParsePath(wd)
})()

var OS_Constants = map[string] NativeConstant {
	"os::PlatformInfo": func(h InteropContext) Value {
		return Tuple(
			StringFromGoString(runtime.GOOS),
			StringFromGoString(runtime.GOARCH),
			ToBool(uint64(^uintptr(0)) == ^uint64(0)),
		)
	},
	"os::Cwd": func(h InteropContext) Value {
		return cwd
	},
	"os::Env": func(h InteropContext) Value {
		return GetEnv(h.GetSysEnv())
	},
	"os::Args": func(h InteropContext) Value {
		return GetArgs(h.GetSysArgs())
	},
	"os::Stdin": func(h InteropContext) Value {
		return h.GetStdIO().Stdin
	},
	"os::Stdout": func(h InteropContext) Value {
		return h.GetStdIO().Stdout
	},
	"os::Stderr": func(h InteropContext) Value {
		return h.GetStdIO().Stderr
	},
	"os::Locale": func(h InteropContext) Value {
		var locale = Locale {
			Language:    StringFromGoString(GetSystemLanguage()),
			TimeZone:    time.Local,
			Alternative: None(),
		}
		return Struct2Prod(locale)
	},
	"os::EntryModulePath": func(h InteropContext) Value {
		return stdlib.ParsePath(h.GetEntryModulePath())
	},
	"os::EntryModuleDirPath": func(h InteropContext) Value {
		var p = h.GetEntryModulePath()
		return stdlib.ParsePath(filepath.Dir(p))
	},
}

func GetEnv(raw ([] string)) Map {
	var m = NewMapOfStringKey()
	for _, item := range raw {
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

func GetArgs(raw ([] string)) ([] String) {
	var args = make([] String, len(raw))
	for i, raw_arg := range raw {
		args[i] = StringFromGoString(raw_arg)
	}
	return args
}

var OS_Functions = map[string] Value {
	"String from Path": func(path stdlib.Path) String {
		return StringFromGoString(path.String())
	},
	"parse-path": func(str String) stdlib.Path {
		return stdlib.ParsePath(GoStringFromString(str))
	},
	"path-join": func(path stdlib.Path, raw Value) stdlib.Path {
		var segments = ListFrom(raw)
		return path.Join(segments.CopyAsGoStrings())
	},
	"walk-dir": func(dir stdlib.Path) rx.Observable {
		return rx.WalkDir(dir.String()).Map(func(val rx.Object) rx.Object {
			var item = val.(rx.FileItem)
			return Tuple(stdlib.ParsePath(item.Path), Struct2Prod(item.State))
		})
	},
	"list-dir": func(dir stdlib.Path) rx.Observable {
		return rx.ListDir(dir.String()).Map(func(val rx.Object) rx.Object {
			var item = val.(rx.FileItem)
			return Tuple(stdlib.ParsePath(item.Path), Struct2Prod(item.State))
		})
	},
	"open-read-only": func(path stdlib.Path) rx.Observable {
		return rx.OpenReadOnly(path.String())
	},
	"open-read-write": func(path stdlib.Path) rx.Observable {
		return rx.OpenReadWrite(path.String())
	},
	"open-read-write-create": func(path stdlib.Path) rx.Observable {
		return rx.OpenReadWriteCreate(path.String(), 0666)
	},
	"open-overwrite": func(path stdlib.Path) rx.Observable {
		return rx.OpenOverwrite(path.String(), 0666)
	},
	"open-append": func(path stdlib.Path) rx.Observable {
		return rx.OpenAppend(path.String(), 0666)
	},
	"file-close": func(f rx.File) rx.Observable {
		return f.Close()
	},
	"file-get-state": func(f rx.File) rx.Observable {
		return f.State().Map(func(state rx.Object) rx.Object {
			return Struct2Prod(state)
		})
	},
	"file-read": func(f rx.File, amount uint) rx.Observable {
		return f.Read(amount)
	},
	"file-write": func(f rx.File, data ([] byte)) rx.Observable {
		return f.Write(data)
	},
	"file-seek-start": func(f rx.File, offset uint64) rx.Observable {
		return f.SeekStart(offset)
	},
	"file-seek-forward": func(f rx.File, offset uint64) rx.Observable {
		return f.SeekForward(offset)
	},
	"file-seek-backward": func(f rx.File, offset uint64) rx.Observable {
		return f.SeekBackward(offset)
	},
	"file-seek-end": func(f rx.File, offset uint64) rx.Observable {
		return f.SeekEnd(offset)
	},
	"file-read-char": func(f rx.File) rx.Observable {
		return f.ReadChar()
	},
	"file-write-char": func(f rx.File, char Char) rx.Observable {
		return f.WriteChar(rune(char))
	},
	"file-read-string": func(f rx.File) rx.Observable {
		return f.ReadRunes().Map(func(line rx.Object) rx.Object {
			return StringFromRuneSlice(line.([] rune))
		})
	},
	"file-write-string": func(f rx.File, str String) rx.Observable {
		return f.WriteString(GoStringFromString(str))
	},
	"file-read-line": func(f rx.File) rx.Observable {
		return f.ReadLineRunes().Map(func(line rx.Object) rx.Object {
			return StringFromRuneSlice(line.([] rune))
		})
	},
	"file-write-line": func(f rx.File, line String) rx.Observable {
		return f.WriteLine(GoStringFromString(line))
	},
	"file-read-lines": func(f rx.File) rx.Observable {
		return f.ReadLinesRuneSlices().Map(func(runes rx.Object) rx.Object {
			return StringFromRuneSlice(runes.([] rune))
		})
	},
	"file-read-all": func(f rx.File) rx.Observable {
		return f.ReadAll()
	},
	"exit": func(code uint8) rx.Observable {
		return rx.NewSync(func() (rx.Object, bool) {
			qt.Quit(func() {
				os.Exit(int(code))
			})
			panic("process should have exited")
		})
	},
}
