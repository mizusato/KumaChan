package richtext


type RenderOptionsLinear struct {
	UseAnsiColor  bool
}

type RenderOptionsHtml struct {
	// empty now
}

func (t Text) RenderLinear(opts RenderOptionsLinear) string {
	panic("not implemented")  // TODO
}

func (t Text) RenderHtml(opts RenderOptionsHtml) string {
	panic("not implemented")  // TODO
}


