package richtext


type Text struct {
	Blocks  [] Block  `kmd:"blocks"`
}
func (t *Text) Write(blocks ...Block) {
	t.Blocks = append(t.Blocks, blocks...)
}
func (t *Text) WriteText(another Text) {
	t.Blocks = append(t.Blocks, another.Blocks...)
}

type Block struct {
	Indent  int64    `kmd:"indent"`
	Lines   [] Line  `kmd:"lines"`
}
func (b Block) ToText() Text {
	return Text { Blocks: [] Block { b } }
}
func (b *Block) WriteLine(content string, tags ...string) {
	b.Lines = append(b.Lines, Line { [] Span { {
		Content: content,
		Tags:    tags,
	} } })
}
func (b *Block) WriteSpan(content string, tags ...string) {
	var span = Span {
		Content: content,
		Tags:    tags,
	}
	if len(b.Lines) > 0 {
		var last = &(b.Lines[len(b.Lines) - 1])
		last.Spans = append(last.Spans, span)
	} else {
		b.Lines = append(b.Lines, Line { Spans: [] Span { span } })
	}
}
func (b *Block) WriteLineFeed() {
	b.Lines = append(b.Lines, Line { Spans: [] Span {} })
}

type Line struct {
	Spans  [] Span  `kmd:"spans"`
}

type Span struct {
	Content  string     `kmd:"content"`
	Tags     [] string  `kmd:"tags"`
	Link     MaybeLink  `kmd:"link"` // for doc links
}
const (
	TAG_DOC          = "doc"
	TAG_EM           = "em"
	TAG_SRC_NORMAL   = "source"
	TAG_SRC_KEYWORD  = "keyword"
	TAG_SRC_FUNCTION = "function"
	TAG_SRC_CONST    = "const"
	TAG_SRC_COMMENT  = "comment"
	TAG_HIGHLIGHT    = "highlight"
	TAG_ERR_NORMAL   = "error"
	TAG_ERR_NOTE     = "note"
	TAG_ERR_INLINE   = "inline"
)

type MaybeLink interface { maybe(Link, MaybeLink) }
func (Link) maybe(Link, MaybeLink) {}
type Link struct {
	Page    string   `kmd:"page"`
	Anchor  string   `kmd:"anchor"`
}


