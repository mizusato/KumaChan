package richtext


type Text struct {
	Blocks  [] Block  `kmd:"blocks"`
}

type Block struct {
	Indent  int64    `kmd:"indent"`
	Lines   [] Line  `kmd:"lines"`
}

type Line struct {
	Spans  [] Span  `kmd:"spans"`
}

type Span struct {
	Text  string     `kmd:"text"`
	Tags  [] string  `kmd:"tags"`
	Link  MaybeLink  `kmd:"link"` // for doc links
}
const (
	TAG_DOC = "doc"
	TAG_SRC_KEYWORD = "keyword"
	TAG_SRC_FUNCTION = "function"
	TAG_SRC_CONST = "const"
	TAG_SRC_COMMENT = "comment"
	TAG_ERR_NORMAL = "error"
	TAG_ERR_NOTE = "note"
	TAG_ERR_EM = "em"
	TAG_ERR_INLINE = "inline"
)

type MaybeLink interface { maybe(Link, MaybeLink) }
func (Link) maybe(Link, MaybeLink) {}
type Link struct {
	Page    string   `kmd:"page"`
	Anchor  string   `kmd:"anchor"`
}


