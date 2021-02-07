package loader

import (
	"os"
	"fmt"
	"errors"
	"reflect"
	"strings"
	"path/filepath"
	"encoding/json"
	"kumachan/compiler/loader/parser/syntax"
	"kumachan/compiler/loader/common"
	"kumachan/compiler/loader/extra"
	"kumachan/lang"
)


var __UnitFileLoaders = [] common.UnitFileLoader {
	extra.WebAssetLoader(),
	extra.PNG_Loader(),
}

type ModuleThunk struct {
	FilePath    string
	FileInfo    os.FileInfo
	Manifest    Manifest
	Content     Structure
	Standalone  bool
	Resources   ResIndex
}

func readModulePath(path string, fs FileSystem, root bool) (ModuleThunk, error) {
	var res = make(ResIndex)
	path = filepath.Clean(path)
	fd, err := fs.Open(path)
	if err != nil { return ModuleThunk{}, err }
	defer func() {
		_ = fd.Close()
	} ()
	fd_info, err := fd.Info()
	if err != nil { return ModuleThunk{}, err }
	if fd_info.IsDir() {
		items, err := fd.ReadDir()
		if err != nil { return ModuleThunk{}, err }
		var has_manifest = false
		var manifest_content ([] byte)
		var manifest_path string
		var unit_files = make([] common.UnitFile, 0)
		for _, item := range items {
			var item_name = item.Name()
			var item_path = filepath.Join(path, item_name)
			if item_name == ManifestFileName {
				var item_content, err = ReadFile(item_path, fs)
				if err != nil { return ModuleThunk{}, err }
				manifest_content = item_content
				manifest_path = item_path
				has_manifest = true
			}
		}
		var manifest Manifest
		if has_manifest {
			var err = json.Unmarshal(manifest_content, &manifest)
			if err != nil {
				return ModuleThunk{}, errors.New(fmt.Sprintf(
					"error decoding module manifest %s: %s",
					manifest_path, err.Error()))
			}
		} else if !(root) {
			manifest = DefaultManifest(path)
		} else {
			return ModuleThunk{}, errors.New(fmt.Sprintf(
				"missing module manifest in %s", path))
		}
		for _, item := range items {
			var item_name = item.Name()
			var item_path = filepath.Join(path, item_name)
			if strings.HasSuffix(item_name, SourceSuffix) {
				var item_content, err = ReadFile(item_path, fs)
				if err != nil { return ModuleThunk{}, err }
				unit_files = append(unit_files, SourceFile {
					Path:    item_path,
					Content: item_content,
				})
			} else {
				var loader common.UnitFileLoader
				var loader_exists = false
				outer: for _, l := range __UnitFileLoaders {
					for _, ext := range l.Extensions {
						if strings.HasSuffix(item_name, ("." + ext)) {
							loader = l
							loader_exists = true
							break outer
						}
					}
				}
				if loader_exists {
					var item_config interface{} = nil
					var config_rv = reflect.ValueOf(manifest.Config)
					var config_t = config_rv.Type()
					for i := 0; i < config_t.NumField(); i += 1 {
						var field_t = config_t.Field(i)
						if field_t.Tag.Get("json") == loader.Name {
							item_config = config_rv.Field(i).Interface()
						}
					}
					var content, err = ReadFile(item_path, fs)
					if err != nil { return ModuleThunk{}, err }
					if loader.IsResource {
						res[item_path] = lang.Resource {
							Kind: loader.Name,
							MIME: loader.GetMIME(item_path),
							Data: content,
						}
					}
					f, err := loader.Load(item_path, content, item_config)
					if err != nil { return ModuleThunk{}, err }
					unit_files = append(unit_files, f)
				}
			}
		}
		var mod_name = manifest.Name
		if mod_name == "" {
			return ModuleThunk{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "name" cannot be empty or omitted`))
		}
		if !(syntax.GetIdentifierFullRegexp().MatchString(mod_name)) {
			return ModuleThunk{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "name" has an invalid value`))
		}
		if strings.ContainsAny(manifest.Project, ": \r\n") {
			return ModuleThunk{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "project" has an invalid value`))
		}
		if strings.ContainsAny(manifest.Version, ": \r\n") {
			return ModuleThunk{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "version" has an invalid value`))
		}
		if strings.ContainsAny(manifest.Vendor, ": \r\n") {
			return ModuleThunk{}, errors.New (
				fmt.Sprintf("invalid module manifest %s: %s",
					manifest_path, `field "vendor" has an invalid value`))
		}
		return ModuleThunk {
			FilePath:  path,
			FileInfo:  fd_info,
			Manifest:  manifest,
			Content:   ModuleFolder { unit_files },
			Resources: res,
		}, nil
	} else {
		var content, err = fd.ReadContent()
		if err != nil { return ModuleThunk{}, err }
		return ModuleThunk {
			FilePath: path,
			FileInfo: fd_info,
			Manifest: Manifest {
				Name: StandaloneScriptModuleName,
			},
			Content:  StandaloneScript {
				File: SourceFile {
					Path:    path,
					Content: content,
				},
			},
			Resources:  res,
			Standalone: true,
		}, nil
	}
}
