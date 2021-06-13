package source


type ErrorMessage struct {
	Description  Text
	Location     Location
}
func (msg ErrorMessage) ToFullText() Text {
	// TODO
}
func (msg ErrorMessage) ToSerializable() SerializableErrorMessage {
	return SerializableErrorMessage {
		Description: msg.Description,
		FilePath:    msg.Location.File.GetPath(),
		Position:    msg.Location.File.DescribePosition(msg.Location.Pos),
		TokenIndex:  int64(msg.Location.Pos.StartTokenIndex),
	}
}

type SerializableErrorMessage struct {
	Description  Text    `kmd:"description"`
	FilePath     string  `kmd:"file-path"`
	Position     string  `kmd:"position"`
	TokenIndex   int64   `kmd:"token-index"`
}


