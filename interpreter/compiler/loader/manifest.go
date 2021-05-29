package loader

import (
	"strings"
	"net/url"
	"path/filepath"
	"kumachan/interpreter/compiler/loader/extra"
)



const DefaultVersion = "dev"

type Manifest struct {
	Vendor   string   `json:"vendor"`
	Project  string   `json:"project"`
	Version  string   `json:"version"`
	Name     string   `json:"name"`
	Kind     string   `json:"kind"`
	Config   Config   `json:"config"`
}
type Config struct {
	Service   ServiceConfig          `json:"service"`
	PNG       extra.PNG_Config       `json:"png"`
	WebAsset  extra.WebAssetConfig   `json:"web_asset"`
	UiXml     extra.UiXmlConfig      `json:"ui_xml"`
}
type ServiceConfig struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
}
func DefaultManifest(path string) Manifest {
	// TODO: modules without explicit manifest should inherit
	//       vendor, project and version from the root module
	var abs_path, err = filepath.Abs(path)
	if err == nil { path = abs_path }
	var dir = filepath.Dir(path)
	var base = filepath.Base(path)
	return Manifest {
		// TODO: sanitize string fields
		Vendor:  "",
		Project: url.PathEscape(strings.ReplaceAll(dir, ":", "/")),
		Version: "",
		Name:    base,
		Config:  Config {
			PNG:      extra.PNG_Config { Public: true },
			WebAsset: extra.WebAssetConfig { Public: true },
		},
	}
}

