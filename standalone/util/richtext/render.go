package richtext

import "strings"


type RenderOptionsLinear struct {
	AnsiPalette   AnsiPalette
}
func (opts RenderOptionsLinear) WrapSpanContent(content string, tags ([] string)) string {
	if opts.AnsiPalette != nil {
		var buf strings.Builder
		for _, tag := range tags {
			buf.WriteString(opts.AnsiPalette(tag))
		}
		buf.WriteString(content)
		buf.WriteString(reset)
		return buf.String()
	} else {
		return content
	}
}

type RenderOptionsHtml struct {
	// empty now
}

func (t Text) RenderLinear(opts RenderOptionsLinear) string {
	var buf strings.Builder
	for _, b := range t.Blocks {
		for _, l := range b.Lines {
			for i, span := range l.Spans {
				buf.WriteString(opts.WrapSpanContent(span.Content, span.Tags))
				if (i + 1) < len(l.Spans) {
					var next = l.Spans[(i + 1)]
					if !(strings.HasSuffix(span.Content, " ")) &&
						!(strings.HasPrefix(next.Content, " ")) {
						buf.WriteRune(' ')
					}
				}
			}
			buf.WriteRune('\n')
		}
	}
	return buf.String()
}

func (t Text) RenderHtml(_ RenderOptionsHtml) string {
	panic("not implemented")  // TODO
}


