package source

import "kumachan/standalone/util/richtext"


type Text = richtext.Text

type File interface {
	GetPath() string
	DescribePosition(Position) string
	FormatMessage(Position, Text) Text
}

type Location struct {
	File  File
	Pos   Position
}

type Position struct {
	StartTokenIndex  uint32
	EndTokenIndex    uint32
}


