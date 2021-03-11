package loader

import (
	. "kumachan/util/error"
	"kumachan/lang/parser"
)


type Error struct {
	Context   Context
	Concrete  ConcreteError
}

type ConcreteError interface { LoaderError() }

func (e E_ReadFileFailed) LoaderError() {}
type E_ReadFileFailed struct {
	FilePath string
	Message  string
}
func (e E_StandaloneImported) LoaderError() {}
type E_StandaloneImported struct {}
func (e E_ParseFailed) LoaderError() {}
type E_ParseFailed struct {
	ParserError  *parser.Error
}
func (e E_NameConflict) LoaderError() {}
type E_NameConflict struct {
	ModuleName  string
	FilePath1   string
	FilePath2   string
}
func (e E_CircularImport) LoaderError() {}
type E_CircularImport struct {
	ModuleName  string
}
func (e E_ConflictAlias) LoaderError() {}
type E_ConflictAlias struct {
	LocalAlias  string
}
func (e E_DuplicateImport) LoaderError() {}
type E_DuplicateImport struct {
	ModuleName  string
}
func (e E_InvalidService) LoaderError() {}
type E_InvalidService struct {
	Reason  string
}

func (err *Error) Desc() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "**\n")
	switch e := err.Concrete.(type) {
	case E_ReadFileFailed:
		msg.WriteText(TS_ERROR, "Cannot open source file:")
		msg.WriteEndText(TS_ERROR, e.Message)
	case E_StandaloneImported:
		msg.WriteText(TS_ERROR, "Standalone scripts are not importable")
	case E_ParseFailed:
		msg.WriteAll(e.ParserError.Message())
	case E_NameConflict:
		msg.WriteText(TS_ERROR, "The module name")
		msg.WriteInnerText(TS_INLINE_CODE, e.ModuleName)
		msg.WriteText(TS_ERROR, "is used by both source files")
		msg.WriteInnerText(TS_INLINE, e.FilePath2)
		msg.WriteText(TS_ERROR, "and")
		msg.WriteEndText(TS_INLINE, e.FilePath1)
	case E_CircularImport:
		msg.WriteText(TS_ERROR, "Circular import of module")
		msg.WriteEndText(TS_INLINE_CODE, e.ModuleName)
	case E_ConflictAlias:
		msg.WriteText(TS_ERROR, "A module already imported as name")
		msg.WriteInnerText(TS_INLINE_CODE, e.LocalAlias)
		msg.WriteText(TS_ERROR, "in current module")
	case E_DuplicateImport:
		msg.WriteText(TS_ERROR, "Duplicate import of module")
		msg.WriteInnerText(TS_INLINE_CODE, e.ModuleName)
	case E_InvalidService:
		msg.WriteText(TS_ERROR, "Invalid service module:")
		msg.WriteEndText(TS_INFO, e.Reason)
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *Error) Message() ErrorMessage {
	return err.Context.GetFullErrorMessage(err.Desc())
}

func (err *Error) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, []ErrorMessage {
		err.Message(),
	})
	return msg.String()
}

