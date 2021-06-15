package attr

import "kumachan/interpreter/lang/common/source"


type Attrs struct {
	Location  source.Location
	Section   *source.Section
	Doc       string
}
type TypeAttrs struct {
	Attrs
	Metadata  TypeMetadata
}
type FieldAttr struct {
	Attrs
	// TODO
}
type FuncAttr struct {
	Attrs
	// TODO
}

type TypeMetadata struct {
	Data     TypeDataConfig     `json:"data"`
	Service  TypeServiceConfig  `json:"service"`
}
type TypeDataConfig struct {
	Name     string  `json:"name"`
	Version  string  `json:"version"`
}
type TypeServiceConfig struct {
	Name     string  `json:"name"`
	Version  string  `json:"version"`
}


