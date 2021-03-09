package loader

import (
	"os"
	"fmt"
	"path/filepath"
	"kumachan/stdlib"
)


var __StdLibModules = stdlib.GetModuleDirectoryNames()
var __StdLibIndex = make(Index)
var __StdLibResIndex = make(ResIndex)
func ImportStdLib(imp_map (map[string] *Module), imp_set (map[string] bool)) {
	for name, mod := range __StdLibIndex {
		imp_map[name] = mod
		imp_set[name] = true
	}
}
func IsStdLibModule(name string) bool {
	var _, is = __StdLibIndex[name]
	return is
}

var _ = __Preload()
func __Preload() interface{} {
	// __PreloadStdLib()
	return nil
}
func __PreloadStdLib() {
	var fs = RealFileSystem {}
	var stdlib_dir = stdlib.GetDirectoryPath()
	var ctx = MakeEntryContext()
	for _, name := range __StdLibModules {
		var file = filepath.Join (stdlib_dir, name)
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

