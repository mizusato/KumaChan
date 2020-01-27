package loader

import (
	"fmt"
	"kumachan/parser"
	."kumachan/error"
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
	return Context{
		ImportPoint: nil,
		BreadCrumbs: make([]Ancestor, 0),
	}
}

func (ctx Context) GenErrMsg() string {
	switch p := ctx.ImportPoint.(type) {
	case ErrorPoint:
		return p.GenErrMsg(fmt.Sprintf (
			"%vunable to import module %v%s%v",
			Red, Bold, ctx.LocalAlias, Reset,
		))
	default:
		return fmt.Sprintf (
			"%v*** abort: unable to load given source file%v\n",
			Red+Bold, Reset,
		)
	}
}

type Error interface { error; LoaderError() }

func (impl E_ReadFileFailed) LoaderError() {}
type E_ReadFileFailed struct {
	FilePath string
	Message  string
	Context  Context
}
func (impl E_ParseFailed) LoaderError() {}
type E_ParseFailed struct {
	PartialAST  *parser.Tree
	ParserError *parser.Error
	Context     Context
}
func (impl E_NameConflict) LoaderError() {}
type E_NameConflict struct {
	ModuleName string
	FilePath1  string
	FilePath2  string
	Context    Context
}
func (impl E_CircularImport) LoaderError() {}
type E_CircularImport struct {
	ModuleName string
	Context    Context
}

func (e E_ReadFileFailed) Error() string {
	var import_msg = e.Context.GenErrMsg()
	var file_msg = fmt.Sprintf (
		"%vcannot open source file: %v%s%v",
		Red, Bold, e.Message, Reset,
	)
	return import_msg + file_msg
}

func (e E_ParseFailed) Error() string {
	var import_msg = e.Context.GenErrMsg()
	var parser_msg = e.ParserError.DetailedMessage(e.PartialAST)
	return import_msg + parser_msg
}

func (e E_NameConflict) Error() string {
	var import_msg = e.Context.GenErrMsg()
	var conflict_msg = fmt.Sprintf (
		"%vmodule name %s used by both source files %s and %s%v",
		Red, e.ModuleName, e.FilePath2, e.FilePath1, Reset,
	)
	return import_msg + conflict_msg
}

func (e E_CircularImport) Error() string {
	var import_msg = e.Context.GenErrMsg()
	var circular_msg = fmt.Sprintf (
		"%vcircular import of module %s%v",
		Red, e.ModuleName, Reset,
	)
	return import_msg + circular_msg
}
