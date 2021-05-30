package loader

import (
	"os"
	"path/filepath"
	"kumachan/interpreter/parser/ast"
	"kumachan/interpreter/compiler/loader/common"
)


func CraftThunk(manifest Manifest, path string, tree ast.Root) ModuleThunk {
	return ModuleThunk {
		FilePath:   path,
		FileInfo:   craftModuleFileInfo(filepath.Base(path)),
		Manifest:   manifest,
		Content:    PredefinedAST { tree },
		Standalone: false,
	}
}

func CraftEmptyThunk(manifest Manifest, path string) ModuleThunk {
	return CraftThunk(manifest, path, common.CreateEmptyAST(path))
}

func craftModuleFileInfo(name string) os.FileInfo {
	return craftedFileInfo {
		name: name,
	}
}

