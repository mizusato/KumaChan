package loader

import (
	"os"
	"fmt"
	"kumachan/parser"
	."kumachan/error"
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

func (ctx Context) GenErrMsg() string {
	switch p := ctx.ImportPoint.(type) {
	case ErrorPoint:
		return p.GenErrMsg(fmt.Sprintf (
			"%vUnable to import module %v%s%v",
			Red, Bold, ctx.LocalAlias, Reset,
		))
	default:
		return fmt.Sprintf (
			"%v*** Unable to load given source file%v",
			Red+Bold, Reset,
		)
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
	PartialAST  *parser.Tree
	ParserError *parser.Error
}
func (e E_NameConflict) LoaderError() {}
type E_NameConflict struct {
	ModuleName string
	FilePath1  string
	FilePath2  string
}
func (e E_CircularImport) LoaderError() {}
type E_CircularImport struct {
	ModuleName string
}
func (e E_ConflictAlias) LoaderError() {}
type E_ConflictAlias struct {
	LocalAlias  string
}

func (err *Error) Error() string {
	var import_error = err.Context.GenErrMsg()
	var detail string
	switch e := err.Concrete.(type) {
	case E_ReadFileFailed:
		detail = fmt.Sprintf (
			"%vCannot open source file: %v%s%v",
			Red, Bold, e.Message, Reset,
		)
	case E_ParseFailed:
		detail = e.ParserError.DetailedMessage(e.PartialAST)
	case E_NameConflict:
		detail = fmt.Sprintf (
			"%vThe module name %v%s%v is used by both source files %v%s%v and %v%s%v",
			Red,
			Bold, e.ModuleName, Reset+Red,
			Bold, e.FilePath2, Reset+Red,
			Bold, e.FilePath1, Reset,
		)
	case E_CircularImport:
		detail = fmt.Sprintf (
			"%vCircular import of module %v%s%v",
			Red, Bold, e.ModuleName, Reset,
		)
	case E_ConflictAlias:
		detail = fmt.Sprintf (
			"%vA module named %v%s%v already imported in current module%v",
			Red, Bold, e.LocalAlias, Reset+Red, Reset,
		)
	default:
		panic("unknown concrete error type")
	}
	return GenCompilationFailedMessage(err.Concrete, []string {
		import_error, detail,
	})
}
