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

type Error interface { error; GetContext() Context }

func (e E_ReadFileFailed) GetContext() Context { return e.Context }
type E_ReadFileFailed struct {
	FilePath string
	Message  string
	Context  Context
}
func (e E_ParseFailed) GetContext() Context { return e.Context }
type E_ParseFailed struct {
	PartialAST  *parser.Tree
	ParserError *parser.Error
	Context     Context
}
func (e E_NameConflict) GetContext() Context { return e.Context }
type E_NameConflict struct {
	ModuleName string
	FilePath1  string
	FilePath2  string
	Context    Context
}
func (e E_CircularImport) GetContext() Context { return e.Context }
type E_CircularImport struct {
	ModuleName string
	Context    Context
}
func (e E_ConflictAlias) GetContext() Context { return e.Context }
type E_ConflictAlias struct {
	LocalAlias  string
	Context     Context
}

func InflateErrorMessage(e Error, detail string) string {
	var ctx = e.GetContext()
	var import_error = ctx.GenErrMsg()
	var err_type = GetErrorTypeName((interface{})(e))
	return fmt.Sprintf (
		"\n%v*** Failed to Compile (%s)%v\n*\n%s\n*\n%s\n",
		Bold, err_type, Reset, import_error, detail,
	)
}

func (e E_ReadFileFailed) Error() string {
	return InflateErrorMessage(e, fmt.Sprintf (
		"%vCannot open source file: %v%s%v",
		Red, Bold, e.Message, Reset,
	))
}

func (e E_ParseFailed) Error() string {
	return InflateErrorMessage (
		e, e.ParserError.DetailedMessage(e.PartialAST),
	)
}

func (e E_NameConflict) Error() string {
	return InflateErrorMessage(e, fmt.Sprintf (
		"%vThe module name %v%s%v is used by both source files %v%s%v and %v%s%v",
		Red,
		Bold, e.ModuleName, Reset+Red,
		Bold, e.FilePath2, Reset+Red,
		Bold, e.FilePath1, Reset,
	))
}

func (e E_CircularImport) Error() string {
	return InflateErrorMessage(e, fmt.Sprintf (
		"%vCircular import of module %v%s%v",
		Red, Bold, e.ModuleName, Reset,
	))
}

func (e E_ConflictAlias) Error() string {
	return InflateErrorMessage(e, fmt.Sprintf (
		"%vA module named %v%s%v already imported in current module%v",
		Red, Bold, e.LocalAlias, Reset+Red, Reset,
	))
}
