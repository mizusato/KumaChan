package loader

import (
	"os"
	. "kumachan/util/error"
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
		BreadCrumbs: make([] Ancestor, 0),
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

func (ctx Context) GetFullErrorMessage(detail ErrorMessage) ErrorMessage {
	var desc = ctx.GetErrorDescription()
	desc.Write(T_LF)
	desc.WriteAll(detail)
	var p, ok = ctx.ImportPoint.(ErrorPoint)
	if ok {
		return FormatErrorAt(p, desc)
	} else {
		return desc
	}
}

