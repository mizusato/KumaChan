package lib


func ByteFrom(i interface{}) uint8 {
	switch x := i.(type) {
	case uint8:
		return x
	case int8:
		return uint8(x)
	default:
		panic("invalid Byte")
	}
}

func WordFrom(i interface{}) uint16 {
	switch x := i.(type) {
	case uint16:
		return x
	case int16:
		return uint16(x)
	default:
		panic("invalid Word")
	}
}

func DwordFrom(i interface{}) uint32 {
	switch x := i.(type) {
	case uint32:
		return x
	case int32:
		return uint32(x)
	default:
		panic("invalid Dword")
	}
}

func QwordFrom(i interface{}) uint64 {
	switch x := i.(type) {
	case uint64:
		return x
	case int64:
		return uint64(x)
	default:
		panic("invalid Qword")
	}
}


var BitwiseFunctions = map[string] interface{} {
	"&bit": func(a bool, b bool) bool {
		return a && b
	},
	"|bit": func(a bool, b bool) bool {
		return a || b
	},
	"^bit": func(a bool) bool {
		return !a
	},
	"&byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) & ByteFrom(b)
	},
	"|byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) | ByteFrom(b)
	},
	"^byte": func(a interface{}) uint8 {
		return ^(ByteFrom(a))
	},
	"<<byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) << ByteFrom(b)
	},
	">>byte": func(a interface{}, b interface{}) uint8 {
		return ByteFrom(a) >> ByteFrom(b)
	},
	"&word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) & WordFrom(b)
	},
	"|word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) | WordFrom(b)
	},
	"^word": func(a interface{}) uint16 {
		return ^(WordFrom(a))
	},
	"<<word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) << ByteFrom(b)
	},
	">>word": func(a interface{}, b interface{}) uint16 {
		return WordFrom(a) >> ByteFrom(b)
	},
	"&dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) & DwordFrom(b)
	},
	"|dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) | DwordFrom(b)
	},
	"^dword": func(a interface{}) uint32 {
		return ^(DwordFrom(a))
	},
	"<<dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) << ByteFrom(b)
	},
	">>dword": func(a interface{}, b interface{}) uint32 {
		return DwordFrom(a) >> ByteFrom(b)
	},
	"&qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) & QwordFrom(b)
	},
	"|qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) | QwordFrom(b)
	},
	"^qword": func(a interface{}) uint64 {
		return ^(QwordFrom(a))
	},
	"<<qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) << ByteFrom(b)
	},
	">>qword": func(a interface{}, b interface{}) uint64 {
		return QwordFrom(a) >> ByteFrom(b)
	},
}
