package loader

import (
	"os"
	"fmt"
	"path/filepath"
	"kumachan/stdlib"
)


const StdlibFolder = "stdlib"

var __StdLibModules = stdlib.GetModuleDirectories()
var __StdLibIndex = make(Index)
var __StdLibResIndex = make(ResIndex)
func ImportStdLib (imp_map (map[string] *Module), imp_set (map[string] bool)) {
	for name, mod := range __StdLibIndex {
		imp_map[name] = mod
		imp_set[name] = true
	}
}

var _ = __Preload()
func __Preload() interface{} {
	__PreloadStdLib()
	return nil
}
func __PreloadStdLib() {
	var fs = RealFileSystem {}
	var exe_path, err = os.Executable()
	if err != nil { panic(err) }
	var ctx = MakeEntryContext()
	for _, name := range __StdLibModules {
		var file = filepath.Join (
			filepath.Dir(exe_path), StdlibFolder, name,
		)
		var _, err = loadModule(file, fs, ctx, __StdLibIndex, __StdLibResIndex)
		if err != nil {
			fmt.Fprintf (
				os.Stderr,
				"%v*** Failed to Load Standard Library%v\n*\n%s\n",
				"\033[1m", "\033[0m", err.Error(),
			)
			os.Exit(3)
		}
	}
}

