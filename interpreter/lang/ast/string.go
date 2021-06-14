package ast


func (impl StringLiteral) Term() {}
type StringLiteral struct {
	Node                          `part:"string"`
	First  StringText             `part:"string_text"`
	Parts  [] VariousStringPart   `list_rec:"string_parts"`
}
type VariousStringPart struct {
	Node               `part:"string_part"`
	Part  StringPart   `use:"first"`
}
type StringPart interface { StringPart() }
func (impl StringText) StringPart() {}
type StringText struct {
	Node             `part:"string_text"`
	Value  [] rune   `content:"SqStr"`
}

func (impl Formatter) Term() {}
type Formatter struct {
	Node                             `part:"formatter"`
	First  FormatterText             `part:"formatter_text"`
	Parts  [] VariousFormatterPart   `list_rec:"formatter_parts"`
}
type VariousFormatterPart struct {
	Node                   `part:"formatter_part"`
	Part   FormatterPart   `use:"first"`
}
type FormatterPart interface { FormatterPart() }
func (impl FormatterText) FormatterPart() {}
type FormatterText struct {
	Node                `part:"formatter_text"`
	Template  [] rune   `content:"DqStr"`
}

func (impl CharLiteral) Term() {}
func (impl CharLiteral) StringPart() {}
func (impl CharLiteral) FormatterPart() {}
type CharLiteral struct {
	Node             `part:"char"`
	Value  [] rune   `content:"Char"`
}

