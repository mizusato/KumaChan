package loader

import (
	"fmt"
	"kumachan/parser"
	."kumachan/error"
)


type ErrorContext struct {
	ImportPoint  MaybeErrorPoint
	LocalAlias   string
}

func (ctx ErrorContext) GenErrMsg() string {
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
	FilePath  string
	Message   string
	Context   ErrorContext
}
func (impl E_ParseFailed) LoaderError() {}
type E_ParseFailed struct {
	PartialAST   *parser.Tree
	ParserError  *parser.Error
	Context      ErrorContext
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

