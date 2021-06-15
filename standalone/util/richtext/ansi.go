package richtext


type AnsiPalette (func(string) string)

const bold = "\033[1m"
const red = "\033[31m"
const green = "\033[32m"
const yellow = "\033[33m"
const blue = "\033[34m"
const magenta = "\033[35m"
const cyan = "\033[36m"
const reset = "\033[0m"

func LightPalette() AnsiPalette { return lightPalette }
var lightPalette AnsiPalette = func(tag string) string {
	switch tag {
	case TAG_EM:         return bold
	case TAG_HIGHLIGHT:  return (bold + red)
	case TAG_ERR_NORMAL: return bold
	case TAG_ERR_INLINE: return (bold + red)
	case TAG_ERR_NOTE:   return (bold + blue)
	default: return ""
	}
}

func DarkPalette() AnsiPalette { return darkPalette }
var darkPalette AnsiPalette = func(tag string) string {
	switch tag {
	case TAG_EM:         return bold
	case TAG_HIGHLIGHT:  return (bold + yellow)
	case TAG_ERR_NORMAL: return bold
	case TAG_ERR_INLINE: return (bold + yellow)
	case TAG_ERR_NOTE:   return (bold + cyan)
	default: return ""
	}
}


