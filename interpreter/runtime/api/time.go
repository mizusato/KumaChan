package api

import (
	. "kumachan/interpreter/runtime/def"
	"time"
)


var TimeConstants = map[string] NativeConstant {
	"Time::UTC": func(h InteropContext) Value {
		return time.UTC
	},
	"Time::Local": func(h InteropContext) Value {
		return time.Local
	},
}
