package loader

import (
	. "kumachan/error"
	"kumachan/parser"
	"kumachan/parser/ast"
	"os"
)


type Context struct {
	ImportPoint  MaybeErrorPoint
	LocalAlias   string
	BreadCrumbs  [] Ancestor
}

type Ancestor struct {
	ModuleName  string
	FileInfo    os.FileInfo
	FilePath    string
}

func MakeEntryContext() Context {
	return Context {
		ImportPoint: nil,
		BreadCrumbs: make([]Ancestor, 0),
	}
}

func (ctx Context) GetErrorDescription() ErrorMessage {
	var _, ok = ctx.ImportPoint.(ErrorPoint)
	if ok {
		var desc = make(ErrorMessage, 0)
		desc.WriteText(TS_ERROR, "Unable to import module")
		desc.Write(T_SPACE)
		desc.WriteText(TS_INLINE_CODE, ctx.LocalAlias)
		return desc
	} else {
		var msg = make(ErrorMessage, 0)
		msg.WriteText(TS_ERROR, "*** Unable to load given source file")
		return msg
	}
}

func (ctx Context) GetFullErrorMessage(note ErrorMessage) ErrorMessage {
	var p, ok = ctx.ImportPoint.(ErrorPoint)
	if ok {
		return FormatErrorAt(p, ctx.GetErrorDescription(), note)
	} else {
		return ctx.GetErrorDescription()
	}
}


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
func (e E_ParseFailed) LoaderError() {}
type E_ParseFailed struct {
	PartialAST  *ast.Tree
	ParserError *parser.Error
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

func (err *Error) Note() ErrorMessage {
	var msg = make(ErrorMessage, 0)
	msg.WriteText(TS_NORMAL, "**\n")
	switch e := err.Concrete.(type) {
	case E_ReadFileFailed:
		msg.WriteText(TS_ERROR, "Cannot open source file:")
		msg.WriteEndText(TS_ERROR, e.Message)
	case E_ParseFailed:
		msg.WriteAll(e.ParserError.DetailedMessage(e.PartialAST))
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
	default:
		panic("unknown error kind")
	}
	return msg
}

func (err *Error) Desc() ErrorMessage {
	return err.Context.GetErrorDescription()
}

func (err *Error) Message() ErrorMessage {
	return err.Context.GetFullErrorMessage(err.Note())
}

func (err *Error) Error() string {
	var msg = MsgFailedToCompile(err.Concrete, []ErrorMessage {
		err.Message(),
	})
	return msg.String()
}
