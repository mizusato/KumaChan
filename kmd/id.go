package kmd


type TypeId struct {
	TypeIdFuzzy
	Version  string
}
type TypeIdFuzzy struct {
	Vendor   string
	Name     string
}
func TheTypeId(vendor string, name string, version string) TypeId {
	return TypeId {
		TypeIdFuzzy: TypeIdFuzzy {
			Vendor: vendor,
			Name:   name,
		},
		Version:    version,
	}
}

type AdapterId struct {
	From  TypeId
	To    TypeId
}

type ValidatorId TypeId

type TransformerPartId interface { TransformerPartId() }
func (SerializerId) TransformerPartId() {}
type SerializerId struct {
	TypeId
}
func (DeserializerId) TransformerPartId() {}
type DeserializerId struct {
	TypeId
}
