package source

import "kumachan/standalone/util/richtext"


type Location struct {
	File  File
	Pos   Position
}
func (l Location) FilePath() string {
	return l.File.GetPath()
}
func (l Location) PosDesc() string {
	return l.File.DescribePosition(l.Pos)
}
func (l Location) FormatMessage(b richtext.Block) richtext.Text {
	return l.File.FormatMessage(l.Pos, b)
}

type Section struct {
	Name   string
	Start  Location
}

type File interface {
	GetPath() string
	DescribePosition(Position) string
	FormatMessage(Position, richtext.Block) richtext.Text
}

type Position struct {
	StartTokenIndex  uint32
	EndTokenIndex    uint32
}


