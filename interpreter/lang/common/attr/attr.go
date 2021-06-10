package attr


type RawAttr struct {
	RawDoc   [] string
	RawTags  map[string] string
}

type Attr struct {
	Doc  string
}
type TypeAttr struct {
	Attr
	TypeTags
}
type FieldAttr struct {
	Attr
	FieldTags
}
type FuncAttr struct {
	Attr
	FuncTags
}

type Tags struct {
	Custom  map[string] string
}
type TypeTags struct {
	Tags
	// TODO
}
type FieldTags struct {
	Tags
	// TODO
}
type FuncTags struct {
	Tags
	// TODO
}


